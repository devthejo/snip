package expect

// Status contains an errormessage and a status code.
import (
	"fmt"

	"google.golang.org/grpc/codes"
)

type Status struct {
	code codes.Code
	msg  string
}

// NewStatus creates a Status with the provided code and message.
func NewStatus(code codes.Code, msg string) *Status {
	return &Status{code, msg}
}

// NewStatusf returns a Status with the provided code and a formatted message.
func NewStatusf(code codes.Code, format string, a ...interface{}) *Status {
	return NewStatus(code, fmt.Sprintf(fmt.Sprintf(format, a...)))
}

// Err is a helper to handle errors.
func (s *Status) Err() error {
	if s == nil || s.code == codes.OK {
		return nil
	}
	return s
}

// Error is here to adhere to the error interface.
func (s *Status) Error() string {
	return s.msg
}
