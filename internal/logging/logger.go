package logging

import (
	"errors"
	"fmt"
	"io"
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

type fakeWSyncer struct {
	io.Writer
}

func (fw *fakeWSyncer) Sync() error {
	return nil
}

func GetZapLoggerWriter(w io.Writer) *zap.Logger {
	return zap.New(zapcore.NewCore(zapcore.NewConsoleEncoder(getEncoderConfig()),
		&fakeWSyncer{Writer: w}, defaultLogLevel))
}

func GetZapLoggerFile(path string) (*zap.Logger, error) {
	return zap.Config{ //nolint:wrapcheck // wrap is redundant here
		Level:       zap.NewAtomicLevelAt(defaultLogLevel),
		Development: false,
		Sampling: &zap.SamplingConfig{
			Initial:    100, //nolint:mnd // After the first 100 log entries with the same level and
			Thereafter: 100, //nolint:mnd // message in the same second zap will log every 100th entry. (quote)
		},
		Encoding:         "json",
		EncoderConfig:    getEncoderConfig(),
		OutputPaths:      []string{path},
		ErrorOutputPaths: []string{path},
	}.Build()
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
