package apperror

import "fmt"

type Error struct {
	Code    string
	Message string
	Details map[string]any
}

func (e *Error) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func New(code string, message string, details map[string]any) *Error {
	return &Error{Code: code, Message: message, Details: details}
}
