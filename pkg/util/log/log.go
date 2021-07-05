package log

import "fmt"

type Logger interface {
	//Named(name string)
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
}

// DummyLogger is a super basic fmt.Println based and implementes the Logger interface

var _ Logger = new(DummyLogger)

type DummyLogger struct {
	name  string
	Lines []string
}

func (l *DummyLogger) saveLine(level string, message string) {
	line := fmt.Sprint(level, l.name, message)
	l.Lines = append(l.Lines, line)
}

func (l *DummyLogger) log(level string, message string) {
	l.saveLine(level, message)
	fmt.Println(level, l.name, message)
}

func (l *DummyLogger) Named(name string) {
	l.name = l.name + name
}

// Debug just saves the line
func (l *DummyLogger) Debug(args ...interface{}) { l.saveLine("DEBUG ", fmt.Sprint(args...)) }

func (l *DummyLogger) Info(args ...interface{}) { l.log("INFO ", fmt.Sprint(args...)) }

func (l *DummyLogger) Warn(args ...interface{}) { l.log("WARN ", fmt.Sprint(args...)) }

func (l *DummyLogger) Error(args ...interface{}) { l.log("ERROR ", fmt.Sprint(args...)) }

func (l *DummyLogger) Panic(args ...interface{}) { l.log("PANIC ", fmt.Sprint(args...)) }

func (l *DummyLogger) Fatal(args ...interface{}) { l.log("FATAL ", fmt.Sprint(args...)) }

func (l *DummyLogger) Debugf(template string, args ...interface{}) {
	l.Debug(fmt.Sprintf(template, args...))
}

func (l *DummyLogger) Infof(template string, args ...interface{}) {
	l.Info(fmt.Sprintf(template, args...))
}

func (l *DummyLogger) Warnf(template string, args ...interface{}) {
	l.Warn(fmt.Sprintf(template, args...))
}

func (l *DummyLogger) Errorf(template string, args ...interface{}) {
	l.Error(fmt.Sprintf(template, args...))
}

func (l *DummyLogger) Panicf(template string, args ...interface{}) {
	l.Panic(fmt.Sprintf(template, args...))
}

func (l *DummyLogger) Fatalf(template string, args ...interface{}) {
	l.Fatal(fmt.Sprintf(template, args...))
}
