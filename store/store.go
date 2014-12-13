package store

import (
	"log"
)

type Store struct {
	m map[string]*Entry
}

func New() *Store {
	s := new(Store)
	s.m = make(map[string]*Entry)
	return s
}

func (s *Store) Put(k string, e *Entry) {
	s.m[k] = e
	log.Printf("Store PUT %s: %s", k, e)
}

func (s *Store) Get(k string) (*Entry, bool) {
	e, ok := s.m[k]
	log.Printf("Store GET %s: %s", k, e)
	return e, ok
}
