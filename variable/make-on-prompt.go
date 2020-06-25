package variable

import (
	"strings"
	"sync"

	"github.com/sirupsen/logrus"
)

func MakeOnPromptOnce(msg string, once *sync.Once) func(v *Var) {
	f := MakeOnPrompt(msg)
	return func(v *Var) {
		once.Do(func() {
			f(v)
		})
	}
}
func MakeOnPrompt(msg string) func(v *Var) {
	return func(v *Var) {
		logrus.Info(strings.Repeat("  ", v.Depth+2) + msg)
	}
}
