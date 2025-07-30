package main

import (
	"context"
	"fmt"
	"time"

	"github.com/ai-shiraz-teams/go-cqrs/internal/command"
	"github.com/ai-shiraz-teams/go-cqrs/internal/command/middleware"
	sharedmiddleware "github.com/ai-shiraz-teams/go-cqrs/internal/middleware"
)

// Test command that returns different types of results
type TestCommand struct {
	ID       int
	Message  string
	ReturnID bool
}

func (c TestCommand) CommandName() string {
	return "test_command"
}

type TestHandler struct{}

func (h *TestHandler) Handle(ctx context.Context, cmd TestCommand) (interface{}, error) {
	if cmd.ReturnID {
		// Return an ID
		return cmd.ID, nil
	}

	if cmd.Message != "" {
		// Return a complex object
		return map[string]interface{}{
			"id":        cmd.ID,
			"message":   cmd.Message,
			"timestamp": time.Now().Unix(),
		}, nil
	}

	// Return nil result
	return nil, nil
}

type TestLogger struct{}

func (l *TestLogger) LogCommandStart(ctx context.Context, commandName string) {
	fmt.Printf("🚀 Starting command: %s\n", commandName)
}

func (l *TestLogger) LogCommandSuccess(ctx context.Context, commandName string, duration time.Duration) {
	fmt.Printf("✅ Command %s completed in %v\n", commandName, duration)
}

func (l *TestLogger) LogCommandError(ctx context.Context, commandName string, duration time.Duration, err error) {
	fmt.Printf("❌ Command %s failed after %v: %v\n", commandName, duration, err)
}

type TestValidator struct{}

func (v *TestValidator) ValidateCommand(ctx context.Context, cmd command.ICommand) error {
	if testCmd, ok := cmd.(TestCommand); ok {
		if testCmd.ID < 0 {
			return fmt.Errorf("ID cannot be negative")
		}
	}
	return nil
}

func main() {
	fmt.Println("=== Testing Command Results Support ===\n")

	logger := &TestLogger{}
	validator := &TestValidator{}

	// Create bus with all middleware types
	bus := command.NewBus(
		middleware.LoggingMiddleware(logger),
		middleware.ValidationMiddleware(validator),
		sharedmiddleware.RecoveryCommandMiddleware(),
		middleware.TimeoutMiddleware(5*time.Second),
	)

	err := bus.Register(&TestHandler{})
	if err != nil {
		panic(fmt.Sprintf("Failed to register handler: %v", err))
	}

	ctx := context.Background()

	// Test 1: Command returning an integer ID
	fmt.Println("📝 Test 1: Command returning integer ID")
	result, err := bus.Dispatch(ctx, TestCommand{ID: 42, ReturnID: true})
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
	} else {
		fmt.Printf("📤 Result: %v (type: %T)\n", result, result)
	}
	fmt.Println()

	// Test 2: Command returning complex object
	fmt.Println("📝 Test 2: Command returning complex object")
	result, err = bus.Dispatch(ctx, TestCommand{ID: 123, Message: "Hello World"})
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
	} else {
		fmt.Printf("📤 Result: %v (type: %T)\n", result, result)
		if resultMap, ok := result.(map[string]interface{}); ok {
			fmt.Printf("   - ID: %v\n", resultMap["id"])
			fmt.Printf("   - Message: %v\n", resultMap["message"])
			fmt.Printf("   - Timestamp: %v\n", resultMap["timestamp"])
		}
	}
	fmt.Println()

	// Test 3: Command returning nil result
	fmt.Println("📝 Test 3: Command returning nil result")
	result, err = bus.Dispatch(ctx, TestCommand{ID: 0})
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
	} else {
		fmt.Printf("📤 Result: %v (type: %T)\n", result, result)
		if result == nil {
			fmt.Printf("   ✅ Nil result correctly preserved through middleware chain\n")
		}
	}
	fmt.Println()

	// Test 4: Command failing validation (should return nil result)
	fmt.Println("📝 Test 4: Command failing validation")
	result, err = bus.Dispatch(ctx, TestCommand{ID: -1, ReturnID: true})
	if err != nil {
		fmt.Printf("❌ Error (expected): %v\n", err)
		fmt.Printf("📤 Result: %v (type: %T)\n", result, result)
		if result == nil {
			fmt.Printf("   ✅ Nil result correctly returned on validation failure\n")
		}
	} else {
		fmt.Printf("❌ Expected error but got success with result: %v\n", result)
	}

	fmt.Println("\n=== Command Results Support Test Complete ===")
}
