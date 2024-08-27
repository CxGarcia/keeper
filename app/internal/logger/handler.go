package log

import (
	stdLog "log"
	"strings"
)

var level int

const (
	LogLevelDebug = iota
	LogLevelInfo
	LogLevelError
)

func Init(l string) {
	level = stringToLogLevel(l)
}

func stringToLogLevel(level string) int {
	switch strings.ToLower(level) {
	case "debug":
		return LogLevelDebug
	case "info":
		return LogLevelInfo
	case "error":
		return LogLevelError
	default:
		return LogLevelInfo
	}
}

func Println(v ...interface{}) {
	stdLog.Println(v...)
}

func Debug(v ...interface{}) {
	if level <= LogLevelDebug {
		stdLog.Print(v...)
	}
}

func Debugf(format string, v ...interface{}) {
	if level <= LogLevelDebug {
		stdLog.Printf(format, v...)
	}
}

func Info(v ...interface{}) {
	if level <= LogLevelInfo {
		stdLog.Print(v...)
	}
}

func Infof(format string, v ...interface{}) {
	if level <= LogLevelInfo {
		stdLog.Printf(format, v...)
	}
}

func Error(v ...interface{}) {
	if level <= LogLevelError {
		stdLog.Print(v...)
	}
}

func Errorf(format string, v ...interface{}) {
	if level <= LogLevelError {
		stdLog.Printf(format, v...)
	}
}

func Fatal(v ...interface{}) {
	stdLog.Fatal(v...)
}

func Fatalf(format string, v ...interface{}) {
	stdLog.Fatalf(format, v...)
}
