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
	encoder := zapcore.NewJSONEncoder(zapcore.EncoderConfig{
		MessageKey:     "msg",
		LevelKey:       "level",
		TimeKey:        "time",
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
	})

	var (
		logLvl zapcore.Level
		err    error
	)
	if cfgLogLvl == "" {
		err = errors.New("log level cannot be empty")
	} else {
		if logLvl, err = zapcore.ParseLevel(cfgLogLvl); err != nil {
			err = fmt.Errorf("cannot parse log level: %w", err)
		} else {
			logLvl = defaultLogLevel
		}
	}
	logger := zap.New(zapcore.NewCore(encoder, os.Stdout, logLvl))
	return logger, err
}

func GetZapLoggerDefaultLevel() *zap.Logger {
	encoder := zapcore.NewJSONEncoder(zapcore.EncoderConfig{
		MessageKey:     "msg",
		LevelKey:       "level",
		TimeKey:        "time",
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
	})
	return zap.New(zapcore.NewCore(encoder, os.Stdout, defaultLogLevel))
}
