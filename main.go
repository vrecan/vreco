package main

import (
	"context"
	SYS "syscall"
	"time"

	"vreco/routes"

	DEATH "github.com/vrecan/death"

	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
)

func main() {
	death := DEATH.NewDeath(SYS.SIGINT, SYS.SIGTERM)

	e := echo.New()
	e.Logger.SetLevel(log.INFO)
	routes.Setup(e)

	go func() {
		err := e.Start(":8080")
		if err != nil {
			e.Logger.Warn(err)
		}
	}()

	death.WaitForDeathWithFunc(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

		defer cancel()
		if err := e.Shutdown(ctx); err != nil {
			e.Logger.Fatal(err)
		}
	})
}
