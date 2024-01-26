package logging

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zapcore"
)

func TestGetZapLoggerConsole(t *testing.T) {
	l := GetZapLoggerConsole()
	l.Sugar().Desugar()
}

func TestGetZapLoggerDefaultLevel(t *testing.T) {
	l := GetZapLoggerDefaultLevel()
	l.Sugar().Desugar()

	assert.Equal(t, defaultLogLevel, l.Level())
}

func TestGetZapLoggerWithLevelEmpty(t *testing.T) {
	_, err := GetZapLoggerWithLevel("")
	assert.ErrorContains(t, err, "log level cannot be empty")
}

func TestGetZapLoggerWithLevelIncorrect(t *testing.T) {
	_, err := GetZapLoggerWithLevel("qweqweqw")
	assert.ErrorContains(t, err, "cannot parse log level")
}

func TestGetZapLoggerWithLevel(t *testing.T) {
	l, err := GetZapLoggerWithLevel("error")
	assert.NoError(t, err)
	assert.Equal(t, zapcore.ErrorLevel, l.Level())
}
