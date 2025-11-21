package logging

import "log"

type Logger interface {
	Info(msg string, args ...interface{})
	Error(msg string, args ...interface{})
}

type StdLogger struct{}

func (StdLogger) Info(msg string, args ...interface{})  { log.Printf("[INFO] "+msg, args...) }
func (StdLogger) Error(msg string, args ...interface{}) { log.Printf("[ERROR] "+msg, args...) }
