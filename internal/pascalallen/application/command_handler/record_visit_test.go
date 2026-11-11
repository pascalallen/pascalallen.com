package command_handler

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/pascalallen/pascalallen.com/internal/pascalallen/application/command"
	"github.com/pascalallen/pascalallen.com/internal/pascalallen/domain/visit"
)

type fakeNotifier struct {
	mu     sync.Mutex
	visits []visit.Visit
	err    error
}

func (f *fakeNotifier) Notify(_ context.Context, v visit.Visit) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.visits = append(f.visits, v)

	return f.err
}

type fakeClock struct{ now time.Time }

func (c *fakeClock) Now() time.Time { return c.now }

func newHandler(n *fakeNotifier, c *fakeClock) *RecordVisitHandler {
	return &RecordVisitHandler{Notifier: n, Now: c.Now}
}

func TestRecordVisitHandler(t *testing.T) {
	start := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)

	t.Run("it notifies once for a first visit", func(t *testing.T) {
		n := &fakeNotifier{}
		h := newHandler(n, &fakeClock{now: start})

		err := h.Handle(context.Background(), &command.RecordVisit{Language: "en-US", UserAgent: "ua", ClientIP: "1.1.1.1"})

		require.NoError(t, err)
		require.Len(t, n.visits, 1)
		assert.Equal(t, visit.Visit{Language: "en-US", UserAgent: "ua"}, n.visits[0])
	})

	t.Run("it notifies at most once per client within the hour", func(t *testing.T) {
		n := &fakeNotifier{}
		clock := &fakeClock{now: start}
		h := newHandler(n, clock)
		cmd := &command.RecordVisit{ClientIP: "1.1.1.1"}

		require.NoError(t, h.Handle(context.Background(), cmd))
		clock.now = start.Add(59 * time.Minute)
		require.NoError(t, h.Handle(context.Background(), cmd))

		assert.Len(t, n.visits, 1)
	})

	t.Run("it notifies again for the same client after the hour", func(t *testing.T) {
		n := &fakeNotifier{}
		clock := &fakeClock{now: start}
		h := newHandler(n, clock)
		cmd := &command.RecordVisit{ClientIP: "1.1.1.1"}

		require.NoError(t, h.Handle(context.Background(), cmd))
		clock.now = start.Add(time.Hour)
		require.NoError(t, h.Handle(context.Background(), cmd))

		assert.Len(t, n.visits, 2)
	})

	t.Run("it notifies separately for different clients", func(t *testing.T) {
		n := &fakeNotifier{}
		h := newHandler(n, &fakeClock{now: start})

		require.NoError(t, h.Handle(context.Background(), &command.RecordVisit{ClientIP: "1.1.1.1"}))
		require.NoError(t, h.Handle(context.Background(), &command.RecordVisit{ClientIP: "2.2.2.2"}))

		assert.Len(t, n.visits, 2)
	})

	t.Run("it truncates inputs to 200 characters", func(t *testing.T) {
		n := &fakeNotifier{}
		h := newHandler(n, &fakeClock{now: start})
		long := strings.Repeat("a", 500)

		require.NoError(t, h.Handle(context.Background(), &command.RecordVisit{Language: long, UserAgent: long, ClientIP: "1.1.1.1"}))

		assert.Len(t, n.visits[0].Language, 200)
		assert.Len(t, n.visits[0].UserAgent, 200)
	})

	t.Run("it truncates by characters not bytes", func(t *testing.T) {
		n := &fakeNotifier{}
		h := newHandler(n, &fakeClock{now: start})
		long := strings.Repeat("é", 300)

		require.NoError(t, h.Handle(context.Background(), &command.RecordVisit{UserAgent: long, ClientIP: "1.1.1.1"}))

		assert.Equal(t, 200, len([]rune(n.visits[0].UserAgent)))
	})

	t.Run("it strips control characters", func(t *testing.T) {
		n := &fakeNotifier{}
		h := newHandler(n, &fakeClock{now: start})

		require.NoError(t, h.Handle(context.Background(), &command.RecordVisit{Language: "e\nn\r\x00", UserAgent: "a\tb\x1b[31m", ClientIP: "1.1.1.1"}))

		assert.Equal(t, "en", n.visits[0].Language)
		assert.Equal(t, "ab[31m", n.visits[0].UserAgent)
	})

	t.Run("it strips format and line separator characters", func(t *testing.T) {
		n := &fakeNotifier{}
		h := newHandler(n, &fakeClock{now: start})

		require.NoError(t, h.Handle(context.Background(), &command.RecordVisit{Language: "a‮b​c", UserAgent: "x y z", ClientIP: "1.1.1.1"}))

		assert.Equal(t, "abc", n.visits[0].Language)
		assert.Equal(t, "xyz", n.visits[0].UserAgent)
	})

	t.Run("it caps notifications across all clients within the window", func(t *testing.T) {
		n := &fakeNotifier{}
		clock := &fakeClock{now: start}
		h := &RecordVisitHandler{Notifier: n, Now: clock.Now, GlobalLimit: 3}

		for i := 0; i < 6; i++ {
			require.NoError(t, h.Handle(context.Background(), &command.RecordVisit{ClientIP: fmt.Sprintf("10.0.0.%d", i)}))
		}
		assert.Len(t, n.visits, 3)

		clock.now = start.Add(time.Hour)
		require.NoError(t, h.Handle(context.Background(), &command.RecordVisit{ClientIP: "10.0.1.1"}))
		assert.Len(t, n.visits, 4)
	})

	t.Run("it applies a default global cap of 30 per window", func(t *testing.T) {
		n := &fakeNotifier{}
		h := newHandler(n, &fakeClock{now: start})

		for i := 0; i < 40; i++ {
			require.NoError(t, h.Handle(context.Background(), &command.RecordVisit{ClientIP: fmt.Sprintf("10.0.0.%d", i)}))
		}

		assert.Len(t, n.visits, 30)
	})

	t.Run("it bounds the rate limit map and skips new clients when full of fresh entries", func(t *testing.T) {
		n := &fakeNotifier{}
		h := &RecordVisitHandler{Notifier: n, Now: (&fakeClock{now: start}).Now, MaxClients: 3}

		for i := 0; i < 5; i++ {
			require.NoError(t, h.Handle(context.Background(), &command.RecordVisit{ClientIP: fmt.Sprintf("10.0.0.%d", i)}))
		}

		assert.Len(t, n.visits, 3)
	})

	t.Run("it evicts expired entries to make room when the map is full", func(t *testing.T) {
		n := &fakeNotifier{}
		clock := &fakeClock{now: start}
		h := &RecordVisitHandler{Notifier: n, Now: clock.Now, MaxClients: 2}

		require.NoError(t, h.Handle(context.Background(), &command.RecordVisit{ClientIP: "a"}))
		require.NoError(t, h.Handle(context.Background(), &command.RecordVisit{ClientIP: "b"}))
		clock.now = start.Add(2 * time.Hour)
		require.NoError(t, h.Handle(context.Background(), &command.RecordVisit{ClientIP: "c"}))

		assert.Len(t, n.visits, 3)
	})

	t.Run("it returns the notifier error wrapped", func(t *testing.T) {
		boom := errors.New("boom")
		h := newHandler(&fakeNotifier{err: boom}, &fakeClock{now: start})

		err := h.Handle(context.Background(), &command.RecordVisit{ClientIP: "1.1.1.1"})

		assert.ErrorIs(t, err, boom)
	})

	t.Run("it rejects a command of another type", func(t *testing.T) {
		h := newHandler(&fakeNotifier{}, &fakeClock{now: start})

		assert.Error(t, h.Handle(context.Background(), nil))
	})
}
