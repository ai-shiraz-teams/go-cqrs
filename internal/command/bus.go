package command

import (
	"context"
	"reflect"
	"sync"

	buserror "github.com/ai-shiraz-teams/go-cqrs/internal/error"
)

type Bus interface {
	Dispatch(ctx context.Context, cmd ICommand) (interface{}, error)
	Register(handler interface{}) error
	RegisterFunc(commandName string, fn func(ctx context.Context, cmd ICommand) (interface{}, error)) error
	IsRegistered(commandName string) bool
	GetRegisteredCommands() []string
}

type DefaultBus struct {
	registry        *handlerRegistry
	middlewareChain *MiddlewareChain
}

func NewBus(middlewares ...Middleware) *DefaultBus {
	return &DefaultBus{
		registry:        newHandlerRegistry(),
		middlewareChain: NewMiddlewareChain(middlewares...),
	}
}

func (b *DefaultBus) Dispatch(ctx context.Context, cmd ICommand) (interface{}, error) {
	if cmd == nil {
		return nil, buserror.NewDispatchError(
			buserror.ErrorCodeInvalidCommand,
			"command cannot be nil",
		)
	}

	commandName := cmd.CommandName()
	if commandName == "" {
		return nil, buserror.NewDispatchError(
			buserror.ErrorCodeInvalidCommand,
			"command name cannot be empty",
		)
	}

	handler, exists := b.registry.getHandler(commandName)
	if !exists {
		return nil, buserror.NewHandlerNotFoundError("command", commandName)
	}

	finalHandler := b.middlewareChain.Execute(handler)
	return finalHandler(ctx, cmd)
}

func (b *DefaultBus) Register(handler interface{}) error {
	if handler == nil {
		return buserror.NewHandlerRegistrationError(
			"handler cannot be nil",
			nil,
		)
	}

	return b.registry.register(handler)
}

func (b *DefaultBus) RegisterFunc(commandName string, fn func(ctx context.Context, cmd ICommand) (interface{}, error)) error {
	if commandName == "" {
		return buserror.NewHandlerRegistrationError(
			"command name cannot be empty",
			nil,
		)
	}

	if fn == nil {
		return buserror.NewHandlerRegistrationError(
			"handler function cannot be nil",
			nil,
		)
	}

	return b.registry.registerFunc(commandName, fn)
}

func (b *DefaultBus) IsRegistered(commandName string) bool {
	return b.registry.isRegistered(commandName)
}

func (b *DefaultBus) GetRegisteredCommands() []string {
	return b.registry.getRegisteredCommands()
}

type handlerRegistry struct {
	mu       sync.RWMutex
	handlers map[string]ExecutorFunc
}

func newHandlerRegistry() *handlerRegistry {
	return &handlerRegistry{
		handlers: make(map[string]ExecutorFunc),
	}
}

func (r *handlerRegistry) register(handler interface{}) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	handlerType := reflect.TypeOf(handler)
	if handlerType == nil {
		return buserror.NewHandlerRegistrationError(
			"handler type cannot be nil",
			nil,
		)
	}

	if handlerType.Kind() != reflect.Ptr || handlerType.Elem().Kind() != reflect.Struct {
		return buserror.NewHandlerRegistrationError(
			"handler must be a pointer to a struct",
			nil,
		)
	}

	handleMethod, exists := handlerType.MethodByName("Handle")
	if !exists {
		return buserror.NewHandlerRegistrationError(
			"handler must implement Handle method",
			nil,
		)
	}

	if err := r.validateHandleMethod(handleMethod.Type); err != nil {
		return err
	}

	commandType := handleMethod.Type.In(2)
	commandName, err := r.extractCommandName(commandType)
	if err != nil {
		return err
	}

	if _, exists := r.handlers[commandName]; exists {
		return buserror.NewHandlerRegistrationError(
			"handler already registered for command: "+commandName,
			nil,
		)
	}

	handlerValue := reflect.ValueOf(handler)
	r.handlers[commandName] = func(ctx context.Context, cmd ICommand) (interface{}, error) {

		if !reflect.TypeOf(cmd).AssignableTo(commandType) {
			return nil, buserror.NewDispatchError(
				buserror.ErrorCodeInvalidCommand,
				"command type mismatch",
			)
		}

		results := handlerValue.MethodByName("Handle").Call([]reflect.Value{
			reflect.ValueOf(ctx),
			reflect.ValueOf(cmd),
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

func (r *handlerRegistry) registerFunc(commandName string, fn func(ctx context.Context, cmd ICommand) (interface{}, error)) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.handlers[commandName]; exists {
		return buserror.NewHandlerRegistrationError(
			"handler already registered for command: "+commandName,
			nil,
		)
	}

	r.handlers[commandName] = fn
	return nil
}

func (r *handlerRegistry) getHandler(commandName string) (ExecutorFunc, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	handler, exists := r.handlers[commandName]
	return handler, exists
}

func (r *handlerRegistry) isRegistered(commandName string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	_, exists := r.handlers[commandName]
	return exists
}

func (r *handlerRegistry) getRegisteredCommands() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	commands := make([]string, 0, len(r.handlers))
	for commandName := range r.handlers {
		commands = append(commands, commandName)
	}
	return commands
}

func (r *handlerRegistry) validateHandleMethod(methodType reflect.Type) error {

	if methodType.NumIn() != 3 || methodType.NumOut() != 2 {
		return buserror.NewHandlerRegistrationError(
			"Handle method must have signature: Handle(context.Context, CommandType) (interface{}, error)",
			nil,
		)
	}

	contextType := reflect.TypeOf((*context.Context)(nil)).Elem()
	if !methodType.In(1).Implements(contextType) {
		return buserror.NewHandlerRegistrationError(
			"first parameter must be context.Context",
			nil,
		)
	}

	commandInterface := reflect.TypeOf((*ICommand)(nil)).Elem()
	if !methodType.In(2).Implements(commandInterface) {
		return buserror.NewHandlerRegistrationError(
			"second parameter must implement ICommand",
			nil,
		)
	}

	errorInterface := reflect.TypeOf((*error)(nil)).Elem()
	if !methodType.Out(1).Implements(errorInterface) {
		return buserror.NewHandlerRegistrationError(
			"second return type must be error",
			nil,
		)
	}

	return nil
}

func (r *handlerRegistry) extractCommandName(commandType reflect.Type) (string, error) {

	commandValue := reflect.New(commandType).Elem()

	cmdInterface := commandValue.Interface().(ICommand)
	commandName := cmdInterface.CommandName()

	if commandName == "" {
		return "", buserror.NewHandlerRegistrationError(
			"command name cannot be empty",
			nil,
		)
	}

	return commandName, nil
}
