package main

import (
	"context"
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/zozole1818/grocery-list/internal"
	"gopkg.in/yaml.v3"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"
)

func main() {

	slog.SetLogLoggerLevel(slog.LevelDebug)

	err := loadEnvs(".env")
	if err != nil {
		slog.Error("Error when loading env file", "error", err)
		return
	}

	ctx := context.Background()
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()

	repo := internal.NewMapRepo()

	port := 8080
	e := echo.New()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.Static("public/static"))
	//e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
	//	AllowOrigins: []string{"http://localhost:5173"},
	//	AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept},
	//})) // uncomment to define CORS

	fridge := internal.NewFridge()
	h := internal.NewHandler(ctx, repo, fridge)

	e.GET("/quotes", h.GetQuote())
	e.GET("/items", h.GetItems())
	e.POST("/items", h.AddItem())
	e.PUT("/items", h.UpdateOrCreateItem())
	e.DELETE("/items/:name", h.DeleteItem())
	e.GET("/notifications", h.Notifications())

	e.GET("/fridges/:name", h.GetFridgeItems())
	e.POST("/fridges/:name/fill", h.FillFridge())
	e.POST("/fridges/:name/empty", h.EmptyFridge())

	e.POST("/emails", h.SendEmails())

	go func() {
		if err := e.Start(":" + strconv.Itoa(port)); err != nil && err != http.ErrServerClosed {
			slog.Error("HTTP server Start error: " + err.Error())
		}
	}()

	// Wait for context cancellation

	select {
	case <-ctx.Done():
		err := fridge.Flush()
		if err != nil {
			slog.Error("error when flushing fridge: " + err.Error())
		}
		mapRepo := repo.(*internal.MapRepo)
		err = mapRepo.Flush()
		if err != nil {
			slog.Error("error when flushing shopping list: " + err.Error())
		}

		break
	}
	//<-ctx.Done()

	// Graceful shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := e.Shutdown(shutdownCtx); err != nil {
		slog.Error("server shutdown error", "error", err)
	}

	<-shutdownCtx.Done()
	slog.Info("Server shutdown complete.")
}

func loadEnvs(path string) error {
	b, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("error when reading %s file: %v", path, err)
	}
	result := make(map[string]string)
	err = yaml.Unmarshal(b, &result)
	if err != nil {
		return fmt.Errorf("error when Unmarshal %s file: %v", path, err)
	}
	var errors []string
	for k, v := range result {
		err = os.Setenv(k, v)
		if err != nil {
			errors = append(errors, fmt.Errorf("error when setting %s env var: %v", k, err).Error())
		}
	}
	if len(errors) > 0 {
		return fmt.Errorf("errors when setting env vars: %s", strings.Join(errors, ", "))
	}
	return nil
}
