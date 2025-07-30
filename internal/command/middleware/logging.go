package middleware

import (
	"context"
	"time"

	"github.com/ai-shiraz-teams/go-cqrs/internal/command"
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
		return func(ctx context.Context, cmd command.ICommand) (interface{}, error) {
			commandName := cmd.CommandName()
			start := time.Now()

			logger.LogCommandStart(ctx, commandName)

			result, err := next(ctx, cmd)
			duration := time.Since(start)

			if err != nil {
				logger.LogCommandError(ctx, commandName, duration, err)
			} else {
				logger.LogCommandSuccess(ctx, commandName, duration)
			}

			return result, err
		}
	}
}
