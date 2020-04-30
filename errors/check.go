package errors

import "github.com/sirupsen/logrus"

func Check(err error) {
	if err != nil {
		logrus.Fatal(err)
	}
}
