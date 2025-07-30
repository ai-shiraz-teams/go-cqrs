package main

import (
	"context"
	"time"

	"github.com/ai-shiraz-teams/go-cqrs/internal/command"
	"github.com/ai-shiraz-teams/go-cqrs/internal/command/middleware"
	sharedmiddleware "github.com/ai-shiraz-teams/go-cqrs/internal/middleware"
)

type SayHelloCommand struct {
	Name string
}

func (c SayHelloCommand) CommandName() string {
	return "say_hello"
}

type SayHelloHandler struct{}

func (h *SayHelloHandler) Handle(ctx context.Context, cmd SayHelloCommand) (interface{}, error) {

	return "Hello, " + cmd.Name + "!", nil
}

type SimpleLogger struct{}

func (l *SimpleLogger) LogCommandStart(ctx context.Context, commandName string) {

}

func (l *SimpleLogger) LogCommandSuccess(ctx context.Context, commandName string, duration time.Duration) {

}

func (l *SimpleLogger) LogCommandError(ctx context.Context, commandName string, duration time.Duration, err error) {

}

func main() {

	logger := &SimpleLogger{}

	bus := command.NewBus(
		middleware.LoggingMiddleware(logger),
		sharedmiddleware.RecoveryCommandMiddleware(),
	)

	err := bus.Register(&SayHelloHandler{})
	if err != nil {
		return
	}

	ctx := context.Background()
	cmd := SayHelloCommand{Name: "World"}

	result, err := bus.Dispatch(ctx, cmd)
	_ = result
	_ = err
}
