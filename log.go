package logx

import (
	"os"
	"sync"
	"time"

	"github.com/spf13/viper"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	lumberjack "gopkg.in/natefinch/lumberjack.v2"
)

var (
	logInstance *zap.Logger
	logOnce     sync.Once
)

// Logger return zap logger instance
func Logger() *zap.Logger {
	logOnce.Do(func() {
		logInstance = logger()
	})

	return logInstance
}

type Level string

const (
	// DebugLevel logs are typically voluminous, and are usually disabled in
	// production.
	DebugLevel Level = "debug"
	// InfoLevel is the default logging priority.
	InfoLevel Level = "info"
	// WarnLevel logs are more important than Info, but don't need individual
	// human review.
	WarnLevel Level = "warn"
	// ErrorLevel logs are high-priority. If an application is running smoothly,
	// it shouldn't generate any error-level logs.
	ErrorLevel Level = "error"
	// DPanicLevel logs are particularly important errors. In development the
	// logger panics after writing the message.
	DPanicLevel Level = "dpanic"
	// PanicLevel logs a message, then panics.
	PanicLevel Level = "panic"
	// FatalLevel logs a message, then calls os.Exit(1).
	FatalLevel Level = "fatal"
)

func SetLogLevel(level Level) {
	viper.Set("log.console_level", string(level))
	logInstance = logger()
}

func logger() *zap.Logger {
	infoWriter := lumberjack.Logger{
		Filename:   defaultViperString("log.info_file", "log/info.log"), // 日志输出地址
		LocalTime:  defaultViperBool("log.filename_with_time", true),    // 日志文件名时间
		MaxSize:    defaultViperInt("log.file_max_size", 100),           // 每个日志文件保存的最大尺寸 单位：M
		MaxBackups: defaultViperInt("log.file_max_backups", 30),         // 日志文件最多保存多少个备份
		MaxAge:     defaultViperInt("log.file_max_age", 30),             // 文件最多保存多少天
		Compress:   defaultViperBool("log.file_compress", true),         // 是否压缩
	}

	errorWriter := lumberjack.Logger{
		Filename:   defaultViperString(viper.GetString("log.error_file"), "log/error.log"), // 日志输出地址
		LocalTime:  defaultViperBool("log.filename_with_time", true),                       // 日志文件名时间
		MaxSize:    defaultViperInt("log.file_max_size", 100),                              // 每个日志文件保存的最大尺寸 单位：M
		MaxBackups: defaultViperInt("log.file_max_backups", 30),                            // 日志文件最多保存多少个备份
		MaxAge:     defaultViperInt("log.file_max_age", 30),                                // 文件最多保存多少天
		Compress:   defaultViperBool("log.file_compress", true),                            // 是否压缩
	}

	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "",
		MessageKey:     "msg",
		StacktraceKey:  "",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
		EncodeName:     zapcore.FullNameEncoder,
	}

	consoleConfig := zapcore.EncoderConfig{
		LevelKey:       "level",
		TimeKey:        "time",
		NameKey:        "",
		CallerKey:      "",
		MessageKey:     "msg",
		StacktraceKey:  "",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalColorLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.FullCallerEncoder,
		EncodeName:     zapcore.FullNameEncoder,
	}

	atomicLevel := zap.NewAtomicLevel()
	atomicLevel.SetLevel(zap.DebugLevel)

	var l zapcore.Level
	_ = l.UnmarshalText([]byte(defaultViperString("log.console_level", "info")))
	atomicLevel.SetLevel(l)

	core := zapcore.NewTee(
		zapcore.NewCore(zapcore.NewJSONEncoder(encoderConfig), zapcore.AddSync(&infoWriter), zap.InfoLevel),
		zapcore.NewCore(zapcore.NewJSONEncoder(encoderConfig), zapcore.AddSync(&errorWriter), zap.ErrorLevel),
		zapcore.NewCore(zapcore.NewConsoleEncoder(consoleConfig), zapcore.AddSync(os.Stdout), atomicLevel),
	)

	development := zap.Development()

	z := zap.New(core, development)

	z = z.With(zap.Int64("uuid", time.Now().Unix()))

	return z
}
