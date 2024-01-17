package logging

import (
	"errors"
	"fmt"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	defaultLogLevel = zapcore.DebugLevel
)

func GetZapLoggerWithLevel(cfgLogLvl string) (*zap.Logger, error) {
	var (
		logLvl = defaultLogLevel
		err    error
	)
	if cfgLogLvl == "" {
		err = errors.New("log level cannot be empty")
	} else {
		if parsedLogLvl, errP := zapcore.ParseLevel(cfgLogLvl); errP != nil {
			err = fmt.Errorf("cannot parse log level: %w", errP)
		} else {
			logLvl = parsedLogLvl
		}
	}
	logger := zap.New(zapcore.NewCore(zapcore.NewJSONEncoder(getEncoderConfig()), os.Stdout, logLvl))
	return logger, err
}

func GetZapLoggerDefaultLevel() *zap.Logger {
	return zap.New(zapcore.NewCore(zapcore.NewJSONEncoder(getEncoderConfig()), os.Stdout, defaultLogLevel))
}

func GetZapLoggerConsole() *zap.Logger {
	return zap.New(zapcore.NewCore(zapcore.NewConsoleEncoder(getEncoderConfig()), os.Stdout, defaultLogLevel))
}

func getEncoderConfig() zapcore.EncoderConfig {
	return zapcore.EncoderConfig{
		MessageKey:     "msg",
		LevelKey:       "level",
		TimeKey:        "time",
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
	}
}
