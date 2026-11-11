package messaging

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	appmessaging "github.com/pascalallen/pascalallen.com/internal/pascalallen/application/messaging"
)

type testCommand struct{}

func (c *testCommand) CommandName() string { return "testCommand" }

type testQuery struct{}

func (q *testQuery) QueryName() string { return "testQuery" }

type commandHandlerFunc func(ctx context.Context, c appmessaging.Command) error

func (f commandHandlerFunc) Handle(ctx context.Context, c appmessaging.Command) error {
	return f(ctx, c)
}

type queryHandlerFunc func(ctx context.Context, q appmessaging.Query) (any, error)

func (f queryHandlerFunc) Handle(ctx context.Context, q appmessaging.Query) (any, error) {
	return f(ctx, q)
}

func discardLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

type ctxKey struct{}

func TestSynchronousCommandBus(t *testing.T) {
	t.Run("it runs the registered handler before returning with the same context", func(t *testing.T) {
		bus := NewSynchronousCommandBus(discardLogger())
		var seen any
		bus.RegisterHandler("testCommand", commandHandlerFunc(func(ctx context.Context, _ appmessaging.Command) error {
			seen = ctx.Value(ctxKey{})
			return nil
		}))

		err := bus.Execute(context.WithValue(context.Background(), ctxKey{}, "v"), &testCommand{})

		require.NoError(t, err)
		assert.Equal(t, "v", seen)
	})

	t.Run("it returns the handler error wrapped", func(t *testing.T) {
		boom := errors.New("boom")
		bus := NewSynchronousCommandBus(discardLogger())
		bus.RegisterHandler("testCommand", commandHandlerFunc(func(context.Context, appmessaging.Command) error { return boom }))

		err := bus.Execute(context.Background(), &testCommand{})

		assert.ErrorIs(t, err, boom)
	})

	t.Run("it fails when no handler is registered", func(t *testing.T) {
		bus := NewSynchronousCommandBus(discardLogger())

		assert.Error(t, bus.Execute(context.Background(), &testCommand{}))
	})
}

func TestSynchronousQueryBus(t *testing.T) {
	t.Run("it returns the handler result with the same context", func(t *testing.T) {
		bus := NewSynchronousQueryBus(discardLogger())
		bus.RegisterHandler("testQuery", queryHandlerFunc(func(ctx context.Context, _ appmessaging.Query) (any, error) {
			return ctx.Value(ctxKey{}), nil
		}))

		got, err := bus.Fetch(context.WithValue(context.Background(), ctxKey{}, "v"), &testQuery{})

		require.NoError(t, err)
		assert.Equal(t, "v", got)
	})

	t.Run("it returns the handler error wrapped", func(t *testing.T) {
		boom := errors.New("boom")
		bus := NewSynchronousQueryBus(discardLogger())
		bus.RegisterHandler("testQuery", queryHandlerFunc(func(context.Context, appmessaging.Query) (any, error) { return nil, boom }))

		_, err := bus.Fetch(context.Background(), &testQuery{})

		assert.ErrorIs(t, err, boom)
	})

	t.Run("it fails when no handler is registered", func(t *testing.T) {
		bus := NewSynchronousQueryBus(discardLogger())

		_, err := bus.Fetch(context.Background(), &testQuery{})

		assert.Error(t, err)
	})
}
