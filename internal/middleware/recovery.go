package middleware

import (
	"context"

	"go-cqrs/internal/command"
	buserror "go-cqrs/internal/error"
	"go-cqrs/internal/query"
)

func RecoveryCommandMiddleware() command.Middleware {
	return func(next command.ExecutorFunc) command.ExecutorFunc {
		return func(ctx context.Context, cmd command.ICommand) (result interface{}, err error) {
			defer func() {
				if r := recover(); r != nil {
					result = nil
					switch x := r.(type) {
					case string:
						err = buserror.NewDispatchError(
							buserror.ErrorCodeDispatchFailed,
							"panic recovered during command execution: "+x,
						)
					case error:
						err = buserror.NewDispatchErrorWithCause(
							buserror.ErrorCodeDispatchFailed,
							"panic recovered during command execution",
							x,
						)
					default:
						err = buserror.NewDispatchError(
							buserror.ErrorCodeDispatchFailed,
							"panic recovered during command execution: unknown error",
						)
					}
				}
			}()

			return next(ctx, cmd)
		}
	}
}

func RecoveryQueryMiddleware() query.QueryMiddleware {
	return func(next query.QueryHandlerFunc) query.QueryHandlerFunc {
		return func(ctx context.Context, q query.Query) (result interface{}, err error) {
			defer func() {
				if r := recover(); r != nil {
					result = nil
					switch x := r.(type) {
					case string:
						err = buserror.NewDispatchError(
							buserror.ErrorCodeDispatchFailed,
							"panic recovered during query execution: "+x,
						)
					case error:
						err = buserror.NewDispatchErrorWithCause(
							buserror.ErrorCodeDispatchFailed,
							"panic recovered during query execution",
							x,
						)
					default:
						err = buserror.NewDispatchError(
							buserror.ErrorCodeDispatchFailed,
							"panic recovered during query execution: unknown error",
						)
					}
				}
			}()

			return next(ctx, q)
		}
	}
}
