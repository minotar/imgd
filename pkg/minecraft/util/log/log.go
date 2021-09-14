package log

import (
	"fmt"
	stdlog "log"
)

type Logger interface {
	Debug(args ...interface{})
	Debugf(template string, args ...interface{})

	Info(args ...interface{})
	Infof(template string, args ...interface{})

	Warn(args ...interface{})
	Warnf(template string, args ...interface{})

	Error(args ...interface{})
	Errorf(template string, args ...interface{})

	Panic(args ...interface{})
	Panicf(template string, args ...interface{})

	Fatal(args ...interface{})
	Fatalf(template string, args ...interface{})

	With(args ...interface{}) Logger
}

// StdLogger is a super basic, simply usinglog.Print, and implementes the Logger interface

var _ Logger = new(StdLogger)

type StdLogger struct {
	Logger *stdlog.Logger
	name   string
}

func NewStdLogger() *StdLogger {
	return &StdLogger{
		Logger: stdlog.Default(),
	}
}

func (l *StdLogger) log(level string, message string) {
	l.Logger.Print(level, l.name, message)
}

func (l *StdLogger) Named(name string) {
	l.name = l.name + name
}

// Debug just saves the line
func (l *StdLogger) Debug(args ...interface{}) { l.log("DEBUG ", fmt.Sprint(args...)) }

func (l *StdLogger) Info(args ...interface{}) { l.log("INFO ", fmt.Sprint(args...)) }

func (l *StdLogger) Warn(args ...interface{}) { l.log("WARN ", fmt.Sprint(args...)) }

func (l *StdLogger) Error(args ...interface{}) { l.log("ERROR ", fmt.Sprint(args...)) }

func (l *StdLogger) Panic(args ...interface{}) { l.Logger.Panic("PANIC ", l.name, fmt.Sprint(args...)) }

func (l *StdLogger) Fatal(args ...interface{}) { l.Logger.Fatal("FATAL ", l.name, fmt.Sprint(args...)) }

func (l *StdLogger) Debugf(template string, args ...interface{}) {
	l.Debug(fmt.Sprintf(template, args...))
}

func (l *StdLogger) Infof(template string, args ...interface{}) {
	l.Info(fmt.Sprintf(template, args...))
}

func (l *StdLogger) Warnf(template string, args ...interface{}) {
	l.Warn(fmt.Sprintf(template, args...))
}

func (l *StdLogger) Errorf(template string, args ...interface{}) {
	l.Error(fmt.Sprintf(template, args...))
}

func (l *StdLogger) Panicf(template string, args ...interface{}) {
	l.Panic(fmt.Sprintf(template, args...))
}

func (l *StdLogger) Fatalf(template string, args ...interface{}) {
	l.Fatal(fmt.Sprintf(template, args...))
}

func (l *StdLogger) With(args ...interface{}) Logger {
	newLogger := &StdLogger{}
	newLogger.Named(fmt.Sprintf("%s=%s", args[0], args[1]))
	return newLogger
}
