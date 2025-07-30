package main

import (
	"context"
	"fmt"
	"time"

	"github.com/ai-shiraz-teams/go-cqrs/internal/command"
	"github.com/ai-shiraz-teams/go-cqrs/internal/command/middleware"
	sharedmiddleware "github.com/ai-shiraz-teams/go-cqrs/internal/middleware"
)

// Command that returns a result - user ID after creation
type CreateUserCommand struct {
	Name  string
	Email string
}

func (c CreateUserCommand) CommandName() string {
	return "create_user"
}

// User represents the created user
type User struct {
	ID      int       `json:"id"`
	Name    string    `json:"name"`
	Email   string    `json:"email"`
	Created time.Time `json:"created"`
}

// Handler that returns the created user as a result
type CreateUserHandler struct{}

func (h *CreateUserHandler) Handle(ctx context.Context, cmd CreateUserCommand) (interface{}, error) {
	// Simulate user creation with business logic
	user := User{
		ID:      42, // In real world, this would come from database
		Name:    cmd.Name,
		Email:   cmd.Email,
		Created: time.Now(),
	}

	// Command returns the created user as result
	return user, nil
}

type CommandLogger struct{}

func (l *CommandLogger) LogCommandStart(ctx context.Context, commandName string) {
	fmt.Printf("[LOG] Starting command: %s\n", commandName)
}

func (l *CommandLogger) LogCommandSuccess(ctx context.Context, commandName string, duration time.Duration) {
	fmt.Printf("[LOG] Command %s completed successfully in %v\n", commandName, duration)
}

func (l *CommandLogger) LogCommandError(ctx context.Context, commandName string, duration time.Duration, err error) {
	fmt.Printf("[LOG] Command %s failed after %v: %v\n", commandName, duration, err)
}

func main() {
	logger := &CommandLogger{}

	// Create command bus with middleware
	bus := command.NewBus(
		middleware.LoggingMiddleware(logger),
		sharedmiddleware.RecoveryCommandMiddleware(),
		middleware.TimeoutMiddleware(5*time.Second),
	)

	// Register the handler
	err := bus.Register(&CreateUserHandler{})
	if err != nil {
		panic(fmt.Sprintf("Failed to register handler: %v", err))
	}

	// Execute command and get result
	ctx := context.Background()
	cmd := CreateUserCommand{
		Name:  "John Doe",
		Email: "john@example.com",
	}

	result, err := bus.Dispatch(ctx, cmd)
	if err != nil {
		panic(fmt.Sprintf("Command execution failed: %v", err))
	}

	// Type assert and use the returned result
	if user, ok := result.(User); ok {
		fmt.Printf("✅ User created successfully!\n")
		fmt.Printf("   ID: %d\n", user.ID)
		fmt.Printf("   Name: %s\n", user.Name)
		fmt.Printf("   Email: %s\n", user.Email)
		fmt.Printf("   Created: %v\n", user.Created.Format(time.RFC3339))
	} else {
		fmt.Printf("❌ Unexpected result type: %T\n", result)
	}
}
