package main

import (
	"context"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/zozole1818/grocery-list/internal"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"
)

func main() {

	slog.SetLogLoggerLevel(slog.LevelDebug)

	ctx := context.Background()
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()

	port := 8080

	e := echo.New()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	//e.Use(middleware.Static("public/static")) // uncomment to serve static files
	//e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
	//	AllowOrigins: []string{"http://localhost:5173"},
	//	AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept},
	//})) // uncomment to define CORS

	h := internal.NewHandler(ctx)

	e.GET("/quotes", h.GetQuote())

	go func() {
		if err := e.Start(":" + strconv.Itoa(port)); err != nil && err != http.ErrServerClosed {
			slog.Error("HTTP server Start error: " + err.Error())
		}
	}()

	// Wait for context cancellation
	<-ctx.Done()

	// Graceful shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := e.Shutdown(shutdownCtx); err != nil {
		slog.Error("server shutdown error", "error", err)
	}

	<-shutdownCtx.Done()
	slog.Info("Server shutdown complete.")
}
