package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/yokitheyo/todo/internal/handler"
	"github.com/yokitheyo/todo/internal/repository/memory"
	"github.com/yokitheyo/todo/internal/service"
	"github.com/yokitheyo/todo/pkg/logger"
)

func main() {
	log := logger.New(getEnv("LOG_LEVEL", "info"), os.Stdout, "json")
	log.Info("starting todo api server")

	memRepo := memory.NewTodoRepository()
	todoService := service.NewTodoService(memRepo)

	timeout := time.Duration(getEnvAsInt("REQUEST_TIMEOUT", 30)) * time.Second
	todoHandler := handler.NewTodoHandler(todoService, log, timeout)

	mux := http.NewServeMux()
	todoHandler.RegisterRoutes(mux)

	port := getEnv("PORT", "8080")
	server := &http.Server{
		Addr:         ":" + port,
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Info("server listening", "port", port)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("shutting down server")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Error("server forced to shutdown", "error", err)
	}

	log.Info("server stopped")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}
