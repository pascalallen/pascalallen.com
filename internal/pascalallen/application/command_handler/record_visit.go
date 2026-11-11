package command_handler

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"
	"unicode"

	"github.com/pascalallen/pascalallen.com/internal/pascalallen/application/command"
	"github.com/pascalallen/pascalallen.com/internal/pascalallen/application/messaging"
	"github.com/pascalallen/pascalallen.com/internal/pascalallen/domain/visit"
)

const (
	maxFieldLength   = 200
	defaultWindow    = time.Hour
	defaultMaxClient = 10000
)

type RecordVisitHandler struct {
	Notifier   visit.Notifier
	Now        func() time.Time
	Window     time.Duration
	MaxClients int

	mu   sync.Mutex
	seen map[string]time.Time
}

func (h *RecordVisitHandler) Handle(ctx context.Context, cmd messaging.Command) error {
	c, ok := cmd.(*command.RecordVisit)
	if !ok {
		return fmt.Errorf("invalid command type: %T", cmd)
	}

	if !h.admit(c.ClientIP) {
		return nil
	}

	v := visit.Visit{Language: sanitize(c.Language), UserAgent: sanitize(c.UserAgent)}
	if err := h.Notifier.Notify(ctx, v); err != nil {
		return fmt.Errorf("notify visit: %w", err)
	}

	return nil
}

func (h *RecordVisitHandler) admit(clientIP string) bool {
	window := h.Window
	if window == 0 {
		window = defaultWindow
	}
	limit := h.MaxClients
	if limit == 0 {
		limit = defaultMaxClient
	}
	now := h.Now()

	h.mu.Lock()
	defer h.mu.Unlock()

	if h.seen == nil {
		h.seen = make(map[string]time.Time)
	}

	if last, found := h.seen[clientIP]; found && now.Sub(last) < window {
		return false
	}

	if _, found := h.seen[clientIP]; !found && len(h.seen) >= limit {
		for ip, last := range h.seen {
			if now.Sub(last) >= window {
				delete(h.seen, ip)
			}
		}
		if len(h.seen) >= limit {
			return false
		}
	}

	h.seen[clientIP] = now

	return true
}

func sanitize(s string) string {
	stripped := strings.Map(func(r rune) rune {
		if unicode.IsControl(r) {
			return -1
		}

		return r
	}, s)

	runes := []rune(stripped)
	if len(runes) > maxFieldLength {
		runes = runes[:maxFieldLength]
	}

	return string(runes)
}
