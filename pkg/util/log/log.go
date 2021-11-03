package log

import (
	"fmt"
	std_log "log"

	"go.uber.org/zap"
)

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

	With(args ...interface{}) Logger
}

// BuiltinLogger is a super basic wrapper around the std_log.Logger and implementes the Logger interface

var _ Logger = new(BuiltinLogger)

type BuiltinLogger struct {
	*std_log.Logger
	Level int
	Null  bool
}

func NewBuiltinLogger(level int) *BuiltinLogger {
	return &BuiltinLogger{Logger: std_log.Default(), Level: level}
}

func (bl *BuiltinLogger) log(level int, tag string, args ...interface{}) {
	if level < bl.Level {
		return
	}
	bl.Print(tag, fmt.Sprint(args...))
}

func (bl *BuiltinLogger) Named(name string) {
	bl.SetPrefix(name)
}

func (bl *BuiltinLogger) With(args ...interface{}) Logger {
	newLogger := NewBuiltinLogger(bl.Level)
	newLogger.Named(fmt.Sprintf("%s=%s", args[0], args[1]))
	return newLogger
}

func (bl *BuiltinLogger) Debug(args ...interface{}) { bl.log(0, "DEBUG ", args...) }
func (bl *BuiltinLogger) Debugf(template string, args ...interface{}) {
	bl.Debug(fmt.Sprintf(template, args...))
}

func (bl *BuiltinLogger) Info(args ...interface{}) { bl.log(1, "INFO ", args...) }
func (bl *BuiltinLogger) Infof(template string, args ...interface{}) {
	bl.Info(fmt.Sprintf(template, args...))
}

func (bl *BuiltinLogger) Warn(args ...interface{}) { bl.log(2, "WARN ", args...) }
func (bl *BuiltinLogger) Warnf(template string, args ...interface{}) {
	bl.Warn(fmt.Sprintf(template, args...))
}

func (bl *BuiltinLogger) Error(args ...interface{}) { bl.log(3, "ERROR ", args...) }
func (bl *BuiltinLogger) Errorf(template string, args ...interface{}) {
	bl.Error(fmt.Sprintf(template, args...))
}

var _ Logger = new(ZapLogger)

type ZapLogger struct {
	Sugar *zap.SugaredLogger
}

func NewZapLogger(_logger *zap.Logger) *ZapLogger {
	return &ZapLogger{
		Sugar: _logger.WithOptions(zap.AddCallerSkip(1)).Sugar(),
	}
}

func (zl *ZapLogger) Debug(args ...interface{}) { zl.Sugar.Debug(args...) }
func (zl *ZapLogger) Info(args ...interface{})  { zl.Sugar.Info(args...) }
func (zl *ZapLogger) Warn(args ...interface{})  { zl.Sugar.Warn(args...) }
func (zl *ZapLogger) Error(args ...interface{}) { zl.Sugar.Error(args...) }
func (zl *ZapLogger) Panic(args ...interface{}) { zl.Sugar.Panic(args...) }
func (zl *ZapLogger) Fatal(args ...interface{}) { zl.Sugar.Fatal(args...) }

func (zl *ZapLogger) Debugf(template string, args ...interface{}) { zl.Sugar.Debugf(template, args...) }
func (zl *ZapLogger) Infof(template string, args ...interface{})  { zl.Sugar.Infof(template, args...) }
func (zl *ZapLogger) Warnf(template string, args ...interface{})  { zl.Sugar.Warnf(template, args...) }
func (zl *ZapLogger) Errorf(template string, args ...interface{}) { zl.Sugar.Errorf(template, args...) }
func (zl *ZapLogger) Panicf(template string, args ...interface{}) { zl.Sugar.Panicf(template, args...) }
func (zl *ZapLogger) Fatalf(template string, args ...interface{}) { zl.Sugar.Fatalf(template, args...) }

func (zl *ZapLogger) With(args ...interface{}) Logger {
	newSugar := zl.Sugar.With(args...)
	return &ZapLogger{newSugar}
}
