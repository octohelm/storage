package sqlite

import "sync"

type mutexSet struct {
	m sync.Map
}

func (s *mutexSet) of(path string) *sync.Mutex {
	got, _ := s.m.LoadOrStore(path, &sync.Mutex{})
	return got.(*sync.Mutex)
}
