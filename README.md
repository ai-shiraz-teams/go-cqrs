# Go CQRS (Command Query Responsibility Segregation)

A lightweight and flexible Command Query Responsibility Segregation (CQRS) implementation for Go applications. This library provides a clean separation between commands and queries, enabling better code organization, testability, and maintainability.

## Features

- **Type-safe Command and Query handling**: Leverage Go's type system for compile-time safety
- **Command results support**: Commands can optionally return results (IDs, confirmations, created entities, etc.)
- **Middleware support**: Easily add cross-cutting concerns like logging, validation, and recovery
- **Flexible registration**: Register handlers with compile-time type checking
- **Clean architecture**: Promote separation of concerns in your applications
- **Zero dependencies**: Built using only Go standard library
- **Performance focused**: Minimal overhead with efficient handler resolution
- **Context support**: Full context.Context integration for cancellation and timeouts
- **Error handling**: Comprehensive error handling with custom error types

## Quick Start

### Installation

```bash
go get github.com/your-username/go-cqrs
```

### Basic Usage

#### Command Example

```go
package main

import (
    "context"
    "fmt"
    "go-cqrs/internal/command"
)

// Define your command
type CreateUserCommand struct {
    Name  string
    Email string
}

func (c CreateUserCommand) CommandName() string {
    return "create_user"
}

// Define your result type
type CreatedUser struct {
    ID    int    `json:"id"`
    Name  string `json:"name"`
    Email string `json:"email"`
}

// Define your handler - commands can return results!
type CreateUserHandler struct{}

func (h *CreateUserHandler) Handle(ctx context.Context, cmd CreateUserCommand) (interface{}, error) {
    // Your business logic here
    // Process the command (save to database, send email, etc.)
    
    // Commands can return results (e.g., created entity, ID, confirmation, etc.)
    user := CreatedUser{
        ID:    42, // In real world, this would come from database
        Name:  cmd.Name,
        Email: cmd.Email,
    }
    
    return user, nil
}

func main() {
    // Create command bus
    bus := command.NewBus()
    
    // Register handler
    bus.Register(&CreateUserHandler{})
    
    // Execute command and receive result
    ctx := context.Background()
    cmd := CreateUserCommand{Name: "John Doe", Email: "john@example.com"}
    result, err := bus.Dispatch(ctx, cmd)
    if err != nil {
        panic(err)
    }
    
    // Type assert and use the returned result
    if user, ok := result.(CreatedUser); ok {
        fmt.Printf("User created with ID: %d\n", user.ID)
    }
}
```

#### Query Example

```go
package main

import (
    "context"
    "fmt"
    "go-cqrs/internal/query"
)

// Define your query
type GetUserQuery struct {
    UserID int
}

func (q GetUserQuery) QueryName() string {
    return "get_user"
}

// Define your result
type User struct {
    ID    int
    Name  string
    Email string
}

// Define your handler
type GetUserHandler struct{}

func (h *GetUserHandler) Handle(ctx context.Context, q GetUserQuery) (User, error) {
    // Your query logic here
    return User{
        ID:    q.UserID,
        Name:  "John Doe",
        Email: "john@example.com",
    }, nil
}

func main() {
    // Create query bus
    bus := query.NewBus()
    
    // Register handler
    bus.Register(&GetUserHandler{})
    
    // Execute query
    ctx := context.Background()
    q := GetUserQuery{UserID: 1}
    result, err := bus.Execute(ctx, q)
    if err != nil {
        panic(err)
    }
    
    // Type assert the result
    if user, ok := result.(User); ok {
        // Use the user data
        _ = user
    }
}
```

## Middleware

The library supports middleware for both commands and queries, allowing you to add cross-cutting concerns:

### Available Middleware

- **Logging**: Log command/query execution with timing information
- **Validation**: Validate commands/queries before execution
- **Recovery**: Recover from panics during handler execution
- **Timeout**: Add timeout support to handlers
- **Caching**: Cache query results (queries only)

### Command Middleware Example

```go
package main

import (
    "context"
    "go-cqrs/internal/command"
    "go-cqrs/internal/middleware"
)

type Logger struct{}

func (l *Logger) LogCommandStart(ctx context.Context, commandName string) {
    // Log start
}

func (l *Logger) LogCommandSuccess(ctx context.Context, commandName string, duration time.Duration) {
    // Log success
}

func (l *Logger) LogCommandError(ctx context.Context, commandName string, duration time.Duration, err error) {
    // Log error
}

func main() {
    logger := &Logger{}
    
    bus := command.NewBus(
        middleware.LoggingCommandMiddleware(logger),
        middleware.RecoveryCommandMiddleware(),
        middleware.ValidatorCommandMiddleware(),
    )
    
    // Register and execute commands...
}
```

### Query Middleware with Caching

```go
package main

import (
    "context"
    "time"
    "go-cqrs/internal/query"
    "go-cqrs/internal/query/middleware"
    sharedmiddleware "go-cqrs/internal/middleware"
)

func main() {
    cache := middleware.NewMemoryCache()
    keyGenerator := &middleware.DefaultCacheKeyGenerator{}
    
    bus := query.NewBus(
        middleware.LoggingMiddleware(logger),
        sharedmiddleware.RecoveryQueryMiddleware(),
        middleware.CachingMiddleware(cache, keyGenerator, 5*time.Minute),
        middleware.TimeoutMiddleware(30*time.Second),
    )
    
    // Register and execute queries...
}
```

## Architecture

### Commands

Commands represent write operations or actions that change the state of your system. They:

- **Can return results**: Commands may return data such as created entity IDs, confirmation objects, or partial results
- **Handle side effects**: Process business logic, database operations, external API calls, etc.
- Must implement the `ICommand` interface
- Are handled by command handlers implementing `ICommandHandler[T]`
- Return `(interface{}, error)` - the result can be any type including nil

### Queries

Queries represent read operations that retrieve data without side effects. They:

- Return data and potentially errors
- Must implement the `Query` interface  
- Are handled by query handlers implementing `QueryHandler[T, R]`

### Buses

- **Command Bus**: Routes commands to their respective handlers
- **Query Bus**: Routes queries to their respective handlers
- Both support middleware for cross-cutting concerns

## Error Handling

The library provides structured error handling:

```go
import "go-cqrs/internal/error"

// Check for specific error types
if busErr, ok := err.(*buserror.BusError); ok {
    switch busErr.Type {
    case buserror.HandlerNotRegistered:
        // Handle missing handler
    case buserror.ExecutionError:
        // Handle execution error
    case buserror.MiddlewareError:
        // Handle middleware error
    }
}
```

## Examples

The `examples/` directory contains comprehensive examples:

- `command-success/`: Basic command execution
- `command-with-result/`: Command that returns a result (ID, created entity, etc.)
- `command-error/`: Command error handling
- `query-success/`: Basic query execution  
- `query-cache/`: Query caching with middleware

Run any example:

```bash
cd examples/command-success
go run main.go
```

## Best Practices

1. **Keep commands simple**: Commands should represent a single business action
2. **Make queries pure**: Queries should not have side effects
3. **Use middleware wisely**: Add only necessary middleware to avoid performance overhead
4. **Handle errors appropriately**: Use the structured error types for proper error handling
5. **Type safety**: Leverage Go's type system for compile-time guarantees
6. **Context propagation**: Always pass context through the execution chain

## Performance Considerations

- Handlers are resolved using type reflection once during registration
- Middleware is executed in the order specified during bus creation
- Caching middleware can significantly improve query performance for expensive operations
- Use context timeouts to prevent long-running operations

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Roadmap

- Event sourcing support
- Distributed caching backends
- Metrics and monitoring integration
- GraphQL integration
- gRPC transport layer
- Advanced validation rules
- Saga pattern implementation

---

Built with Go and designed for simplicity, performance, and maintainability.
