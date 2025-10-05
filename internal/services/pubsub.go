package services

import (
	"errors"
	"sync"

	"github.com/google/uuid"
)

type PubSubService interface {
	Publish(topic string, message string) error
	Subscribe(topic string) (chan string, uuid.UUID)
	Unsubscribe(topic string, id uuid.UUID) error
}

type InMemoryPubSubService struct {
	subscribers map[string]map[uuid.UUID]chan string
	mu          sync.RWMutex
}

func NewInMemoryPubSubService() *InMemoryPubSubService {
	return &InMemoryPubSubService{
		subscribers: make(map[string]map[uuid.UUID]chan string),
		mu:          sync.RWMutex{},
	}
}

func (s *InMemoryPubSubService) Publish(topic string, message string) error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if channels, exists := s.subscribers[topic]; exists {
		for _, ch := range channels {
			ch <- message
		}
	}
	return nil
}

func (s *InMemoryPubSubService) Subscribe(topic string) (chan string, uuid.UUID) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.subscribers[topic]; !exists {
		s.subscribers[topic] = make(map[uuid.UUID]chan string)
	}
	id := uuid.New()
	msgChan := make(chan string)
	s.subscribers[topic][id] = msgChan
	return msgChan, id

}

func (s *InMemoryPubSubService) Unsubscribe(topic string, id uuid.UUID) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if channels, exists := s.subscribers[topic]; exists {
		if ch, ok := channels[id]; ok {
			close(ch)
			delete(channels, id)
			if len(channels) == 0 {
				delete(s.subscribers, topic)
			}
			return nil
		}
		return errors.New("subscription not found")
	}

	return errors.New("no subscribers for topic")
}
