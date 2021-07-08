package logger

import (
	"errors"
	"fmt"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
	"strings"
	"sync"
	"time"
)

var once sync.Once

type AppLogger struct {
	zLogger *zap.Logger
}

var instantiated *AppLogger

type CallbackFunction func(zapcore.Level, *map[string]interface{})

var externalLogger CallbackFunction
var externalLoggerName string
var level zapcore.Level

func initLogger() {
	once.Do(func() {
		level := zap.InfoLevel
		atom := zap.NewAtomicLevelAt(level)

		// zap logger
		encoderConfig := zapcore.EncoderConfig{
			TimeKey:        "time",
			LevelKey:       "level",
			NameKey:        "logger",
			MessageKey:     "msg",
			StacktraceKey:  "stacktrace",
			LineEnding:     zapcore.DefaultLineEnding,
			EncodeLevel:    zapcore.LowercaseColorLevelEncoder, // 小写编码器
			EncodeTime:     zapcore.ISO8601TimeEncoder,         // ISO8601 UTC 时间格式
			EncodeDuration: zapcore.SecondsDurationEncoder,
			EncodeCaller:   zapcore.FullCallerEncoder, // 全路径编码器
		}

		zconfig := zap.Config{
			Level:         atom,          // 日志级别
			Development:   true,          // 开发模式，堆栈跟踪
			Encoding:      "console",     // 输出格式 console 或 json
			EncoderConfig: encoderConfig, // 编码器配置
			//InitialFields:    map[string]interface{}{"serviceName": "JAVDB"}, // 初始化字段，如：添加一个服务器名称
			OutputPaths:      []string{"stdout"}, // 输出到指定文件 stdout（标准输出，正常颜色） stderr（错误输出，红色）
			ErrorOutputPaths: []string{"stderr"},
		}

		zlogger, err := zconfig.Build()
		if err != nil {
			fmt.Println("logger init error: ", err)
			os.Exit(1)
		}
		zap.ReplaceGlobals(zlogger)

		logger := AppLogger{
			zLogger: zlogger,
		}
		instantiated = &logger
	})
}

// Info is a convenient alias for Root().Info
func Info(msg string, ctx ...zap.Field) {
	initLogger()
	zap.L().Info(msg, ctx...)
	//instantiated.zLogger.Info(msg, ctx...)
	callExternalLogger(zap.InfoLevel, msg, ctx...)
}

// Error is a convenient alias for Root().Error
func Warn(msg string, ctx ...zap.Field) {
	initLogger()
	instantiated.zLogger.Warn(msg, ctx...)
	callExternalLogger(zap.ErrorLevel, msg, ctx...)
}

// Error is a convenient alias for Root().Error
func Error(msg string, ctx ...zap.Field) {
	initLogger()
	instantiated.zLogger.Error(msg, ctx...)
	callExternalLogger(zap.ErrorLevel, msg, ctx...)
}

func callExternalLogger(level zapcore.Level, msg string, fields ...zap.Field) {
	data := make(map[string]interface{})
	if externalLogger != nil {
		for _, value := range fields {
			switch value.Type {
			case zapcore.ErrorType:
				data[value.Key] = value.Interface.(error).Error()
			case zapcore.StringType:
				data[value.Key] = value.Interface.(string)
			}
		}
		data["msg"] = msg
		data["level"] = level.String()
		data["time"] = time.Now().Format(time.RFC3339)
		data["@timestamp"] = time.Now().UTC().UnixNano() / 1000000
		externalLogger(zap.FatalLevel, &data)
	}
}

func RegistExternalCallback(name string, callback CallbackFunction) (err error) {
	if externalLogger != nil && !strings.EqualFold(externalLoggerName, name) {
		return errors.New("function already registed: " + name)
	}
	externalLogger = callback
	externalLoggerName = name
	return nil
}
