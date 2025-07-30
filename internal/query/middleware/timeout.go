package middleware

import (
	"context"
	"time"

	buserror "go-cqrs/internal/error"
	"go-cqrs/internal/query"
)

func TimeoutMiddleware(timeout time.Duration) query.QueryMiddleware {
	return func(next query.QueryHandlerFunc) query.QueryHandlerFunc {
		return func(ctx context.Context, q query.Query) (interface{}, error) {

			ctx, cancel := context.WithTimeout(ctx, timeout)
			defer cancel()

			type result struct {
				data interface{}
				err  error
			}

			done := make(chan result, 1)
			go func() {
				data, err := next(ctx, q)
				done <- result{data: data, err: err}
			}()

			select {
			case res := <-done:
				return res.data, res.err
			case <-ctx.Done():
				return nil, buserror.NewDispatchErrorWithCause(
					buserror.ErrorCodeDispatchFailed,
					"query execution timed out",
					ctx.Err(),
				)
			}
		}
	}
}
