package errors

import (
	"fmt"
	"regexp"
	"strconv"
)

var ErrorStatusRegexp = regexp.MustCompile(`Process exited with status (\d+) *`)

type ErrorWithCode struct {
	Err  error
	Code int
}

func CreateErrorWithCodeFromError(err error) (bool, error) {
	msg := err.Error()
	if ErrorStatusRegexp.MatchString(msg) {
		submatch := ErrorStatusRegexp.FindStringSubmatch(msg)
		code := submatch[1]
		if codeInt, errConv := strconv.Atoi(code); errConv == nil {
			return true, &ErrorWithCode{
				Err:  err,
				Code: codeInt,
			}
		}
	}
	return false, err
}

func (e *ErrorWithCode) Error() string {
	return fmt.Sprintf("%s", e.Err)
}
