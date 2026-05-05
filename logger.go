package yahoofinance

import "log"

type Logger interface {
	Info(msg string, args ...interface{})
	Warn(msg string, args ...interface{})
	Error(msg string, args ...interface{})
	Debug(msg string, args ...interface{})
}

type defaultLogger struct {
	logger *log.Logger
}

func newDefaultLogger() *defaultLogger {
	return &defaultLogger{logger: log.Default()}
}

func (l *defaultLogger) Info(msg string, args ...interface{})  { l.logger.Printf("[INFO] "+msg, args...) }
func (l *defaultLogger) Warn(msg string, args ...interface{})  { l.logger.Printf("[WARN] "+msg, args...) }
func (l *defaultLogger) Error(msg string, args ...interface{}) { l.logger.Printf("[ERROR] "+msg, args...) }
func (l *defaultLogger) Debug(msg string, args ...interface{}) { l.logger.Printf("[DEBUG] "+msg, args...) }
