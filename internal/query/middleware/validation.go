package middleware

import (
	"context"

	buserror "go-cqrs/internal/error"
	"go-cqrs/internal/query"
)

// Validator defines the interface for query validation middleware.
type Validator interface {
	ValidateQuery(ctx context.Context, q query.Query) error
}

// ValidationMiddleware creates a middleware that validates queries before execution.
func ValidationMiddleware(validator Validator) query.QueryMiddleware {
	return func(next query.QueryHandlerFunc) query.QueryHandlerFunc {
		return func(ctx context.Context, q query.Query) (interface{}, error) {
			if err := validator.ValidateQuery(ctx, q); err != nil {
				return nil, buserror.NewDispatchErrorWithCause(
					buserror.ErrorCodeInvalidQuery,
					"query validation failed",
					err,
				)
			}
			return next(ctx, q)
		}
	}
}
