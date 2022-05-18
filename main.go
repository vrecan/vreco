package main

import (
	"context"
	"fmt"
	"net/http"
	SYS "syscall"
	"time"

	"vreco/routes"

	DEATH "github.com/vrecan/death"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
)

func main() {
	death := DEATH.NewDeath(SYS.SIGINT, SYS.SIGTERM)

	e := echo.New()
	e.Logger.SetLevel(log.INFO)
	e.Use(middleware.StaticWithConfig(middleware.StaticConfig{
		Root:   "static",
		Browse: false,
	}))
	e.HTTPErrorHandler = customHTTPErrorHandler

	routes.Setup(e)

	go func() {
		err := e.Start(":8080")
		if err != nil {
			e.Logger.Warn(err)
		}
	}()

	death.WaitForDeathWithFunc(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()
		if err := e.Shutdown(ctx); err != nil {
			e.Logger.Fatal(err)
		}
	})
}

func customHTTPErrorHandler(err error, c echo.Context) {
	code := http.StatusInternalServerError
	if he, ok := err.(*echo.HTTPError); ok {
		code = he.Code
	}
	c.Logger().Error(err)
	errorPage := fmt.Sprintf("%d.html", code)
	if err := c.File(errorPage); err != nil {
		c.Logger().Error(err)
	}
}
