package http

import (
	"github.com/adwski/vidi/internal/generators"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// GetEchoWithDefaultMiddleware returns configured echo middleware
// to be used together with http servers.
func GetEchoWithDefaultMiddleware() *echo.Echo {
	e := echo.New()

	gen := generators.NewID()

	e.Use(middleware.Recover())

	e.Use(middleware.RequestIDWithConfig(middleware.RequestIDConfig{
		Skipper:      middleware.DefaultSkipper,
		Generator:    gen.GetStringOrPanic,
		TargetHeader: echo.HeaderXRequestID,
	}))

	loggerCfg := middleware.DefaultLoggerConfig // + colorer
	// time just like in zap
	loggerCfg.CustomTimeFormat = "2006-01-02T15:04:05.000Z0700"
	// add level for uniformity
	loggerCfg.Format = `{"level":"info","time":"${time_custom}","request_id":"${id}","remote_ip":"${remote_ip}",` +
		`"host":"${host}","method":"${method}","uri":"${uri}","user_agent":"${user_agent}",` +
		`"status":${status},"error":"${error}","latency":"${latency_human}"` +
		`,"bytes_in":${bytes_in},"bytes_out":${bytes_out}}` + "\n"
	e.Use(middleware.LoggerWithConfig(loggerCfg))
	return e
}
