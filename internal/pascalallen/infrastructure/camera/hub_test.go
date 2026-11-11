package camera

import (
	"context"
	"errors"
	"io"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/pascalallen/pascalallen.com/internal/pascalallen/domain/camera"
)

type fakeReader struct {
	*io.PipeReader
	closed chan struct{}
	gate   chan struct{}
	stuck  bool
	once   sync.Once
}

func (f *fakeReader) Close() error {
	f.once.Do(func() { close(f.closed) })
	if f.gate != nil {
		<-f.gate
	}
	if f.stuck {
		return nil
	}

	return f.PipeReader.Close()
}

type feed struct {
	w      *io.PipeWriter
	closed chan struct{}
}

type fakeSource struct {
	opens atomic.Int32
	feeds chan feed
	errs  chan error
	gates chan chan struct{}
	stuck bool
}

func newFakeSource() *fakeSource {
	return &fakeSource{
		feeds: make(chan feed, 1024),
		errs:  make(chan error, 16),
		gates: make(chan chan struct{}, 4),
	}
}

func (f *fakeSource) open(ctx context.Context) (io.ReadCloser, error) {
	f.opens.Add(1)
	select {
	case err := <-f.errs:
		return nil, err
	default:
	}
	pr, pw := io.Pipe()
	r := &fakeReader{PipeReader: pr, closed: make(chan struct{}), stuck: f.stuck}
	select {
	case g := <-f.gates:
		r.gate = g
	default:
	}
	f.feeds <- feed{w: pw, closed: r.closed}

	return r, nil
}

func (f *fakeSource) next(t *testing.T) feed {
	t.Helper()
	select {
	case fd := <-f.feeds:
		return fd
	case <-time.After(2 * time.Second):
		t.Fatal("source was not opened")
	}

	return feed{}
}

type fakeTimer struct {
	f       func()
	stopped bool
	fired   bool
}

type fakeTimers struct {
	mu sync.Mutex
	ts []*fakeTimer
}

func (g *fakeTimers) afterFunc(d time.Duration, f func()) func() bool {
	t := &fakeTimer{f: f}
	g.mu.Lock()
	g.ts = append(g.ts, t)
	g.mu.Unlock()

	return func() bool {
		g.mu.Lock()
		defer g.mu.Unlock()
		active := !t.stopped && !t.fired
		t.stopped = true

		return active
	}
}

func (g *fakeTimers) pending() int {
	g.mu.Lock()
	defer g.mu.Unlock()
	n := 0
	for _, t := range g.ts {
		if !t.stopped && !t.fired {
			n++
		}
	}

	return n
}

func (g *fakeTimers) fireFirst() {
	g.mu.Lock()
	t := g.ts[0]
	g.mu.Unlock()
	t.f()
}

func (g *fakeTimers) fire() {
	g.mu.Lock()
	var fns []func()
	for _, t := range g.ts {
		if !t.stopped && !t.fired {
			t.fired = true
			fns = append(fns, t.f)
		}
	}
	g.mu.Unlock()
	for _, f := range fns {
		f()
	}
}

type fakeClock struct {
	mu  sync.Mutex
	now time.Time
}

func (c *fakeClock) Now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.now
}

func (c *fakeClock) advance(d time.Duration) {
	c.mu.Lock()
	c.now = c.now.Add(d)
	c.mu.Unlock()
}

type rig struct {
	src   *fakeSource
	timer *fakeTimers
	clock *fakeClock
	hub   *Hub
}

func newRig(sleeps chan time.Duration, mod func(*Config)) *rig {
	r := &rig{
		src:   newFakeSource(),
		timer: &fakeTimers{},
		clock: &fakeClock{now: time.Unix(1000, 0)},
	}
	cfg := Config{
		Source: r.src.open,
		Sleep: func(ctx context.Context, d time.Duration) error {
			if sleeps == nil {
				<-ctx.Done()
				return ctx.Err()
			}
			sleeps <- d
			return ctx.Err()
		},
		AfterFunc:      r.timer.afterFunc,
		Now:            r.clock.Now,
		Buffer:         2,
		InitialBackoff: time.Second,
		MaxBackoff:     4 * time.Second,
		StartTimeout:   2 * time.Second,
		Grace:          5 * time.Second,
		FailureTTL:     30 * time.Second,
	}
	if mod != nil {
		mod(&cfg)
	}
	r.hub = NewHub(cfg)

	return r
}

func recv(t *testing.T, ch <-chan []byte) []byte {
	t.Helper()
	select {
	case b := <-ch:
		return b
	case <-time.After(2 * time.Second):
		t.Fatal("no chunk received")
	}

	return nil
}

func (r *rig) join(t *testing.T, ctx context.Context) (<-chan []byte, func(), feed) {
	t.Helper()
	type result struct {
		ch  <-chan []byte
		un  func()
		err error
	}
	done := make(chan result, 1)
	go func() {
		ch, un, err := r.hub.Subscribe(ctx)
		done <- result{ch, un, err}
	}()
	f := r.src.next(t)
	go func() { _, _ = f.w.Write([]byte("first")) }()
	res := <-done
	require.NoError(t, res.err)

	return res.ch, res.un, f
}

func (r *rig) subCount() int {
	r.hub.mu.Lock()
	defer r.hub.mu.Unlock()

	return len(r.hub.subs)
}

func TestHub(t *testing.T) {
	t.Run("frames reach two subscribers", func(t *testing.T) {
		r := newRig(nil, nil)
		ctx := t.Context()

		ch1, un1, f := r.join(t, ctx)
		defer un1()
		assert.Equal(t, []byte("first"), recv(t, ch1))

		ch2, un2, err := r.hub.Subscribe(ctx)
		require.NoError(t, err)
		defer un2()

		_, err = f.w.Write([]byte("frame"))
		require.NoError(t, err)

		assert.Equal(t, []byte("frame"), recv(t, ch1))
		assert.Equal(t, []byte("frame"), recv(t, ch2))
		assert.EqualValues(t, 1, r.src.opens.Load())
	})

	t.Run("a slow subscriber drops frames without blocking the fast one", func(t *testing.T) {
		r := newRig(nil, nil)
		ctx := t.Context()

		fast, un1, f := r.join(t, ctx)
		defer un1()
		_ = recv(t, fast)
		slow, un2, err := r.hub.Subscribe(ctx)
		require.NoError(t, err)
		defer un2()

		for i := 0; i < 10; i++ {
			_, err := f.w.Write([]byte{byte(i)})
			require.NoError(t, err)
			assert.Equal(t, []byte{byte(i)}, recv(t, fast))
		}

		assert.Len(t, slow, 2)
	})

	t.Run("the source stops only after the last subscriber leaves and the grace period ends", func(t *testing.T) {
		r := newRig(nil, nil)
		ctx := t.Context()

		_, un1, f := r.join(t, ctx)
		_, un2, err := r.hub.Subscribe(ctx)
		require.NoError(t, err)

		un1()
		assert.Equal(t, 0, r.timer.pending())
		r.timer.fire()
		select {
		case <-f.closed:
			t.Fatal("source stopped while a subscriber remained")
		case <-time.After(50 * time.Millisecond):
		}

		un2()
		assert.Equal(t, 1, r.timer.pending())
		select {
		case <-f.closed:
			t.Fatal("source stopped before the grace period ended")
		case <-time.After(50 * time.Millisecond):
		}

		r.timer.fire()
		select {
		case <-f.closed:
		case <-time.After(2 * time.Second):
			t.Fatal("source was not stopped")
		}
	})

	t.Run("a subscriber joining inside the grace period reuses the running source", func(t *testing.T) {
		r := newRig(nil, nil)
		ctx := t.Context()

		_, un1, f := r.join(t, ctx)
		un1()
		require.Equal(t, 1, r.timer.pending())

		ch, un2, err := r.hub.Subscribe(ctx)
		require.NoError(t, err)
		defer un2()
		assert.Equal(t, 0, r.timer.pending())
		r.timer.fire()

		_, err = f.w.Write([]byte("still"))
		require.NoError(t, err)
		assert.Equal(t, []byte("still"), recv(t, ch))
		assert.EqualValues(t, 1, r.src.opens.Load())
	})

	t.Run("a new source is not opened until the previous one has been closed and reaped", func(t *testing.T) {
		r := newRig(nil, nil)
		gate := make(chan struct{})
		r.src.gates <- gate
		ctx := t.Context()

		_, un1, f := r.join(t, ctx)
		un1()
		r.timer.fire()
		select {
		case <-f.closed:
		case <-time.After(2 * time.Second):
			t.Fatal("old source was not asked to close")
		}

		type result struct {
			ch  <-chan []byte
			un  func()
			err error
		}
		done := make(chan result, 1)
		go func() {
			ch, un, err := r.hub.Subscribe(ctx)
			done <- result{ch, un, err}
		}()

		time.Sleep(100 * time.Millisecond)
		assert.EqualValues(t, 1, r.src.opens.Load())

		close(gate)
		f2 := r.src.next(t)
		go func() { _, _ = f2.w.Write([]byte("new")) }()
		res := <-done
		require.NoError(t, res.err)
		defer res.un()
		assert.Equal(t, []byte("new"), recv(t, res.ch))
		assert.EqualValues(t, 2, r.src.opens.Load())
	})

	t.Run("a grace timer from an earlier departure cannot stop a source after a rejoin and a second departure", func(t *testing.T) {
		r := newRig(nil, nil)
		ctx := t.Context()

		_, un1, f := r.join(t, ctx)
		un1()
		_, un2, err := r.hub.Subscribe(ctx)
		require.NoError(t, err)
		un2()

		r.timer.fireFirst()

		select {
		case <-f.closed:
			t.Fatal("a stale grace timer stopped the source")
		case <-time.After(50 * time.Millisecond):
		}
	})

	t.Run("unsubscribe is idempotent", func(t *testing.T) {
		r := newRig(nil, nil)

		_, un, _ := r.join(t, t.Context())
		un()
		un()
		assert.Equal(t, 1, r.timer.pending())
	})

	t.Run("a context cancel removes the subscriber and the source stops after grace", func(t *testing.T) {
		r := newRig(nil, nil)
		ctx, cancel := context.WithCancel(t.Context())

		_, _, f := r.join(t, ctx)
		cancel()

		assert.Eventually(t, func() bool { return r.subCount() == 0 }, 2*time.Second, 5*time.Millisecond)
		assert.Eventually(t, func() bool { return r.timer.pending() == 1 }, 2*time.Second, 5*time.Millisecond)
		r.timer.fire()
		select {
		case <-f.closed:
		case <-time.After(2 * time.Second):
			t.Fatal("source was not stopped after cancel")
		}
	})

	t.Run("subscribing with a context that is cancelling concurrently never panics or leaks a subscriber", func(t *testing.T) {
		r := newRig(nil, func(c *Config) { c.StartTimeout = 50 * time.Millisecond })
		var wg sync.WaitGroup

		for i := 0; i < 300; i++ {
			ctx, cancel := context.WithCancel(t.Context())
			wg.Add(2)
			go func() {
				defer wg.Done()
				cancel()
			}()
			go func() {
				defer wg.Done()
				_, un, err := r.hub.Subscribe(ctx)
				if err == nil {
					un()
				}
			}()
		}
		wg.Wait()

		assert.Eventually(t, func() bool { return r.subCount() == 0 }, 2*time.Second, 5*time.Millisecond)
	})

	t.Run("a source exit restarts it after the initial backoff", func(t *testing.T) {
		sleeps := make(chan time.Duration, 16)
		r := newRig(sleeps, nil)

		ch, un, f := r.join(t, t.Context())
		defer un()
		_ = recv(t, ch)

		require.NoError(t, f.w.Close())
		assert.Equal(t, time.Second, <-sleeps)

		f2 := r.src.next(t)
		go func() { _, _ = f2.w.Write([]byte("again")) }()
		assert.Equal(t, []byte("again"), recv(t, ch))
	})

	t.Run("repeated failed starts double the backoff up to the cap", func(t *testing.T) {
		sleeps := make(chan time.Duration, 16)
		r := newRig(sleeps, func(c *Config) { c.MaxFailures = 10 })

		ch, un, f := r.join(t, t.Context())
		defer un()
		_ = recv(t, ch)

		for i := 0; i < 5; i++ {
			r.src.errs <- errors.New("boom")
		}
		require.NoError(t, f.w.Close())

		var got []time.Duration
		for i := 0; i < 5; i++ {
			got = append(got, <-sleeps)
		}
		assert.Equal(t, []time.Duration{
			time.Second, time.Second, 2 * time.Second, 4 * time.Second, 4 * time.Second,
		}, got)
	})

	t.Run("a source that fails to start yields unavailable", func(t *testing.T) {
		r := newRig(nil, nil)
		r.src.errs <- errors.New("no such file")
		ctx, cancel := context.WithTimeout(t.Context(), 500*time.Millisecond)
		defer cancel()

		_, _, err := r.hub.Subscribe(ctx)

		assert.ErrorIs(t, err, camera.ErrUnavailable)
		assert.False(t, r.hub.Available())
	})

	t.Run("a source that exits immediately yields unavailable", func(t *testing.T) {
		r := newRig(nil, nil)
		go func() { _ = r.src.next(t).w.Close() }()
		ctx, cancel := context.WithTimeout(t.Context(), 500*time.Millisecond)
		defer cancel()

		_, _, err := r.hub.Subscribe(ctx)

		assert.ErrorIs(t, err, camera.ErrUnavailable)
		assert.False(t, r.hub.Available())
	})

	t.Run("a missing binary is unavailable without spawning anything", func(t *testing.T) {
		src := newFakeSource()
		h := NewHub(Config{Source: src.open, Check: func() bool { return false }})

		_, _, err := h.Subscribe(t.Context())

		assert.ErrorIs(t, err, camera.ErrUnavailable)
		assert.False(t, h.Available())
		assert.EqualValues(t, 0, src.opens.Load())
	})

	t.Run("a source that never produces a frame times out and reports unavailable", func(t *testing.T) {
		r := newRig(nil, func(c *Config) { c.StartTimeout = 20 * time.Millisecond })

		_, _, err := r.hub.Subscribe(t.Context())

		assert.ErrorIs(t, err, camera.ErrUnavailable)
		assert.False(t, r.hub.Available())
	})

	t.Run("a failure stops counting against availability once the failure window passes", func(t *testing.T) {
		r := newRig(nil, nil)
		r.src.errs <- errors.New("boom")
		_, _, err := r.hub.Subscribe(t.Context())
		require.ErrorIs(t, err, camera.ErrUnavailable)
		require.False(t, r.hub.Available())

		r.clock.advance(29 * time.Second)
		assert.False(t, r.hub.Available())

		r.clock.advance(2 * time.Second)
		assert.True(t, r.hub.Available())
	})

	t.Run("a subscribe attempt after the failure window starts a fresh source", func(t *testing.T) {
		r := newRig(nil, nil)
		r.src.errs <- errors.New("boom")
		_, _, err := r.hub.Subscribe(t.Context())
		require.ErrorIs(t, err, camera.ErrUnavailable)
		r.timer.fire()
		r.clock.advance(time.Minute)

		_, un, _ := r.join(t, t.Context())
		defer un()

		assert.True(t, r.hub.Available())
	})

	t.Run("a cancelled context before the first frame returns the context error", func(t *testing.T) {
		r := newRig(nil, nil)
		ctx, cancel := context.WithCancel(t.Context())
		cancel()

		_, _, err := r.hub.Subscribe(ctx)

		assert.ErrorIs(t, err, context.Canceled)
	})

	t.Run("it is available again after a good run following a failure", func(t *testing.T) {
		r := newRig(nil, nil)
		r.src.errs <- errors.New("boom")
		_, _, err := r.hub.Subscribe(t.Context())
		require.ErrorIs(t, err, camera.ErrUnavailable)
		r.timer.fire()

		_, un, _ := r.join(t, t.Context())
		defer un()

		assert.True(t, r.hub.Available())
	})

	t.Run("a new subscriber gets unavailable while restarts keep failing and existing ones are closed", func(t *testing.T) {
		sleeps := make(chan time.Duration, 16)
		r := newRig(sleeps, nil)
		ch, un, f := r.join(t, t.Context())
		defer un()
		_ = recv(t, ch)
		for i := 0; i < 5; i++ {
			r.src.errs <- errors.New("boom")
		}
		require.NoError(t, f.w.Close())

		assert.Equal(t, time.Second, <-sleeps)
		assert.Equal(t, time.Second, <-sleeps)
		assert.Equal(t, 2*time.Second, <-sleeps)

		_, _, err := r.hub.Subscribe(t.Context())
		assert.ErrorIs(t, err, camera.ErrUnavailable)

		select {
		case _, ok := <-ch:
			assert.False(t, ok)
		case <-time.After(2 * time.Second):
			t.Fatal("the existing subscriber was not closed")
		}
		assert.False(t, r.hub.Available())
		assert.Equal(t, 0, r.subCount())
	})

	t.Run("a stale reader that outlives its source cannot deliver bytes to the next run", func(t *testing.T) {
		r := newRig(nil, nil)
		r.src.stuck = true
		ctx := t.Context()

		_, un1, old := r.join(t, ctx)
		un1()
		r.timer.fire()
		select {
		case <-old.closed:
		case <-time.After(2 * time.Second):
			t.Fatal("old source was not asked to close")
		}
		r.src.stuck = false

		type result struct {
			ch  <-chan []byte
			un  func()
			err error
		}
		done := make(chan result, 1)
		go func() {
			ch, un, err := r.hub.Subscribe(ctx)
			done <- result{ch, un, err}
		}()
		fresh := r.src.next(t)

		stale := make(chan struct{})
		go func() {
			defer close(stale)
			_, _ = old.w.Write([]byte("stale"))
		}()
		<-stale
		time.Sleep(50 * time.Millisecond)
		go func() { _, _ = fresh.w.Write([]byte("new")) }()

		res := <-done
		require.NoError(t, res.err)
		defer res.un()
		assert.Equal(t, []byte("new"), recv(t, res.ch))
		_ = old.w.Close()
	})

	t.Run("a stale reader cannot mark the next run available or clear its failure", func(t *testing.T) {
		r := newRig(nil, nil)
		r.src.stuck = true
		ctx, cancel := context.WithCancel(t.Context())

		done := make(chan struct{})
		go func() {
			defer close(done)
			_, _, _ = r.hub.Subscribe(ctx)
		}()
		old := r.src.next(t)
		cancel()
		<-done
		r.timer.fire()
		<-old.closed

		r.src.stuck = false
		r.src.errs <- errors.New("boom")
		_, _, err := r.hub.Subscribe(t.Context())
		require.ErrorIs(t, err, camera.ErrUnavailable)
		require.False(t, r.hub.Available())

		go func() { _, _ = old.w.Write([]byte("stale")) }()

		assert.Never(t, func() bool { return r.hub.Available() }, 100*time.Millisecond, 5*time.Millisecond)
		_ = old.w.Close()
	})

	t.Run("subscribing and leaving repeatedly leaks no goroutines", func(t *testing.T) {
		r := newRig(nil, nil)
		cycle := func() {
			ctx, cancel := context.WithCancel(t.Context())
			_, un, _ := r.join(t, ctx)
			un()
			r.timer.fire()
			cancel()
		}
		cycle()
		time.Sleep(100 * time.Millisecond)
		runtime.GC()
		before := runtime.NumGoroutine()

		for i := 0; i < 5; i++ {
			cycle()
		}

		deadline := time.Now().Add(3 * time.Second)
		for runtime.NumGoroutine() > before && time.Now().Before(deadline) {
			time.Sleep(10 * time.Millisecond)
		}

		assert.LessOrEqual(t, runtime.NumGoroutine(), before)
	})
}

func TestGStreamerSource(t *testing.T) {
	t.Run("a missing binary returns an error", func(t *testing.T) {
		t.Setenv("PATH", t.TempDir())

		rc, err := GStreamerSource(t.Context())

		assert.Error(t, err)
		assert.Nil(t, rc)
		assert.False(t, GStreamerAvailable())
	})
}
