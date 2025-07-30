package query

import (
	"context"
	"reflect"
	"sync"
	"time"

	buserror "go-cqrs/internal/error"
	"go-cqrs/internal/metrics"
)

type QueryMiddleware func(next QueryHandlerFunc) QueryHandlerFunc

type QueryHandlerFunc func(ctx context.Context, query Query) (interface{}, error)

func (f QueryHandlerFunc) Handle(ctx context.Context, query Query) (interface{}, error) {
	return f(ctx, query)
}

type Bus interface {
	Execute(ctx context.Context, query Query) (interface{}, error)
	Register(handler interface{}) error
	RegisterFunc(queryName string, fn func(ctx context.Context, query Query) (interface{}, error)) error
	IsRegistered(queryName string) bool
	GetRegisteredQueries() []string
	GetMetrics() metrics.BusMetrics
}

type DefaultBus struct {
	registry    *handlerRegistry
	middlewares []QueryMiddleware
	metrics     metrics.BusMetrics
}

func NewBus(middlewares ...QueryMiddleware) *DefaultBus {
	return &DefaultBus{
		registry:    newHandlerRegistry(),
		middlewares: middlewares,
		metrics:     metrics.NewDefaultMetrics(),
	}
}

func (b *DefaultBus) Execute(ctx context.Context, query Query) (interface{}, error) {
	if query == nil {
		return nil, buserror.NewDispatchError(
			buserror.ErrorCodeInvalidQuery,
			"query cannot be nil",
		)
	}

	queryName := query.QueryName()
	if queryName == "" {
		return nil, buserror.NewDispatchError(
			buserror.ErrorCodeInvalidQuery,
			"query name cannot be empty",
		)
	}

	baseHandler, exists := b.registry.GetHandler(queryName)
	if !exists {
		return nil, buserror.NewHandlerNotFoundError("query", queryName)
	}

	finalHandler := baseHandler
	for i := len(b.middlewares) - 1; i >= 0; i-- {
		finalHandler = b.middlewares[i](finalHandler)
	}

	start := time.Now()
	result, err := finalHandler(ctx, query)
	duration := time.Since(start)

	if err != nil {
		b.metrics.RecordFailure(queryName, duration)
	} else {
		b.metrics.RecordExecution(queryName, duration)
	}

	return result, err
}

func ExecuteTyped[T any](bus Bus, ctx context.Context, query IQuery[T]) (T, error) {
	var zero T
	genericQuery, ok := query.(Query)
	if !ok {
		return zero, buserror.NewDispatchError(
			buserror.ErrorCodeInvalidQuery,
			"query must implement the Query interface",
		)
	}
	result, err := bus.Execute(ctx, genericQuery)
	if err != nil {
		return zero, err
	}
	typedResult, ok := result.(T)
	if !ok {
		return zero, buserror.NewDispatchError(
			buserror.ErrorCodeDispatchFailed,
			"query result type mismatch",
		)
	}
	return typedResult, nil
}

func (b *DefaultBus) Register(handler interface{}) error {
	return b.registry.Register(handler)
}

func (b *DefaultBus) RegisterFunc(queryName string, fn func(ctx context.Context, query Query) (interface{}, error)) error {
	if queryName == "" {
		return buserror.NewHandlerRegistrationError("query name cannot be empty", nil)
	}
	if fn == nil {
		return buserror.NewHandlerRegistrationError("handler function cannot be nil", nil)
	}
	return b.registry.RegisterFunc(queryName, QueryHandlerFunc(fn))
}

func (b *DefaultBus) IsRegistered(queryName string) bool {
	return b.registry.IsRegistered(queryName)
}

func (b *DefaultBus) GetRegisteredQueries() []string {
	return b.registry.GetRegisteredQueries()
}

func (b *DefaultBus) GetMetrics() metrics.BusMetrics {
	return b.metrics
}

type handlerRegistry struct {
	mu       sync.RWMutex
	handlers map[string]QueryHandlerFunc
}

func newHandlerRegistry() *handlerRegistry {
	return &handlerRegistry{
		handlers: make(map[string]QueryHandlerFunc),
	}
}

func (r *handlerRegistry) Register(handler interface{}) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	handlerType := reflect.TypeOf(handler)
	if handlerType.Kind() != reflect.Ptr || handlerType.Elem().Kind() != reflect.Struct {
		return buserror.NewHandlerRegistrationError("handler must be a pointer to a struct", nil)
	}
	handleMethod, exists := handlerType.MethodByName("Handle")
	if !exists {
		return buserror.NewHandlerRegistrationError("handler must implement Handle method", nil)
	}
	methodType := handleMethod.Type
	if methodType.NumIn() != 3 || methodType.NumOut() != 2 {
		return buserror.NewHandlerRegistrationError("Handle method signature invalid", nil)
	}
	queryType := methodType.In(2)
	queryValue := reflect.New(queryType).Elem()
	if !queryValue.Type().Implements(reflect.TypeOf((*Query)(nil)).Elem()) {
		return buserror.NewHandlerRegistrationError("second parameter must implement Query interface", nil)
	}
	queryInterface := queryValue.Interface().(Query)
	queryName := queryInterface.QueryName()
	if _, exists := r.handlers[queryName]; exists {
		return buserror.NewHandlerRegistrationError("handler already registered for query: "+queryName, nil)
	}
	handlerValue := reflect.ValueOf(handler)
	r.handlers[queryName] = func(ctx context.Context, query Query) (interface{}, error) {
		results := handlerValue.MethodByName("Handle").Call([]reflect.Value{
			reflect.ValueOf(ctx),
			reflect.ValueOf(query),
		})
		var result interface{}
		var err error
		if len(results) >= 1 && results[0].IsValid() && results[0].CanInterface() {
			result = results[0].Interface()
		}
		if len(results) >= 2 && results[1].IsValid() && !results[1].IsNil() {
			err = results[1].Interface().(error)
		}
		return result, err
	}
	return nil
}

func (r *handlerRegistry) RegisterFunc(queryName string, fn QueryHandlerFunc) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.handlers[queryName]; exists {
		return buserror.NewHandlerRegistrationError("handler already registered for query: "+queryName, nil)
	}
	r.handlers[queryName] = fn
	return nil
}

func (r *handlerRegistry) GetHandler(queryName string) (QueryHandlerFunc, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	handler, exists := r.handlers[queryName]
	return handler, exists
}

func (r *handlerRegistry) IsRegistered(queryName string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, exists := r.handlers[queryName]
	return exists
}

func (r *handlerRegistry) GetRegisteredQueries() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	queries := make([]string, 0, len(r.handlers))
	for queryName := range r.handlers {
		queries = append(queries, queryName)
	}
	return queries
}
