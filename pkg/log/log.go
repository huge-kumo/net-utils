package log

import (
	"fmt"
	"go.uber.org/zap"
	"os"
)

var log *zap.Logger

func init() {
	cfg := zap.Config{
		Level:            zap.NewAtomicLevelAt(zap.DebugLevel),
		Development:      true,
		Encoding:         "console",
		EncoderConfig:    zap.NewDevelopmentEncoderConfig(),
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
	}
	var err error
	log, err = cfg.Build()

	if err != nil {
		fmt.Println("日志模块初始化失败 " + err.Error())
		os.Exit(0)
	}

}

func GetInstance() *zap.SugaredLogger {
	return log.Sugar()
}
