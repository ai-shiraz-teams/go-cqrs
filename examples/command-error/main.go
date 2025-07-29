package main

import (
	"context"
	"errors"
	"time"

	"go-cqrs/internal/command"
	"go-cqrs/internal/command/middleware"
	sharedmiddleware "go-cqrs/internal/middleware"
)

type FailCommand struct {
	ErrorMessage string
}

func (c FailCommand) CommandName() string {
	return "fail"
}

type FailHandler struct{}

func (h *FailHandler) Handle(ctx context.Context, cmd FailCommand) error {
	if cmd.ErrorMessage == "" {
		return errors.New("intentional failure")
	}
	return errors.New(cmd.ErrorMessage)
}

type DetailedLogger struct{}

func (l *DetailedLogger) LogCommandStart(ctx context.Context, commandName string) {

}

func (l *DetailedLogger) LogCommandSuccess(ctx context.Context, commandName string, duration time.Duration) {

}

func (l *DetailedLogger) LogCommandError(ctx context.Context, commandName string, duration time.Duration, err error) {

}

type SimpleValidator struct{}

func (v *SimpleValidator) ValidateCommand(ctx context.Context, cmd command.ICommand) error {
	if failCmd, ok := cmd.(FailCommand); ok {
		if len(failCmd.ErrorMessage) > 100 {
			return errors.New("error message too long (max 100 characters)")
		}
	}
	return nil
}

func main() {

	logger := &DetailedLogger{}
	validator := &SimpleValidator{}

	bus := command.NewBus(
		middleware.LoggingMiddleware(logger),
		middleware.ValidationMiddleware(validator),
		sharedmiddleware.RecoveryCommandMiddleware(),
	)

	err := bus.Register(&FailHandler{})
	if err != nil {
		return
	}

	ctx := context.Background()

	cmd1 := FailCommand{}
	err = bus.Dispatch(ctx, cmd1)
	_ = err

	cmd2 := FailCommand{ErrorMessage: "custom validation failed"}
	err = bus.Dispatch(ctx, cmd2)
	_ = err

	cmd3 := FailCommand{ErrorMessage: "this is a very long error message that exceeds the maximum allowed length of 100 characters for testing validation failure"}
	err = bus.Dispatch(ctx, cmd3)
	_ = err

	registeredCommands := bus.GetRegisteredCommands()
	_ = registeredCommands

	isRegistered := bus.IsRegistered("fail")
	_ = isRegistered

	isNotRegistered := bus.IsRegistered("non-existent")
	_ = isNotRegistered
}
