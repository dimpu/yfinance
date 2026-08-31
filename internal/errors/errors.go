package errors

import "fmt"

type BadRequestError struct {
	Message string
}

func (e *BadRequestError) Error() string {
	return fmt.Sprintf("bad request: %s", e.Message)
}

type HTTPError struct {
	StatusCode int
	Body       string
}

func (e *HTTPError) Error() string {
	return fmt.Sprintf("HTTP %d: %s", e.StatusCode, e.Body)
}

type InvalidOptionsError struct {
	Field string
	Msg   string
}

func (e *InvalidOptionsError) Error() string {
	return fmt.Sprintf("invalid option %s: %s", e.Field, e.Msg)
}

type FailedValidationError struct {
	Result interface{}
	Errors []string
}

func (e *FailedValidationError) Error() string {
	return fmt.Sprintf("validation failed: %v", e.Errors)
}
