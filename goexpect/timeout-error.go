package expect

import (
	"fmt"
	"time"
)

// TimeoutError is the error returned by all Expect functions upon timer expiry.
type TimeoutError int

// Error implements the Error interface.
func (t TimeoutError) Error() string {
	return fmt.Sprintf("expect: timer expired after %d seconds", time.Duration(t)/time.Second)
}
