package middleware

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func GetEchoWithDefaultMiddleware() *echo.Echo {
	e := echo.New()
	e.Use(middleware.RequestID()) // TODO configure ID gen func
	e.Use(middleware.Logger())    // TODO use with zap.Logger
	e.Use(middleware.Recover())

	return e
}
