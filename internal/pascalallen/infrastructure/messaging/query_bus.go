package messaging

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"github.com/pascalallen/pascalallen.com/internal/pascalallen/application/messaging"
)

type SynchronousQueryBus struct {
	mu       sync.RWMutex
	handlers map[string]messaging.QueryHandler
	logger   *slog.Logger
}

func NewSynchronousQueryBus(logger *slog.Logger) messaging.QueryBus {
	return &SynchronousQueryBus{
		handlers: make(map[string]messaging.QueryHandler),
		logger:   logger,
	}
}

func (bus *SynchronousQueryBus) RegisterHandler(queryName string, handler messaging.QueryHandler) {
	bus.mu.Lock()
	defer bus.mu.Unlock()

	bus.logger.Info("registering query handler", "queryName", queryName)
	bus.handlers[queryName] = handler
}

func (bus *SynchronousQueryBus) Fetch(ctx context.Context, query messaging.Query) (any, error) {
	bus.mu.RLock()
	handler, found := bus.handlers[query.QueryName()]
	bus.mu.RUnlock()

	if !found {
		bus.logger.Warn("no handler registered for query", "queryName", query.QueryName())
		return nil, fmt.Errorf("no handler registered for query: %s", query.QueryName())
	}

	result, err := handler.Handle(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("handle query %s: %w", query.QueryName(), err)
	}

	return result, nil
}
