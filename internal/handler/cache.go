package handler

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/jassi-singh/mini-gatekeep/internal/services"
)

type CacheHandler struct {
	cacheService  services.CacheService
	pubsubService services.PubSubService
	logger        *slog.Logger
}

func NewCacheHandler(cacheService services.CacheService, pubsub services.PubSubService, logger *slog.Logger) *CacheHandler {
	return &CacheHandler{cacheService: cacheService, logger: logger, pubsubService: pubsub}
}

func (h *CacheHandler) GetFromCache(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("key")
	h.logger.Info("Received request for cache", "key", key)
	value, exists := h.cacheService.Get(key)

	if exists && value != "" {
		h.logger.Info("Cache hit", "key", key, "value", value)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(value))
		return
	}
	h.logger.Info("Cache miss", "key", key)
	// try to get lock for this in key in cache
	value, err := handleCacheMiss(h.logger, h.cacheService, h.pubsubService, key)
	if err != nil {
		h.logger.Error("Error handling cache miss", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(value))

}

func getFromSlowQuery(logger *slog.Logger) string {
	logger.Info("Fetching slow query...")
	time.Sleep(5 * time.Second)
	logger.Info("Fetched value from slow query")
	return "value_from_slow_query"
}

func handleCacheMiss(logger *slog.Logger, cacheService services.CacheService, pubsubService services.PubSubService, key string) (string, error) {
	value := ""
	set := cacheService.Set(key, value, &services.SetOptions{NX: true, TTL: 10 * time.Second})
	if set {
		value = getFromSlowQuery(logger)
		cacheService.Set(key, value, &services.SetOptions{TTL: 15 * time.Second})
		pubsubService.Publish(key, value)
		logger.Info("Set value in cache", "key", key, "value", value)
		return value, nil
	} else {
		logger.Info("Waiting for value to be set in cache", "key", key)
		msgChan, id := pubsubService.Subscribe(key)
		defer pubsubService.Unsubscribe(key, id)
		value, exists := cacheService.Get(key)
		if exists && value != "" {
			logger.Info("Found value in cache after subscribing", "key", key)
			return value, nil
		}

		select {
		case msg := <-msgChan:
			logger.Info("Received value from pubsub", "key", key, "value", msg)
			return msg, nil
		case <-time.After(10 * time.Second):
			logger.Error("Timeout waiting for value to be set in cache", "key", key)
			return handleCacheMiss(logger, cacheService, pubsubService, key)
		}
	}
}
