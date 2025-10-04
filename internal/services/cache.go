package services

import (
	"sync"
	"time"
)

type SetOptions struct {
	TTL time.Duration // Time to live in seconds
	NX  bool          // Only set if key does not exist
}

type CacheService interface {
	Get(key string) (string, bool)
	Set(key, value string, opts *SetOptions) bool
	Delete(key string) error
}

type InMemoryCacheService struct {
	cache       map[string]string
	expirations map[string]time.Time
	mu          sync.RWMutex
}

func NewInMemoryCacheService() *InMemoryCacheService {
	cacheService := &InMemoryCacheService{
		cache:       make(map[string]string),
		expirations: make(map[string]time.Time),
		mu:          sync.RWMutex{},
	}

	// Start a goroutine to clean up expired keys
	go cacheService.cleanupExpiredKeys()

	return cacheService
}

func (s *InMemoryCacheService) cleanupExpiredKeys() {
	ticker := time.NewTicker(1 * time.Millisecond)
	for range ticker.C {
		s.mu.Lock()
		now := time.Now()
		for key, expTime := range s.expirations {
			if now.After(expTime) {
				delete(s.cache, key)
				delete(s.expirations, key)
			}
		}
		s.mu.Unlock()
	}
}

func (s *InMemoryCacheService) Get(key string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	value, exists := s.cache[key]
	return value, exists
}

func (s *InMemoryCacheService) Set(key, value string, opts *SetOptions) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if opts != nil && opts.NX {
		if _, exists := s.cache[key]; exists {
			return false
		}
	}

	s.cache[key] = value
	if opts != nil && opts.TTL > 0 {
		s.expirations[key] = time.Now().Add(opts.TTL)
	}

	return true
}

func (s *InMemoryCacheService) Delete(key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.cache, key)
	return nil
}
