package utils

import (
	"log"
	"os"
)

// Logger 是一个简单的日志包装器
type Logger struct {
	infoLogger    *log.Logger
	warningLogger *log.Logger
	errorLogger   *log.Logger
	debugLogger   *log.Logger
}

// NewLogger 创建一个新的日志记录器
func NewLogger() *Logger {
	return &Logger{
		infoLogger:    log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile),
		warningLogger: log.New(os.Stdout, "WARNING: ", log.Ldate|log.Ltime|log.Lshortfile),
		errorLogger:   log.New(os.Stderr, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile),
		debugLogger:   log.New(os.Stdout, "DEBUG: ", log.Ldate|log.Ltime|log.Lshortfile),
	}
}

// Info 记录信息级别日志
func (l *Logger) Info(v ...interface{}) {
	l.infoLogger.Println(v...)
}

// Infof 记录格式化信息级别日志
func (l *Logger) Infof(format string, v ...interface{}) {
	l.infoLogger.Printf(format, v...)
}

// Warning 记录警告级别日志
func (l *Logger) Warning(v ...interface{}) {
	l.warningLogger.Println(v...)
}

// Warningf 记录格式化警告级别日志
func (l *Logger) Warningf(format string, v ...interface{}) {
	l.warningLogger.Printf(format, v...)
}

// Error 记录错误级别日志
func (l *Logger) Error(v ...interface{}) {
	l.errorLogger.Println(v...)
}

// Debug 记录调试级别日志
func (l *Logger) Debug(v ...interface{}) {
	l.debugLogger.Println(v...)
}

// Debugf 记录格式化调试级别日志
func (l *Logger) Debugf(format string, v ...interface{}) {
	l.debugLogger.Printf(format, v...)
}

// Errorf 记录格式化错误级别日志
func (l *Logger) Errorf(format string, v ...interface{}) {
	l.errorLogger.Printf(format, v...)
}

// GetStandardLogger returns a standard log.Logger for compatibility
func (l *Logger) GetStandardLogger() *log.Logger {
	return l.infoLogger
}
