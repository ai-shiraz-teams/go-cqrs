package buserror

import "fmt"

type BusError interface {
	error

	ErrorCode() string
}

type DispatchError struct {
	code    string
	message string
	cause   error
}

func NewDispatchError(code, message string) *DispatchError {
	return &DispatchError{
		code:    code,
		message: message,
	}
}

func NewDispatchErrorWithCause(code, message string, cause error) *DispatchError {
	return &DispatchError{
		code:    code,
		message: message,
		cause:   cause,
	}
}

func (e *DispatchError) Error() string {
	if e.cause != nil {
		return fmt.Sprintf("%s: %v", e.message, e.cause)
	}
	return e.message
}

func (e *DispatchError) ErrorCode() string {
	return e.code
}

func (e *DispatchError) Unwrap() error {
	return e.cause
}

const (
	ErrorCodeHandlerNotFound     = "HANDLER_NOT_FOUND"
	ErrorCodeHandlerRegistration = "HANDLER_REGISTRATION"
	ErrorCodeDispatchFailed      = "DISPATCH_FAILED"
	ErrorCodeInvalidCommand      = "INVALID_COMMAND"
	ErrorCodeInvalidQuery        = "INVALID_QUERY"
)

func NewHandlerNotFoundError(handlerType, name string) *DispatchError {
	return NewDispatchError(
		ErrorCodeHandlerNotFound,
		fmt.Sprintf("%s handler not found for: %s", handlerType, name),
	)
}

func NewHandlerRegistrationError(message string, cause error) *DispatchError {
	return NewDispatchErrorWithCause(ErrorCodeHandlerRegistration, message, cause)
}

func NewDispatchFailedError(message string, cause error) *DispatchError {
	return NewDispatchErrorWithCause(ErrorCodeDispatchFailed, message, cause)
}
