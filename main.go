package main

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jassi-singh/mini-gatekeep/internal/handler"
	"github.com/jassi-singh/mini-gatekeep/internal/services"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	logger.Info("Initializing server...")

	router := chi.NewRouter()
	router.Use(middleware.Logger)

	cacheService := services.NewInMemoryCacheService()
	pubsubService := services.NewInMemoryPubSubService()

	cacheHandler := handler.NewCacheHandler(cacheService, pubsubService, logger)

	router.Get("/cache", cacheHandler.GetFromCache)

	logger.Info("Starting server on :8080")
	err := http.ListenAndServe(":8080", router)
	if err != nil {
		logger.Error("Failed to start server", slog.String("error", err.Error()))
	}

}
