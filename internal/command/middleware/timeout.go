package middleware

import (
	"context"
	"time"

	"go-cqrs/internal/command"
	buserror "go-cqrs/internal/error"
)

func TimeoutMiddleware(timeout time.Duration) command.Middleware {
	return func(next command.ExecutorFunc) command.ExecutorFunc {
		return func(ctx context.Context, cmd command.ICommand) error {

			ctx, cancel := context.WithTimeout(ctx, timeout)
			defer cancel()

			done := make(chan error, 1)
			go func() {
				done <- next(ctx, cmd)
			}()

			select {
			case err := <-done:
				return err
			case <-ctx.Done():
				return buserror.NewDispatchErrorWithCause(
					buserror.ErrorCodeDispatchFailed,
					"command execution timed out",
					ctx.Err(),
				)
			}
		}
	}
}
