package errors

import "github.com/sirupsen/logrus"

func UnexpectedType(m map[string]interface{}, key string, typ string) {
	logrus.Fatalf("unexpected "+typ+" "+key+" type %T value %v", m[key], m[key])
}
