package query

import "context"

type Query interface {
	QueryName() string
}

type IQuery[TResult any] interface {
	QueryName() string
}

type QueryHandler interface {
	Handle(ctx context.Context, query Query) (interface{}, error)
}

type IQueryHandler[TQuery IQuery[TResult], TResult any] interface {
	Handle(ctx context.Context, query TQuery) (TResult, error)
}
