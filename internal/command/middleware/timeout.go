package middleware

import (
	"context"
	"time"

	"github.com/ai-shiraz-teams/go-cqrs/internal/command"
	buserror "github.com/ai-shiraz-teams/go-cqrs/internal/error"
)

func TimeoutMiddleware(timeout time.Duration) command.Middleware {
	return func(next command.ExecutorFunc) command.ExecutorFunc {
		return func(ctx context.Context, cmd command.ICommand) (interface{}, error) {

			ctx, cancel := context.WithTimeout(ctx, timeout)
			defer cancel()

			done := make(chan struct {
				result interface{}
				err    error
			}, 1)
			go func() {
				result, err := next(ctx, cmd)
				done <- struct {
					result interface{}
					err    error
				}{result: result, err: err}
			}()

			select {
			case res := <-done:
				return res.result, res.err
			case <-ctx.Done():
				return nil, buserror.NewDispatchErrorWithCause(
					buserror.ErrorCodeDispatchFailed,
					"command execution timed out",
					ctx.Err(),
				)
			}
		}
	}
}
