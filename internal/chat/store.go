package chat

import "sync"

type Store struct {
	mu       sync.RWMutex
	messages []Message
	limit    int
}

func NewStore(limit int) *Store {
	return &Store{
		messages: make([]Message, 0, limit),
		limit:    limit,
	}
}

func (s *Store) Add(msg Message) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.messages = append(s.messages, msg)
	if len(s.messages) > s.limit {
		s.messages = s.messages[len(s.messages)-s.limit:]
	}
}

func (s *Store) Prepend(msgs []Message) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.messages = append(msgs, s.messages...)
}

func (s *Store) SetLimit(n int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.limit = n
}

func (s *Store) List() []Message {
	s.mu.RLock()
	defer s.mu.RUnlock()

	out := make([]Message, len(s.messages))
	for i, m := range s.messages {
		out[len(s.messages)-1-i] = m
	}
	return out
}

func (s *Store) Len() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.messages)
}
