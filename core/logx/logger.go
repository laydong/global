// Package logx
// logx: this is extend package, use https://github.com/uber-go/zap
package logx

import (
	"fmt"
	"github.com/climber-dong/global/core/logx/logger"
	"github.com/climber-dong/global/genv"
	"github.com/climber-dong/global/gtools"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"io"
	"log"
	"os"
	"time"
)

const (
	defaultChildPath    = "logx/%Y-%m-%d.log" // 默认子目录
	defaultRotationSize = 128 * 1024 * 1024   // 默认大小为128M
	defaultRotationTime = 24 * time.Hour      // 默认每天轮转一次

	LevelInfo  = "info"
	LevelWarn  = "warn"
	LevelError = "error"
)

var sugar *zap.Logger

type Config struct {
	appName       string        // 应用名
	appMode       string        // 应用环境
	logType       string        // 日志类型
	logPath       string        // 日志主路径
	childPath     string        // 日志子路径+文件名
	RotationSize  int64         // 单个文件大小
	RotationCount uint          // 可以保留的文件个数
	RotationTime  time.Duration // 日志分割的时间
	MaxAge        time.Duration // 日志最大保留的天数
}

func GetSugar() *zap.Logger {
	if sugar == nil {
		cfg := Config{
			appName:       genv.AppName(),
			appMode:       genv.RunMode(),
			logType:       genv.LogType(),
			logPath:       genv.LogPath(),
			childPath:     defaultChildPath,
			RotationSize:  defaultRotationSize,
			RotationCount: genv.LogMaxCount(),
			RotationTime:  defaultRotationTime,
			MaxAge:        genv.LogMaxAge(),
		}

		sugar = InitSugar(&cfg)
	}

	return sugar
}

func InitSugar(lc *Config) *zap.Logger {
	loglevel := zapcore.InfoLevel
	defaultLogLevel := zap.NewAtomicLevel()
	defaultLogLevel.SetLevel(loglevel)

	logPath := fmt.Sprintf("%s/%s/%s", lc.logPath, lc.appName, lc.childPath)

	var core zapcore.Core
	// 打印至文件中
	if lc.logType == "file" {
		configs := zap.NewProductionEncoderConfig()
		configs.FunctionKey = "func"
		configs.EncodeTime = timeEncoder

		w := zapcore.AddSync(GetWriter(logPath, lc))

		core = zapcore.NewCore(
			zapcore.NewJSONEncoder(configs),
			w,
			defaultLogLevel,
		)
		log.Printf("[app] logger success")
	} else {
		// 打印在控制台
		consoleEncoder := zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig())
		core = zapcore.NewCore(consoleEncoder, zapcore.Lock(os.Stdout), defaultLogLevel)
		log.Printf("[app] logger success")
	}

	filed := zap.Fields(zap.String("app_name", lc.appName), zap.String("app_mode", lc.appMode))
	return zap.New(core, filed, zap.AddCaller(), zap.AddCallerSkip(3))
}

func Info(logId, template string, args ...interface{}) {
	msg, fields := dealWithArgs(template, args...)
	writer(logId, LevelInfo, msg, fields...)
}

func Warn(logId, template string, args ...interface{}) {
	msg, fields := dealWithArgs(template, args...)
	writer(logId, LevelWarn, msg, fields...)
}

func Error(logId, template string, args ...interface{}) {
	msg, fields := dealWithArgs(template, args...)
	writer(logId, LevelError, msg, fields...)
}

func dealWithArgs(tmp string, args ...interface{}) (msg string, f []zap.Field) {
	var tmpArgs []interface{}
	for _, item := range args {
		if zapField, ok := item.(zap.Field); ok {
			f = append(f, zapField)
		} else {
			tmpArgs = append(tmpArgs, item)
		}
	}
	msg = fmt.Sprintf(tmp, tmpArgs...)
	return
}

func writer(logId, level, msg string, fields ...zap.Field) {
	fields = append(fields, zap.String(gtools.RequestIdKey, logId))

	switch level {
	case LevelInfo:
		GetSugar().Info(msg, fields...)
	case LevelWarn:
		GetSugar().Warn(msg, fields...)
	case LevelError:
		GetSugar().Error(msg, fields...)
	}
	return
}

func timeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	var layout = "2006-01-02 15:04:05"
	type appendTimeEncoder interface {
		AppendTimeLayout(time.Time, string)
	}

	if enc, ok := enc.(appendTimeEncoder); ok {
		enc.AppendTimeLayout(t, layout)
		return
	}

	enc.AppendString(t.Format(layout))
}

// GetWriter 按天切割按大小切割
// filename 文件名
// RotationSize 每个文件的大小
// MaxAge 文件最大保留天数
// RotationCount 最大保留文件个数
// RotationTime 设置文件分割时间
// RotationCount 设置保留的最大文件数量
func GetWriter(filename string, lc *Config) io.Writer {
	// 生成rotatelogs的Logger 实际生成的文件名 stream-2021-5-20.logger
	// demo.log是指向最新日志的连接
	// 保存7天内的日志，每1小时(整点)分割一第二天志
	var options []logger.Option
	options = append(options,
		logger.WithRotationSize(lc.RotationSize),
		logger.WithRotationCount(lc.RotationCount),
		logger.WithRotationTime(lc.RotationTime),
		logger.WithMaxAge(lc.MaxAge))

	hook, err := logger.New(
		filename,
		options...,
	)

	if err != nil {
		panic(err)
	}
	return hook
}

type Field = zap.Field

func String(key string, value interface{}) zap.Field {
	v := fmt.Sprintf("%v", value)
	return zap.String(key, v)
}
