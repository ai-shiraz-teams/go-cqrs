package main

import (
	"context"
	"errors"
	"time"

	sharedmiddleware "go-cqrs/internal/middleware"
	"go-cqrs/internal/query"
	"go-cqrs/internal/query/middleware"
)

type User struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type GetUserQuery struct {
	ID int
}

func (q GetUserQuery) QueryName() string {
	return "get_user"
}

type GetUserHandler struct{}

func (h *GetUserHandler) Handle(ctx context.Context, q GetUserQuery) (User, error) {

	users := map[int]User{
		1: {ID: 1, Name: "Alice Johnson", Email: "alice@example.com"},
		2: {ID: 2, Name: "Bob Smith", Email: "bob@example.com"},
		3: {ID: 3, Name: "Charlie Brown", Email: "charlie@example.com"},
	}

	user, exists := users[q.ID]
	if !exists {
		return User{}, errors.New("user not found")
	}

	return user, nil
}

type QueryLogger struct{}

func (l *QueryLogger) LogQueryStart(ctx context.Context, queryName string) {

}

func (l *QueryLogger) LogQuerySuccess(ctx context.Context, queryName string, duration time.Duration, result interface{}) {

}

func (l *QueryLogger) LogQueryError(ctx context.Context, queryName string, duration time.Duration, err error) {

}

func main() {

	logger := &QueryLogger{}

	bus := query.NewBus(
		middleware.LoggingMiddleware(logger),
		sharedmiddleware.RecoveryQueryMiddleware(),
		middleware.TimeoutMiddleware(5*time.Second),
	)

	err := bus.Register(&GetUserHandler{})
	if err != nil {
		return
	}

	ctx := context.Background()

	query1 := GetUserQuery{ID: 1}
	result, err := bus.Execute(ctx, query1)
	_ = result
	_ = err

	query2 := GetUserQuery{ID: 2}
	result, err = bus.Execute(ctx, query2)
	_ = result
	_ = err

	query3 := GetUserQuery{ID: 999}
	result, err = bus.Execute(ctx, query3)
	_ = result
	_ = err

	typedResult, err := query.ExecuteTyped[User](bus, ctx, query1)
	_ = typedResult
	_ = err

	registeredQueries := bus.GetRegisteredQueries()
	_ = registeredQueries

	isRegistered := bus.IsRegistered("get_user")
	_ = isRegistered

	isNotRegistered := bus.IsRegistered("non-existent")
	_ = isNotRegistered
}
