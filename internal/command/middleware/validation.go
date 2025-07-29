package middleware

import (
	"context"

	"go-cqrs/internal/command"
	buserror "go-cqrs/internal/error"
)

type Validator interface {
	ValidateCommand(ctx context.Context, cmd command.ICommand) error
}

func ValidationMiddleware(validator Validator) command.Middleware {
	return func(next command.ExecutorFunc) command.ExecutorFunc {
		return func(ctx context.Context, cmd command.ICommand) error {
			if err := validator.ValidateCommand(ctx, cmd); err != nil {
				return buserror.NewDispatchErrorWithCause(
					buserror.ErrorCodeInvalidCommand,
					"command validation failed",
					err,
				)
			}

			return next(ctx, cmd)
		}
	}
}

type ValidatorFunc func(ctx context.Context, cmd command.ICommand) error

func (f ValidatorFunc) ValidateCommand(ctx context.Context, cmd command.ICommand) error {
	return f(ctx, cmd)
}

type NoOpValidator struct{}

func (v *NoOpValidator) ValidateCommand(ctx context.Context, cmd command.ICommand) error {
	return nil
}
