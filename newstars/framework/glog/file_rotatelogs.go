package glog

import (
	"go.uber.org/zap/zapcore"
	"newstars/framework/config"
	"os"
)

var FileRotatelogs = new(fileRotatelogs)

type fileRotatelogs struct{}

// GetWriteSyncer 获取 zapcore.WriteSyncer
// Author [SliverHorn](https://github.com/SliverHorn)
func (r *fileRotatelogs) GetWriteSyncer(level string) zapcore.WriteSyncer {
	fileWriter := NewCutter(config.GVA_CONFIG.Zap.Director, level, WithCutterFormat("2006-01-02"))

	//feishuEerrList := []string{
	//	zapcore.ErrorLevel.String(), zapcore.WarnLevel.String(), zapcore.DPanicLevel.String(), zapcore.PanicLevel.String(), zapcore.FatalLevel.String(),
	//}
	//
	//if slices.Contains(feishuEerrList, level) == true {
	//	feishuWriter := feishu.NewWriter()
	//	if config.GVA_CONFIG.Zap.LogInConsole {
	//		return zapcore.NewMultiWriteSyncer(zapcore.AddSync(os.Stdout), zapcore.AddSync(fileWriter), zapcore.AddSync(feishuWriter))
	//	}
	//
	//	return zapcore.NewMultiWriteSyncer(zapcore.AddSync(fileWriter), zapcore.AddSync(feishuWriter))
	//}

	if config.GVA_CONFIG.Zap.LogInConsole {
		return zapcore.NewMultiWriteSyncer(zapcore.AddSync(os.Stdout), zapcore.AddSync(fileWriter))
	}
	return zapcore.AddSync(fileWriter)
}
