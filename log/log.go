package log

import (
	"fmt"
	"net/http"

	"github.com/fatih/color"
	"github.com/sirupsen/logrus"
)

const (
	FATAL = 0 // fatal only
	WARN  = 1 // warn + fatal
	INFO  = 2 // all
)

var (
	// IsShellMode determines what to print to.
	// if false, use logrus. if true, print raw to tty
	IsShellMode = false

	successCol  = color.New(color.FgGreen).SprintFunc()
	infoCol     = color.New(color.FgWhite).SprintFunc()
	warnCol     = color.New(color.FgYellow).SprintFunc()
	errCol      = color.New(color.FgRed).SprintFunc()
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

func Success(format string, args ...interface{}) {
	if IsShellMode {
		s := fmt.Sprintf(successCol(format), args...)
		fmt.Println(s)
		return
	}
	logrus.Info(fmt.Sprintf(format, args...))
}

func Prompt(p string) {
	fmt.Print(infoCol(p))
}

func Info(format string, args ...interface{}) {
	if IsShellMode {
		s := fmt.Sprintf(infoCol(format), args...)
		fmt.Println(s)
		return
	}
	logrus.Info(fmt.Sprintf(format, args...))
}

func WInfo(w http.ResponseWriter, format string, args ...interface{}) {
	fmt.Fprintf(w, format, args...)
	Info(format, args...)
}

func Warn(format string, args ...interface{}) {
	if IsShellMode {
		s := fmt.Sprintf(warnCol(format), args...)
		fmt.Println(s)
		return
	}
	logrus.Warnf(format, args...)
}

func WWarn(w http.ResponseWriter, format string, args ...interface{}) {
	fmt.Fprintf(w, format, args...)
	Warn(format, args...)
}

func Fatal(err error) {
	if IsShellMode {
		s := fmt.Sprintf(errCol("fatal: %s"), err.Error())
		fmt.Println(s)
		panic(err)
	}
	logrus.Fatal(err)
}
