package routes

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func Setup(e *echo.Echo) {
	e.GET("/health", func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})
	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusForbidden, "forbidden")
	})
}
