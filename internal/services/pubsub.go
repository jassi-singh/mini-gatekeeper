package services

import (
	"errors"
	"sync"
)

type PubSubService interface {
	Publish(topic string, message string) error
	Subscribe(topic string) (chan string, error)
	Unsubscribe(topic string) error
}

type InMemoryPubSubService struct {
	subscribers map[string]chan string
	mu          sync.RWMutex
}

func NewInMemoryPubSubService() *InMemoryPubSubService {
	return &InMemoryPubSubService{
		subscribers: make(map[string]chan string),
		mu:          sync.RWMutex{},
	}
}

func (s *InMemoryPubSubService) Publish(topic string, message string) error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if msgChan, exists := s.subscribers[topic]; exists {
		msgChan <- message
	}
	return nil
}

func (s *InMemoryPubSubService) Subscribe(topic string) (chan string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	msgChan := make(chan string)
	s.subscribers[topic] = msgChan
	return msgChan, nil

}

func (s *InMemoryPubSubService) Unsubscribe(topic string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.subscribers[topic]; exists {
		delete(s.subscribers, topic)
		return nil
	}
	return errors.New("no subscribers for topic")
}
