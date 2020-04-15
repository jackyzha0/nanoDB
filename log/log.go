package log

import (
	"fmt"
	"net/http"

	"github.com/sirupsen/logrus"
)

const (
	FATAL = 0 // fatal only
	WARN  = 1 // warn + fatal
	INFO  = 2 // all
)

func SetLoggingLevel(l int) {
	switch l {
	case FATAL:
		logrus.SetLevel(logrus.FatalLevel)
	case WARN:
		logrus.SetLevel(logrus.WarnLevel)
	case INFO:
		logrus.SetLevel(logrus.InfoLevel)
	}
}

func Info(format string, args ...interface{}) {
	logrus.Info(fmt.Sprintf(format, args...))
}

func WInfo(w http.ResponseWriter, format string, args ...interface{}) {
	fmt.Fprintf(w, format, args...)
	Info(format, args...)
}

func Warn(format string, args ...interface{}) {
	logrus.Warnf(format, args...)
}

func WWarn(w http.ResponseWriter, format string, args ...interface{}) {
	fmt.Fprintf(w, format, args...)
	Warn(format, args...)
}

func Fatal(err error) {
	logrus.Fatal(err)
}
