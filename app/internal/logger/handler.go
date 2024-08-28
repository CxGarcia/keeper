package logger

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

const (
	LogLevelDebug = iota
	LogLevelInfo
	LogLevelError
)

type Logger struct {
	level         int
	file          *os.File
	writer        *bufio.Writer
	buffer        []string
	bufferSize    int
	flushInterval time.Duration
	mu            sync.Mutex
}

var (
	logger *Logger
	once   sync.Once
)

func Init(level string, logFilePath string, bufferSize int, flushInterval time.Duration) error {
	var err error
	once.Do(func() {
		logger, err = newLogger(level, logFilePath, bufferSize, flushInterval)
	})
	return err
}

func newLogger(level string, logFilePath string, bufferSize int, flushInterval time.Duration) (*Logger, error) {
	logLevel := stringToLogLevel(level)

	dir := filepath.Dir(logFilePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}

	file, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}

	l := &Logger{
		level:         logLevel,
		file:          file,
		writer:        bufio.NewWriter(file),
		buffer:        make([]string, 0, bufferSize),
		bufferSize:    bufferSize,
		flushInterval: flushInterval,
	}

	go l.flushRoutine()

	return l, nil
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

func levelToString(level int) string {
	switch level {
	case LogLevelDebug:
		return "DEBUG"
	case LogLevelInfo:
		return "INFO"
	case LogLevelError:
		return "ERROR"
	default:
		return "INFO"
	}
}

func levelToColor(level int) string {
	switch level {
	case LogLevelDebug:
		return "\033[36m" // Cyan for Debug
	case LogLevelInfo:
		return "\033[32m" // Green for Info
	case LogLevelError:
		return "\033[31m" // Red for Error
	default:
		return "\033[0m" // Default color
	}
}

func (l *Logger) bufferLog(message string) {
	l.buffer = append(l.buffer, message)
	if len(l.buffer) >= l.bufferSize {
		l.flush()
	}
}

func (l *Logger) flush() {
	if len(l.buffer) == 0 {
		return
	}

	for _, msg := range l.buffer {
		l.writer.WriteString(msg + "\n")
	}
	l.writer.Flush()
	l.buffer = l.buffer[:0]
}

func (l *Logger) flushRoutine() {
	ticker := time.NewTicker(l.flushInterval)
	defer ticker.Stop()

	for range ticker.C {
		l.mu.Lock()
		l.flush()
		l.mu.Unlock()
	}
}

func formatLogMessage(level int, message string) string {
	return fmt.Sprintf(
		"[%s] %s - %s",
		levelToString(level), time.Now().Format("2006-01-02 15:04:05"), message,
	)
}

func (l *Logger) log(level int, v ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()

	message := formatLogMessage(level, fmt.Sprint(v...))
	if level >= l.level {
		fmt.Printf("%s%s\033[0m\n", levelToColor(level), message)
	}

	l.bufferLog(message)
}

func (l *Logger) logf(level int, format string, v ...interface{}) {
	l.log(level, fmt.Sprintf(format, v...))
}

func Debug(v ...interface{}) {
	if logger != nil {
		logger.log(LogLevelDebug, v...)
	}
}

func Debugf(format string, v ...interface{}) {
	if logger != nil {
		logger.logf(LogLevelDebug, format, v...)
	}
}

func Info(v ...interface{}) {
	if logger != nil {
		logger.log(LogLevelInfo, v...)
	}
}

func Infof(format string, v ...interface{}) {
	if logger != nil {
		logger.logf(LogLevelInfo, format, v...)
	}
}

func Error(v ...interface{}) (err error) {
	if logger != nil {
		logger.log(LogLevelError, v...)
	}

	return fmt.Errorf("%v", v...)
}

func Errorf(format string, v ...interface{}) (err error) {
	if logger != nil {
		logger.logf(LogLevelError, format, v...)
	}

	return fmt.Errorf(format, v...)
}

func Fatal(v ...interface{}) {
	Error(v...)
	Close()
	os.Exit(1)
}

func Fatalf(format string, v ...interface{}) {
	Errorf(format, v...)
	Close()
	os.Exit(1)
}

func Close() error {
	if logger != nil {
		logger.mu.Lock()
		defer logger.mu.Unlock()

		logger.flush()
		if err := logger.writer.Flush(); err != nil {
			return err
		}
		return logger.file.Close()
	}
	return nil
}
