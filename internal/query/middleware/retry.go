package middleware

import (
	"context"
	"time"

	buserror "go-cqrs/internal/error"
	"go-cqrs/internal/query"
)

func RetryMiddleware(maxRetries int, baseDelay time.Duration) query.QueryMiddleware {
	return func(next query.QueryHandlerFunc) query.QueryHandlerFunc {
		return func(ctx context.Context, q query.Query) (interface{}, error) {
			var lastResult interface{}
			var lastErr error

			for attempt := 0; attempt <= maxRetries; attempt++ {

				result, err := next(ctx, q)
				if err == nil {
					return result, nil
				}

				lastResult = result
				lastErr = err

				if attempt == maxRetries {
					break
				}

				if !shouldRetryQuery(err) {
					return result, err
				}

				delay := time.Duration(attempt+1) * baseDelay

				select {
				case <-time.After(delay):

				case <-ctx.Done():
					return nil, buserror.NewDispatchErrorWithCause(
						buserror.ErrorCodeDispatchFailed,
						"query retry cancelled due to context cancellation",
						ctx.Err(),
					)
				}
			}

			return lastResult, buserror.NewDispatchErrorWithCause(
				buserror.ErrorCodeDispatchFailed,
				"query failed after retries",
				lastErr,
			)
		}
	}
}

func shouldRetryQuery(err error) bool {

	if busErr, ok := err.(buserror.BusError); ok {
		switch busErr.ErrorCode() {
		case buserror.ErrorCodeDispatchFailed:
			return true
		case buserror.ErrorCodeHandlerNotFound:
			return false
		case buserror.ErrorCodeInvalidQuery:
			return false
		default:
			return false
		}
	}
	return false
}
