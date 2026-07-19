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

func (s *Store) List() []Message {
	s.mu.RLock()
	defer s.mu.RUnlock()

	out := make([]Message, len(s.messages))
	copy(out, s.messages)
	return out
}

func (s *Store) Len() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.messages)
}
