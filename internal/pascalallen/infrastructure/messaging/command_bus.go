package messaging

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"github.com/pascalallen/pascalallen.com/internal/pascalallen/application/messaging"
)

type SynchronousCommandBus struct {
	mu       sync.RWMutex
	handlers map[string]messaging.CommandHandler
	logger   *slog.Logger
}

func NewSynchronousCommandBus(logger *slog.Logger) messaging.CommandBus {
	return &SynchronousCommandBus{
		handlers: make(map[string]messaging.CommandHandler),
		logger:   logger,
	}
}

func (bus *SynchronousCommandBus) RegisterHandler(commandName string, handler messaging.CommandHandler) {
	bus.mu.Lock()
	defer bus.mu.Unlock()

	bus.logger.Info("registering command handler", "commandName", commandName)
	bus.handlers[commandName] = handler
}

func (bus *SynchronousCommandBus) Execute(ctx context.Context, command messaging.Command) error {
	bus.mu.RLock()
	handler, found := bus.handlers[command.CommandName()]
	bus.mu.RUnlock()

	if !found {
		bus.logger.Warn("no handler registered for command", "commandName", command.CommandName())
		return fmt.Errorf("no handler registered for command: %s", command.CommandName())
	}

	if err := handler.Handle(ctx, command); err != nil {
		return fmt.Errorf("handle command %s: %w", command.CommandName(), err)
	}

	return nil
}
