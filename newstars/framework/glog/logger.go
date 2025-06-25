package glog

import (
	"fmt"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"newstars/framework/config"
	"os"
	"sync"
)

var (
	logger      *zap.Logger
	sugarLogger *zap.SugaredLogger
	once        sync.Once
)

func Init(config *config.Zap) *zap.Logger {
	once.Do(func() {
		logger = newLogger(config)
		sugarLogger = logger.Sugar()
		zap.ReplaceGlobals(logger)
	})
	return logger
}

// Zap 获取 zap.Logger
// Author [SliverHorn](https://github.com/SliverHorn)
func newLogger(config *config.Zap) (logger *zap.Logger) {
	if ok, _ := PathExists(config.Director); !ok { // 判断是否有Director文件夹
		fmt.Printf("create %v directory\n", config.Director)
		_ = os.Mkdir(config.Director, os.ModePerm)
	}

	cores := Zap.GetZapCores()
	logger = zap.New(zapcore.NewTee(cores...))

	logger = logger.WithOptions(zap.AddCaller(), zap.AddCallerSkip(1))

	return logger
}
