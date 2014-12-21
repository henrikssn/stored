package store

import (
	"errors"
	"log"
	"sync"
)

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
	log.Printf("Store.Get(%s)=%s", key, resp)
	return nil
}

func (r *Store) Put(item *StoreItem, ack *bool) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.items[item.Key] = item.Value
	*ack = true

	log.Printf("Store.Put(%s)", item)
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

	log.Printf("Store.Delete(%s)", key)
	return nil
}

func (r *Store) Clear(skip bool, ack *bool) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.items = make(map[string][]byte)
	*ack = true

	return nil
}
