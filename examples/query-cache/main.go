package main

import (
	"context"
	"errors"
	"time"

	sharedmiddleware "go-cqrs/internal/middleware"
	"go-cqrs/internal/query"
	"go-cqrs/internal/query/middleware"
)

type Product struct {
	ID          int     `json:"id"`
	Name        string  `json:"name"`
	Price       float64 `json:"price"`
	Description string  `json:"description"`
}

type GetProductQuery struct {
	ProductID int
}

func (q GetProductQuery) QueryName() string {
	return "get_product"
}

type GetProductHandler struct{}

func (h *GetProductHandler) Handle(ctx context.Context, q GetProductQuery) (Product, error) {

	time.Sleep(500 * time.Millisecond)

	products := map[int]Product{
		1: {ID: 1, Name: "Laptop Pro", Price: 1299.99, Description: "High-performance laptop"},
		2: {ID: 2, Name: "Wireless Mouse", Price: 29.99, Description: "Ergonomic wireless mouse"},
		3: {ID: 3, Name: "Mechanical Keyboard", Price: 149.99, Description: "RGB mechanical keyboard"},
	}

	product, exists := products[q.ProductID]
	if !exists {
		return Product{}, errors.New("product not found")
	}

	return product, nil
}

type CacheLogger struct{}

func (l *CacheLogger) LogQueryStart(ctx context.Context, queryName string) {

}

func (l *CacheLogger) LogQuerySuccess(ctx context.Context, queryName string, duration time.Duration, result interface{}) {

}

func (l *CacheLogger) LogQueryError(ctx context.Context, queryName string, duration time.Duration, err error) {

}

type ProductCacheKeyGenerator struct{}

func (g *ProductCacheKeyGenerator) GenerateKey(ctx context.Context, q query.Query) string {
	if productQuery, ok := q.(GetProductQuery); ok {
		return "get_product:" + string(rune(productQuery.ProductID))
	}
	return q.QueryName()
}

func main() {

	logger := &CacheLogger{}
	cache := middleware.NewMemoryCache()
	keyGenerator := &ProductCacheKeyGenerator{}

	bus := query.NewBus(
		middleware.LoggingMiddleware(logger),
		sharedmiddleware.RecoveryQueryMiddleware(),
		middleware.TimeoutMiddleware(10*time.Second),
		middleware.CachingMiddleware(cache, keyGenerator, 30*time.Second),
	)

	err := bus.Register(&GetProductHandler{})
	if err != nil {
		return
	}

	ctx := context.Background()
	productQuery := GetProductQuery{ProductID: 1}

	start := time.Now()
	result, err := bus.Execute(ctx, productQuery)
	duration1 := time.Since(start)
	_ = result
	_ = err
	_ = duration1

	start = time.Now()
	result, err = bus.Execute(ctx, productQuery)
	duration2 := time.Since(start)
	_ = result
	_ = err
	_ = duration2

	start = time.Now()
	result, err = bus.Execute(ctx, productQuery)
	duration3 := time.Since(start)
	_ = result
	_ = err
	_ = duration3

	differentQuery := GetProductQuery{ProductID: 2}
	start = time.Now()
	result, err = bus.Execute(ctx, differentQuery)
	duration4 := time.Since(start)
	_ = result
	_ = err
	_ = duration4

	time.Sleep(1 * time.Second)

	start = time.Now()
	result, err = bus.Execute(ctx, productQuery)
	duration5 := time.Since(start)
	_ = result
	_ = err
	_ = duration5

	cache.Clear()

	start = time.Now()
	result, err = bus.Execute(ctx, productQuery)
	durationAfterClear := time.Since(start)
	_ = result
	_ = err
	_ = durationAfterClear

	product1Query := GetProductQuery{ProductID: 1}
	product10Query := GetProductQuery{ProductID: 10}

	_, err1 := bus.Execute(ctx, product1Query)
	_, err2 := bus.Execute(ctx, product10Query)
	_ = err1
	_ = err2
}
