package domain

import "sync"

type Store struct {
	mu       sync.RWMutex
	buffer   []GCEvent
	start    int
	count    int
	capacity int
}

func NewStore(windowSize int) *Store {
	if windowSize <= 0 {
		windowSize = 1
	}

	return &Store{
		buffer:   make([]GCEvent, windowSize),
		capacity: windowSize,
	}
}

func (s *Store) Add(event GCEvent) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.count < s.capacity {
		idx := (s.start + s.count) % s.capacity
		s.buffer[idx] = event
		s.count++
		return
	}

	s.buffer[s.start] = event
	s.start = (s.start + 1) % s.capacity
}

func (s *Store) Recent() []GCEvent {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]GCEvent, s.count)
	for i := 0; i < s.count; i++ {
		idx := (s.start + i) % s.capacity
		result[i] = s.buffer[idx]
	}

	return result
}

func (s *Store) Len() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.count
}

func (s *Store) Capacity() int {
	return s.capacity
}
