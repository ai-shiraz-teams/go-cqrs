package middleware

import (
	"context"

	"go-cqrs/internal/command"
	buserror "go-cqrs/internal/error"
)

// Validator defines the interface for command validation middleware.
type Validator interface {
	ValidateCommand(ctx context.Context, cmd command.ICommand) error
}

// ValidationMiddleware creates a middleware that validates commands before execution.
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
