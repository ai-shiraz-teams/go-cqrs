package middleware

import (
	"context"

	buserror "go-cqrs/internal/error"
	"go-cqrs/internal/query"
)

type Validator interface {
	ValidateQuery(ctx context.Context, q query.Query) error
}

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

type QueryValidatorFunc func(ctx context.Context, q query.Query) error

func (f QueryValidatorFunc) ValidateQuery(ctx context.Context, q query.Query) error {
	return f(ctx, q)
}

type NoOpValidator struct{}

func (v *NoOpValidator) ValidateQuery(ctx context.Context, q query.Query) error {
	return nil
}
