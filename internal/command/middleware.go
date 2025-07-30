package command

import "context"

type Middleware func(next ExecutorFunc) ExecutorFunc

type ExecutorFunc func(ctx context.Context, cmd ICommand) (interface{}, error)

type MiddlewareChain struct {
	middlewares []Middleware
}

func NewMiddlewareChain(middlewares ...Middleware) *MiddlewareChain {
	chain := &MiddlewareChain{
		middlewares: make([]Middleware, len(middlewares)),
	}
	copy(chain.middlewares, middlewares)
	return chain
}

func (c *MiddlewareChain) Add(middleware Middleware) {
	c.middlewares = append(c.middlewares, middleware)
}

func (c *MiddlewareChain) Execute(finalHandler ExecutorFunc) ExecutorFunc {

	result := finalHandler
	for i := len(c.middlewares) - 1; i >= 0; i-- {
		result = c.middlewares[i](result)
	}
	return result
}
