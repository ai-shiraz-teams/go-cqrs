package command

import "context"

type ICommand interface {
	CommandName() string
}

type ICommandHandler[T ICommand] interface {
	Handle(ctx context.Context, command T) (interface{}, error)
}
