package glog

import (
	"go.uber.org/zap"
)

// 原生日志方法
func Debug(msg string, fields ...zap.Field) {
	logger.Debug(msg, fields...)
}

func Info(msg string, fields ...zap.Field) {
	logger.Info(msg, fields...)
}

func Warn(msg string, fields ...zap.Field) {
	logger.Warn(msg, fields...)
}

func Error(msg string, fields ...zap.Field) {
	logger.Error(msg, fields...)
}

func Fatal(msg string, fields ...zap.Field) {
	logger.Fatal(msg, fields...)
}

func Panic(msg string, fields ...zap.Field) {
	logger.Panic(msg, fields...)
}

// Sugared 日志方法（无格式化）
func SDebug(args ...interface{}) {
	sugarLogger.Debug(args...)
}

func SInfo(args ...interface{}) {
	sugarLogger.Info(args...)
}

func SWarn(args ...interface{}) {
	sugarLogger.Warn(args...)
}

func SError(args ...interface{}) {
	sugarLogger.Error(args...)
}

func SFatal(args ...interface{}) {
	sugarLogger.Fatal(args...)
}

func SPanic(args ...interface{}) {
	sugarLogger.Panic(args...)
}

// Sugared 日志方法（格式化）
func SDebugf(template string, args ...interface{}) {
	sugarLogger.Debugf(template, args...)
}

func SInfof(template string, args ...interface{}) {
	sugarLogger.Infof(template, args...)
}

func SWarnf(template string, args ...interface{}) {
	sugarLogger.Warnf(template, args...)
}

func SErrorf(template string, args ...interface{}) {
	sugarLogger.Errorf(template, args...)
}

func SFatalf(template string, args ...interface{}) {
	sugarLogger.Fatalf(template, args...)
}

func SPanicf(template string, args ...interface{}) {
	sugarLogger.Panicf(template, args...)
}
