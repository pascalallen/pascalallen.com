package messaging

import "context"

type Command interface {
	CommandName() string
}

type CommandHandler interface {
	Handle(ctx context.Context, command Command) error
}

type CommandBus interface {
	RegisterHandler(commandName string, handler CommandHandler)
	Execute(ctx context.Context, command Command) error
}

type Query interface {
	QueryName() string
}

type QueryHandler interface {
	Handle(ctx context.Context, query Query) (any, error)
}

type QueryBus interface {
	RegisterHandler(queryName string, handler QueryHandler)
	Fetch(ctx context.Context, query Query) (any, error)
}
