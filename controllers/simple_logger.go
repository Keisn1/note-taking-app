package controllers

import (
	"log"
)

type SimpleLogger struct {
	logger *log.Logger
}

func (l *SimpleLogger) Infof(format string, a ...any) {
	l.logger.Printf(format, a...)
}

func NewSimpleLogger() SimpleLogger {
	return SimpleLogger{logger: log.Default()}
}