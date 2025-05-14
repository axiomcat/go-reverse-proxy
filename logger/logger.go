package logger

import (
	"log"
	"sync"
)

type LogLevel int

const (
	Logging LogLevel = iota
	Debug
)

type Logger struct {
	logLevel LogLevel
}

var instance *Logger
var once sync.Once

func GetInstance(logLevel LogLevel) *Logger {
	once.Do(func() {
		instance = &Logger{}
		instance.logLevel = logLevel
	})
	return instance
}

func (l *Logger) UpdateLogLevel(logLevel LogLevel) {
	l.logLevel = logLevel
}

func (l *Logger) Log(message string) {
	if l.logLevel >= Logging {
		log.Println("[LOG]", message)
	}
}

func (l *Logger) Debug(message string) {
	if l.logLevel >= Debug {
		log.Println("[DEBUG]", message)
	}
}

func (l *Logger) Fatal(v ...any) {
	log.Fatalln(v...)
}
