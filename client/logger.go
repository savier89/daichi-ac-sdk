package client

import (
	"fmt"
	"io"
	"os"
	"sync"
	"time"
)

// LogLevel — уровни логирования
type LogLevel int

const (
	LogNone LogLevel = iota
	LogError
	LogWarn
	LogInfo
	LogDebug
)

// levelToString — преобразует уровень в строку
func levelToString(level LogLevel) string {
	switch level {
	case LogDebug:
		return "DEBUG"
	case LogInfo:
		return "INFO"
	case LogWarn:
		return "WARN"
	case LogError:
		return "ERROR"
	default:
		return ""
	}
}

// Logger — унифицированный логгер
type Logger struct {
	level  LogLevel
	mu     sync.Mutex
	output io.Writer
}

// NewLogger — создает новый логгер
func NewLogger(level LogLevel, output io.Writer) *Logger {
	if output == nil {
		output = os.Stderr
	}
	return &Logger{
		level:  level,
		output: output,
	}
}

// SetLevel — устанавливает уровень логирования
func (l *Logger) SetLevel(level LogLevel) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.level = level
}

// log — универсальная функция логирования
func (l *Logger) log(level LogLevel, format string, args ...interface{}) {
	if level > l.level {
		return
	}

	color := ""
	reset := "\033[0m"
	levelStr := levelToString(level)

	switch level {
	case LogDebug:
		color = "\033[36m" // Cyan
	case LogInfo:
		color = "\033[32m" // Green
	case LogWarn:
		color = "\033[33m" // Yellow
	case LogError:
		color = "\033[31m" // Red
	default:
		return
	}

	timestamp := time.Now().Format("2006/01/02 15:04:05")
	_, _ = fmt.Fprintf(l.output, "%s [%s] %s\n", color+timestamp+reset, levelStr, fmt.Sprintf(format, args...))
}

// Debug — выводит debug-логи
func (l *Logger) Debug(format string, args ...interface{}) {
	l.log(LogDebug, format, args...)
}

// Info — выводит info-логи
func (l *Logger) Info(format string, args ...interface{}) {
	l.log(LogInfo, format, args...)
}

// Warn — выводит warn-логи
func (l *Logger) Warn(format string, args ...interface{}) {
	l.log(LogWarn, format, args...)
}

// Error — выводит error-логи
func (l *Logger) Error(format string, args ...interface{}) {
	l.log(LogError, format, args...)
}
