package middleware

import (
	"context"
	"time"

	"github.com/ai-shiraz-teams/go-cqrs/internal/query"
)

// Logger defines the interface for query logging middleware.
type Logger interface {
	LogQueryStart(ctx context.Context, queryName string)
	LogQuerySuccess(ctx context.Context, queryName string, duration time.Duration, result interface{})
	LogQueryError(ctx context.Context, queryName string, duration time.Duration, err error)
}

// LoggingMiddleware creates a middleware that logs query execution details.
func LoggingMiddleware(logger Logger) query.QueryMiddleware {
	return func(next query.QueryHandlerFunc) query.QueryHandlerFunc {
		return func(ctx context.Context, q query.Query) (interface{}, error) {
			queryName := q.QueryName()
			start := time.Now()

			logger.LogQueryStart(ctx, queryName)

			result, err := next(ctx, q)
			duration := time.Since(start)

			if err != nil {
				logger.LogQueryError(ctx, queryName, duration, err)
			} else {
				logger.LogQuerySuccess(ctx, queryName, duration, result)
			}

			return result, err
		}
	}
}
