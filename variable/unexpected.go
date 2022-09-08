package variable

import (
	"github.com/devthejo/snip/errors"
	"github.com/sirupsen/logrus"
)

func UnexpectedTypeVarValue(k string, v interface{}) {
	logrus.Fatalf("Unexpected var type %T value %v for key %v", v, v, k)
}
func UnexpectedTypeVar(m map[string]interface{}, key string) {
	errors.UnexpectedType(m, key, "var")
}
