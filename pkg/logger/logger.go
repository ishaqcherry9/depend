package logger

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/natefinch/lumberjack"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	formatConsole = "console"
	formatJSON    = "json"

	levelDebug = "DEBUG"
	levelInfo  = "INFO"
	levelWarn  = "WARN"
	levelError = "ERROR"
)

var defaultLogger *zap.Logger
var defaultSugaredLogger *zap.SugaredLogger

func getLogger() *zap.Logger {
	checkNil()
	return defaultLogger.WithOptions(zap.AddCallerSkip(1))
}

func getSugaredLogger() *zap.SugaredLogger {
	checkNil()
	return defaultSugaredLogger.WithOptions(zap.AddCallerSkip(1))
}

func Init(opts ...Option) (*zap.Logger, error) {
	o := defaultOptions()
	o.apply(opts...)
	isSave := o.isSave
	levelName := o.level
	encoding := o.encoding

	var err error
	var zapLog *zap.Logger
	var str string
	if !isSave {
		zapLog, err = log2Terminal(levelName, encoding)
		if err != nil {
			panic(err)
		}
		str = fmt.Sprintf("initialize logger finish, config is output to 'terminal', format=%s, level=%s", encoding, levelName)
	} else {
		zapLog = log2File(encoding, levelName, o.fileConfig)
		str = fmt.Sprintf("initialize logger finish, config is output to 'file', format=%s, level=%s, file=%s", encoding, levelName, o.fileConfig.filename)
	}

	if len(o.hooks) > 0 {
		zapLog = zapLog.WithOptions(zap.Hooks(o.hooks...))
	}

	defaultLogger = zapLog
	defaultSugaredLogger = defaultLogger.Sugar()
	Info(str)

	return defaultLogger, err
}

func log2Terminal(levelName string, encoding string) (*zap.Logger, error) {
	js := fmt.Sprintf(`{
      		"level": "%s",
            "encoding": "%s",
      		"outputPaths": ["stdout"],
            "errorOutputPaths": ["stdout"]
		}`, levelName, encoding)

	var config zap.Config
	err := json.Unmarshal([]byte(js), &config)
	if err != nil {
		return nil, err
	}

	config.EncoderConfig = zap.NewProductionEncoderConfig()
	if encoding == formatConsole {
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	} else {
		config.EncoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	}
	config.EncoderConfig.EncodeTime = timeFormatter
	return config.Build()
}

func log2File(encoding string, levelName string, fo *fileOptions) *zap.Logger {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	var encoder zapcore.Encoder
	if encoding == formatConsole {
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	} else {
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	}

	ws := zapcore.AddSync(&lumberjack.Logger{
		Filename:   fo.filename,
		MaxSize:    fo.maxSize,
		MaxBackups: fo.maxBackups,
		MaxAge:     fo.maxAge,
		Compress:   fo.isCompression,
	})
	core := zapcore.NewCore(encoder, ws, getLevelSize(levelName))

	return zap.New(core, zap.AddCaller())
}

func getLevelSize(levelName string) zapcore.Level {
	levelName = strings.ToUpper(levelName)
	switch levelName {
	case levelDebug:
		return zapcore.DebugLevel
	case levelInfo:
		return zapcore.InfoLevel
	case levelWarn:
		return zapcore.WarnLevel
	case levelError:
		return zapcore.ErrorLevel
	}
	return zapcore.DebugLevel
}

func timeFormatter(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format("2006-01-02 15:04:05.000"))
}

func GetWithSkip(skip int) *zap.Logger {
	checkNil()
	return defaultLogger.WithOptions(zap.AddCallerSkip(skip))
}

func Get() *zap.Logger {
	checkNil()
	return defaultLogger
}

func checkNil() {
	if defaultLogger == nil {
		_, err := Init()
		if err != nil {
			panic(err)
		}
	}
}
