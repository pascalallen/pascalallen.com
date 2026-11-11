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
	readSize              = 65536
)

type Source func(ctx context.Context) (io.ReadCloser, error)

type Sleeper func(ctx context.Context, d time.Duration) error

type Config struct {
	Source         Source
	Check          func() bool
	Sleep          Sleeper
	Buffer         int
	InitialBackoff time.Duration
	MaxBackoff     time.Duration
	StartTimeout   time.Duration
}

type run struct {
	ready   chan struct{}
	settled bool
	ok      bool
}

type subscriber struct {
	ch chan []byte
}

type Hub struct {
	cfg    Config
	mu     sync.Mutex
	subs   map[*subscriber]struct{}
	cancel context.CancelFunc
	cur    *run
	failed bool
}

var _ camera.Stream = (*Hub)(nil)

func NewHub(cfg Config) *Hub {
	if cfg.Check == nil {
		cfg.Check = func() bool { return true }
	}
	if cfg.Sleep == nil {
		cfg.Sleep = sleep
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

	return !h.failed
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
	if h.cancel == nil {
		runCtx, cancel := context.WithCancel(context.Background())
		h.cancel = cancel
		h.cur = &run{ready: make(chan struct{})}
		go h.loop(runCtx, h.cur)
	}
	r := h.cur
	h.mu.Unlock()

	var once sync.Once
	var stopAfter func() bool
	unsubscribe := func() {
		once.Do(func() {
			stopAfter()
			h.remove(s)
		})
	}
	stopAfter = context.AfterFunc(ctx, unsubscribe)

	timer := time.NewTimer(h.cfg.StartTimeout)
	defer timer.Stop()
	select {
	case <-r.ready:
	case <-ctx.Done():
		unsubscribe()
		return nil, nil, ctx.Err()
	case <-timer.C:
		unsubscribe()
		return nil, nil, camera.ErrUnavailable
	}

	h.mu.Lock()
	ok := r.ok
	h.mu.Unlock()
	if !ok {
		unsubscribe()
		return nil, nil, camera.ErrUnavailable
	}

	return s.ch, unsubscribe, nil
}

func (h *Hub) remove(s *subscriber) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.subs, s)
	if len(h.subs) == 0 && h.cancel != nil {
		h.cancel()
		h.cancel = nil
	}
}

func (h *Hub) settle(ctx context.Context, r *run, ok bool) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if ctx.Err() != nil {
		return
	}
	h.failed = !ok
	if r.settled {
		return
	}
	r.settled = true
	r.ok = ok
	close(r.ready)
}

func (h *Hub) broadcast(ctx context.Context, p []byte) {
	chunk := make([]byte, len(p))
	copy(chunk, p)

	h.mu.Lock()
	defer h.mu.Unlock()
	if ctx.Err() != nil {
		return
	}
	for s := range h.subs {
		select {
		case s.ch <- chunk:
		default:
		}
	}
}

func (h *Hub) loop(ctx context.Context, r *run) {
	backoff := h.cfg.InitialBackoff
	for ctx.Err() == nil {
		gotData := h.once(ctx, r)
		if ctx.Err() != nil {
			return
		}
		if !gotData {
			h.settle(ctx, r, false)
		}
		if err := h.cfg.Sleep(ctx, backoff); err != nil {
			return
		}
		if gotData {
			backoff = h.cfg.InitialBackoff
		} else {
			backoff = min(backoff*2, h.cfg.MaxBackoff)
		}
	}
}

func (h *Hub) once(ctx context.Context, r *run) bool {
	rc, err := h.cfg.Source(ctx)
	if err != nil {
		log.Printf("camera hub: start: %s", err)
		return false
	}
	stop := context.AfterFunc(ctx, func() { _ = rc.Close() })
	defer func() {
		stop()
		_ = rc.Close()
	}()

	gotData := false
	buf := make([]byte, readSize)
	for {
		n, err := rc.Read(buf)
		if n > 0 {
			if !gotData {
				gotData = true
				h.settle(ctx, r, true)
			}
			h.broadcast(ctx, buf[:n])
		}
		if err != nil {
			return gotData
		}
	}
}
