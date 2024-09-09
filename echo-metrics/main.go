package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/labstack/echo-contrib/echoprometheus"
	"github.com/labstack/echo/v4"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	e := echo.New()
	e.Use(echoprometheus.NewMiddleware("hello"))
	e.GET("/metrics", echoprometheus.NewHandler())
	e.GET("/hello", func(c echo.Context) error {
		return c.String(http.StatusOK, "hello!")
	})

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		<-ctx.Done()
		tctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := e.Shutdown(tctx); err != nil {
			log.Fatalf("shutting down the server: %v", err)
		}
	}()

	if err := e.Start(":8080"); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("shutting down the server: %v", err)
	}
	wg.Wait()
}
