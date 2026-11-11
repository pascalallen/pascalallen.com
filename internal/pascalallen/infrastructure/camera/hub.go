package camera

import (
	"context"
	"io"
	"log"
	"sync"
	"time"

	"github.com/pascalallen/pascalallen.com/internal/pascalallen/domain/camera"
)

const (
	defaultBuffer         = 30
	defaultInitialBackoff = time.Second
	defaultMaxBackoff     = 30 * time.Second
	defaultStartTimeout   = 10 * time.Second
	defaultGrace          = 5 * time.Second
	defaultMaxFailures    = 3
	readSize              = 65536
)

type Source func(ctx context.Context) (io.ReadCloser, error)

type Sleeper func(ctx context.Context, d time.Duration) error

type Config struct {
	Source         Source
	Check          func() bool
	Sleep          Sleeper
	AfterFunc      func(d time.Duration, f func()) (stop func() bool)
	Now            func() time.Time
	Buffer         int
	InitialBackoff time.Duration
	MaxBackoff     time.Duration
	StartTimeout   time.Duration
	Grace          time.Duration
	FailureTTL     time.Duration
	MaxFailures    int
}

type run struct {
	ctx      context.Context
	cancel   context.CancelFunc
	ready    chan struct{}
	settled  bool
	healthy  bool
	reaped   chan struct{}
	reapOnce sync.Once
}

func (r *run) markReaped() {
	r.reapOnce.Do(func() { close(r.reaped) })
}

type subscriber struct {
	ch     chan []byte
	closed bool
}

type source struct {
	rc   io.ReadCloser
	once sync.Once
}

func (s *source) close() {
	s.once.Do(func() { _ = s.rc.Close() })
}

type Hub struct {
	cfg       Config
	mu        sync.Mutex
	subs      map[*subscriber]struct{}
	cur       *run
	last      *run
	graceSeq  uint64
	graceStop func() bool
	failedAt  time.Time
	hasFailed bool
}

var _ camera.Stream = (*Hub)(nil)

func NewHub(cfg Config) *Hub {
	if cfg.Check == nil {
		cfg.Check = func() bool { return true }
	}
	if cfg.Sleep == nil {
		cfg.Sleep = sleep
	}
	if cfg.AfterFunc == nil {
		cfg.AfterFunc = func(d time.Duration, f func()) func() bool {
			return time.AfterFunc(d, f).Stop
		}
	}
	if cfg.Now == nil {
		cfg.Now = time.Now
	}
	if cfg.Buffer <= 0 {
		cfg.Buffer = defaultBuffer
	}
	if cfg.InitialBackoff <= 0 {
		cfg.InitialBackoff = defaultInitialBackoff
	}
	if cfg.MaxBackoff <= 0 {
		cfg.MaxBackoff = defaultMaxBackoff
	}
	if cfg.StartTimeout <= 0 {
		cfg.StartTimeout = defaultStartTimeout
	}
	if cfg.Grace <= 0 {
		cfg.Grace = defaultGrace
	}
	if cfg.FailureTTL <= 0 {
		cfg.FailureTTL = cfg.MaxBackoff
	}
	if cfg.MaxFailures <= 0 {
		cfg.MaxFailures = defaultMaxFailures
	}

	return &Hub{cfg: cfg, subs: make(map[*subscriber]struct{})}
}

func sleep(ctx context.Context, d time.Duration) error {
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}

func (h *Hub) Available() bool {
	if !h.cfg.Check() {
		return false
	}
	h.mu.Lock()
	defer h.mu.Unlock()

	return !h.hasFailed || h.cfg.Now().Sub(h.failedAt) >= h.cfg.FailureTTL
}

func (h *Hub) Subscribe(ctx context.Context) (<-chan []byte, func(), error) {
	if !h.cfg.Check() {
		return nil, nil, camera.ErrUnavailable
	}
	if err := ctx.Err(); err != nil {
		return nil, nil, err
	}

	s := &subscriber{ch: make(chan []byte, h.cfg.Buffer)}

	h.mu.Lock()
	h.subs[s] = struct{}{}
	if h.graceStop != nil {
		h.graceStop()
		h.graceStop = nil
		h.graceSeq++
	}
	if h.cur == nil {
		h.startLocked()
	}
	r := h.cur
	settled := r.settled
	h.mu.Unlock()

	var once sync.Once
	leave := func() {
		once.Do(func() { h.remove(s) })
	}
	stop := context.AfterFunc(ctx, leave)
	release := func() {
		stop()
		leave()
	}

	if !settled {
		timer := time.NewTimer(h.cfg.StartTimeout)
		defer timer.Stop()
		select {
		case <-r.ready:
		case <-ctx.Done():
			release()
			return nil, nil, ctx.Err()
		case <-timer.C:
			h.mu.Lock()
			h.recordFailureLocked()
			h.mu.Unlock()
			release()
			return nil, nil, camera.ErrUnavailable
		}
	}

	h.mu.Lock()
	ok := r.healthy && !s.closed
	h.mu.Unlock()
	if !ok {
		release()
		return nil, nil, camera.ErrUnavailable
	}

	return s.ch, release, nil
}

func (h *Hub) startLocked() {
	ctx, cancel := context.WithCancel(context.Background())
	r := &run{
		ctx:    ctx,
		cancel: cancel,
		ready:  make(chan struct{}),
		reaped: make(chan struct{}),
	}
	prev := h.last
	h.last = r
	h.cur = r
	go h.loop(r, prev)
}

func (h *Hub) recordFailureLocked() {
	h.hasFailed = true
	h.failedAt = h.cfg.Now()
}

func (h *Hub) remove(s *subscriber) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if _, ok := h.subs[s]; !ok {
		return
	}
	delete(h.subs, s)
	if len(h.subs) > 0 || h.cur == nil {
		return
	}
	h.graceSeq++
	seq := h.graceSeq
	r := h.cur
	h.graceStop = h.cfg.AfterFunc(h.cfg.Grace, func() { h.expire(r, seq) })
}

func (h *Hub) expire(r *run, seq uint64) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.cur != r || h.graceSeq != seq || len(h.subs) > 0 {
		return
	}
	h.graceStop = nil
	h.cur = nil
	r.cancel()
}

func (h *Hub) abort(r *run) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.cur != r {
		return
	}
	for s := range h.subs {
		delete(h.subs, s)
		s.closed = true
		close(s.ch)
	}
	if h.graceStop != nil {
		h.graceStop()
		h.graceStop = nil
		h.graceSeq++
	}
	h.recordFailureLocked()
	h.cur = nil
	r.healthy = false
	if !r.settled {
		r.settled = true
		close(r.ready)
	}
	r.cancel()
}

func (h *Hub) good(r *run) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.cur != r || r.ctx.Err() != nil {
		return
	}
	r.healthy = true
	h.hasFailed = false
	if !r.settled {
		r.settled = true
		close(r.ready)
	}
}

func (h *Hub) fail(r *run) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.cur != r || r.ctx.Err() != nil {
		return
	}
	r.healthy = false
	h.recordFailureLocked()
	if !r.settled {
		r.settled = true
		close(r.ready)
	}
}

func (h *Hub) broadcast(r *run, p []byte) {
	chunk := make([]byte, len(p))
	copy(chunk, p)

	h.mu.Lock()
	defer h.mu.Unlock()
	if h.cur != r || r.ctx.Err() != nil {
		return
	}
	for s := range h.subs {
		select {
		case s.ch <- chunk:
		default:
		}
	}
}

func (h *Hub) loop(r *run, prev *run) {
	defer r.markReaped()
	if prev != nil {
		select {
		case <-prev.reaped:
		case <-r.ctx.Done():
			return
		}
	}

	backoff := h.cfg.InitialBackoff
	failures := 0
	for r.ctx.Err() == nil {
		gotData := h.once(r)
		if r.ctx.Err() != nil {
			return
		}
		if gotData {
			failures = 0
			backoff = h.cfg.InitialBackoff
		} else {
			failures++
			h.fail(r)
			if failures >= h.cfg.MaxFailures {
				h.abort(r)
				return
			}
		}
		if err := h.cfg.Sleep(r.ctx, backoff); err != nil {
			return
		}
		if !gotData {
			backoff = min(backoff*2, h.cfg.MaxBackoff)
		}
	}
}

func (h *Hub) once(r *run) bool {
	rc, err := h.cfg.Source(r.ctx)
	if err != nil {
		log.Printf("camera hub: start: %s", err)
		return false
	}
	src := &source{rc: rc}
	stop := context.AfterFunc(r.ctx, func() {
		src.close()
		r.markReaped()
	})
	defer func() {
		stop()
		src.close()
	}()

	gotData := false
	buf := make([]byte, readSize)
	for {
		n, err := rc.Read(buf)
		if n > 0 {
			if !gotData {
				gotData = true
				h.good(r)
			}
			h.broadcast(r, buf[:n])
		}
		if err != nil {
			return gotData
		}
	}
}
