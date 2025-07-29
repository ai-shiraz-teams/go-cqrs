package middleware

import (
	"context"
	"time"

	"go-cqrs/internal/command"
)

// Logger defines the interface for command logging middleware.
type Logger interface {
	LogCommandStart(ctx context.Context, commandName string)
	LogCommandSuccess(ctx context.Context, commandName string, duration time.Duration)
	LogCommandError(ctx context.Context, commandName string, duration time.Duration, err error)
}

// LoggingMiddleware creates a middleware that logs command execution details.
func LoggingMiddleware(logger Logger) command.Middleware {
	return func(next command.ExecutorFunc) command.ExecutorFunc {
		return func(ctx context.Context, cmd command.ICommand) error {
			commandName := cmd.CommandName()
			start := time.Now()

			logger.LogCommandStart(ctx, commandName)

			err := next(ctx, cmd)
			duration := time.Since(start)

			if err != nil {
				logger.LogCommandError(ctx, commandName, duration, err)
			} else {
				logger.LogCommandSuccess(ctx, commandName, duration)
			}

			return err
		}
	}
}
