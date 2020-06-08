package errors

import "fmt"

type ErrorWithCode struct {
	Err  error
	Code int
}

func (e *ErrorWithCode) Error() string {
	return fmt.Sprintf("%s", e.Err)
}
