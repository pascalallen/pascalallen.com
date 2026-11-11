package camera

import (
	"context"
	"errors"
	"io"
	"runtime"
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
}

func (f *fakeReader) Close() error {
	select {
	case <-f.closed:
	default:
		close(f.closed)
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
}

func newFakeSource() *fakeSource {
	return &fakeSource{feeds: make(chan feed, 16), errs: make(chan error, 16)}
}

func (f *fakeSource) open(ctx context.Context) (io.ReadCloser, error) {
	f.opens.Add(1)
	select {
	case err := <-f.errs:
		return nil, err
	default:
	}
	pr, pw := io.Pipe()
	r := &fakeReader{PipeReader: pr, closed: make(chan struct{})}
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

func subscribeWith(t *testing.T, h *Hub, fd func() feed, ctx context.Context) (<-chan []byte, func(), feed) {
	t.Helper()
	type result struct {
		ch  <-chan []byte
		un  func()
		err error
	}
	done := make(chan result, 1)
	go func() {
		ch, un, err := h.Subscribe(ctx)
		done <- result{ch, un, err}
	}()
	f := fd()
	go func() { _, _ = f.w.Write([]byte("first")) }()
	res := <-done
	require.NoError(t, res.err)

	return res.ch, res.un, f
}

func newHub(src *fakeSource, sleeps chan time.Duration) *Hub {
	return NewHub(Config{
		Source: src.open,
		Sleep: func(ctx context.Context, d time.Duration) error {
			if sleeps == nil {
				<-ctx.Done()
				return ctx.Err()
			}
			sleeps <- d
			return ctx.Err()
		},
		Buffer:         2,
		InitialBackoff: time.Second,
		MaxBackoff:     4 * time.Second,
		StartTimeout:   2 * time.Second,
	})
}

func TestHub(t *testing.T) {
	t.Run("frames reach two subscribers", func(t *testing.T) {
		src := newFakeSource()
		h := newHub(src, nil)
		ctx := t.Context()

		ch1, un1, f := subscribeWith(t, h, func() feed { return src.next(t) }, ctx)
		defer un1()
		assert.Equal(t, []byte("first"), recv(t, ch1))

		ch2, un2, err := h.Subscribe(ctx)
		require.NoError(t, err)
		defer un2()

		_, err = f.w.Write([]byte("frame"))
		require.NoError(t, err)

		assert.Equal(t, []byte("frame"), recv(t, ch1))
		assert.Equal(t, []byte("frame"), recv(t, ch2))
		assert.EqualValues(t, 1, src.opens.Load())
	})

	t.Run("a slow subscriber drops frames without blocking the fast one", func(t *testing.T) {
		src := newFakeSource()
		h := newHub(src, nil)
		ctx := t.Context()

		fast, un1, f := subscribeWith(t, h, func() feed { return src.next(t) }, ctx)
		defer un1()
		_ = recv(t, fast)
		slow, un2, err := h.Subscribe(ctx)
		require.NoError(t, err)
		defer un2()

		for i := 0; i < 10; i++ {
			_, err := f.w.Write([]byte{byte(i)})
			require.NoError(t, err)
			assert.Equal(t, []byte{byte(i)}, recv(t, fast))
		}

		assert.Len(t, slow, 2)
	})

	t.Run("the last unsubscribe stops the source", func(t *testing.T) {
		src := newFakeSource()
		h := newHub(src, nil)
		ctx := t.Context()

		_, un1, f := subscribeWith(t, h, func() feed { return src.next(t) }, ctx)
		_, un2, err := h.Subscribe(ctx)
		require.NoError(t, err)

		un1()
		select {
		case <-f.closed:
			t.Fatal("source stopped while a subscriber remained")
		case <-time.After(50 * time.Millisecond):
		}

		un2()
		select {
		case <-f.closed:
		case <-time.After(2 * time.Second):
			t.Fatal("source was not stopped")
		}
	})

	t.Run("unsubscribe is idempotent", func(t *testing.T) {
		src := newFakeSource()
		h := newHub(src, nil)

		_, un, _ := subscribeWith(t, h, func() feed { return src.next(t) }, t.Context())
		un()
		un()
	})

	t.Run("a context cancel removes the subscriber and stops the source", func(t *testing.T) {
		src := newFakeSource()
		h := newHub(src, nil)
		ctx, cancel := context.WithCancel(t.Context())

		_, _, f := subscribeWith(t, h, func() feed { return src.next(t) }, ctx)
		cancel()

		select {
		case <-f.closed:
		case <-time.After(2 * time.Second):
			t.Fatal("source was not stopped after cancel")
		}
		assert.Eventually(t, func() bool {
			h.mu.Lock()
			defer h.mu.Unlock()
			return len(h.subs) == 0
		}, 2*time.Second, 5*time.Millisecond)
	})

	t.Run("a source exit restarts it after the initial backoff", func(t *testing.T) {
		src := newFakeSource()
		sleeps := make(chan time.Duration, 16)
		h := newHub(src, sleeps)

		ch, un, f := subscribeWith(t, h, func() feed { return src.next(t) }, t.Context())
		defer un()
		_ = recv(t, ch)

		require.NoError(t, f.w.Close())
		assert.Equal(t, time.Second, <-sleeps)

		f2 := src.next(t)
		go func() { _, _ = f2.w.Write([]byte("again")) }()
		assert.Equal(t, []byte("again"), recv(t, ch))
	})

	t.Run("repeated failed starts double the backoff up to the cap", func(t *testing.T) {
		src := newFakeSource()
		sleeps := make(chan time.Duration, 16)
		h := newHub(src, sleeps)

		ch, un, f := subscribeWith(t, h, func() feed { return src.next(t) }, t.Context())
		defer un()
		_ = recv(t, ch)

		for i := 0; i < 5; i++ {
			src.errs <- errors.New("boom")
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
		src := newFakeSource()
		src.errs <- errors.New("no such file")
		h := newHub(src, nil)

		_, _, err := h.Subscribe(t.Context())

		assert.ErrorIs(t, err, camera.ErrUnavailable)
		assert.False(t, h.Available())
	})

	t.Run("a source that exits immediately yields unavailable", func(t *testing.T) {
		src := newFakeSource()
		h := newHub(src, nil)
		go func() { _ = src.next(t).w.Close() }()

		_, _, err := h.Subscribe(t.Context())

		assert.ErrorIs(t, err, camera.ErrUnavailable)
		assert.False(t, h.Available())
	})

	t.Run("a missing binary is unavailable without spawning anything", func(t *testing.T) {
		src := newFakeSource()
		h := NewHub(Config{Source: src.open, Check: func() bool { return false }})

		_, _, err := h.Subscribe(t.Context())

		assert.ErrorIs(t, err, camera.ErrUnavailable)
		assert.False(t, h.Available())
		assert.EqualValues(t, 0, src.opens.Load())
	})

	t.Run("a source that never produces a frame times out as unavailable", func(t *testing.T) {
		src := newFakeSource()
		h := NewHub(Config{Source: src.open, StartTimeout: 20 * time.Millisecond})

		_, _, err := h.Subscribe(t.Context())

		assert.ErrorIs(t, err, camera.ErrUnavailable)
	})

	t.Run("a cancelled context before the first frame returns the context error", func(t *testing.T) {
		src := newFakeSource()
		h := newHub(src, nil)
		ctx, cancel := context.WithCancel(t.Context())
		cancel()

		_, _, err := h.Subscribe(ctx)

		assert.ErrorIs(t, err, context.Canceled)
	})

	t.Run("it is available again after a good run following a failure", func(t *testing.T) {
		src := newFakeSource()
		src.errs <- errors.New("boom")
		h := newHub(src, nil)
		_, _, err := h.Subscribe(t.Context())
		require.ErrorIs(t, err, camera.ErrUnavailable)

		_, un, _ := subscribeWith(t, h, func() feed { return src.next(t) }, t.Context())
		defer un()

		assert.True(t, h.Available())
	})

	t.Run("subscribing and leaving repeatedly leaks no goroutines", func(t *testing.T) {
		src := newFakeSource()
		h := newHub(src, nil)
		cycle := func() {
			ctx, cancel := context.WithCancel(t.Context())
			_, un, _ := subscribeWith(t, h, func() feed { return src.next(t) }, ctx)
			un()
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
