package store

import (
	"errors"
	"log"
	"sync"
)

var _ = log.Printf

type (
	Store struct {
		items map[string][]byte
		mu    *sync.RWMutex
	}

	StoreItem struct {
		Key   string
		Value []byte
	}
)

var (
	NotFoundError = errors.New("Key not found")
)

func New() *Store {
	return &Store{
		items: make(map[string][]byte),
		mu:    &sync.RWMutex{},
	}
}

func (r *Store) Get(key string, resp *StoreItem) (err error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	item, found := r.items[key]

	if !found {
		return NotFoundError
	}

	*resp = StoreItem{key, item}
	return nil
}

func (r *Store) Put(item *StoreItem, added *bool) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	_, ok := r.items[item.Key]
	*added = !ok

	r.items[item.Key] = item.Value

	return nil
}

func (r *Store) Delete(key string, ack *bool) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	var found bool
	_, found = r.items[key]

	if !found {
		return NotFoundError
	}

	delete(r.items, key)
	*ack = true

	return nil
}

func (r *Store) Clear(skip bool, ack *bool) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.items = make(map[string][]byte)
	*ack = true

	return nil
}
