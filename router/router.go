package router

import (
	"github.com/henrikssn/stored/store"
	"github.com/stathat/consistent"
	"log"
	"sync"
	"time"
)

var _ = log.Printf

type (
	Router struct {
		clients  map[string]*store.Client
		hashRing *consistent.Consistent
		mu       *sync.RWMutex
		Replicas int
	}
)

func New() *Router {
	r := &Router{
		clients:  make(map[string]*store.Client),
		hashRing: consistent.New(),
		mu:       &sync.RWMutex{},
		Replicas: 2,
	}
	return r
}

func (r *Router) Get(key string, resp *store.StoreItem) (err error) {
	c, err := r.getClientForKey(key)
	item, err := c.Get(key)
	if err != nil {
		return err
	}
	*resp = *item
	return nil
}

func (r *Router) Put(item *store.StoreItem, added *bool) error {
	cs, err := r.getClientsForKey(item.Key)
	if err != nil {
		return err
	}
	for _, c := range cs {
		*added, err = c.Put(item)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *Router) Delete(key string, ack *bool) error {
	cs, err := r.getClientsForKey(key)
	if err != nil {
		return err
	}
	for _, c := range cs {
		_, err := c.Delete(key)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *Router) AddStore(addr string, ok *bool) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	c, err := store.NewClient(addr, 500*time.Millisecond)
	if err != nil {
		return err
	}
	r.clients[addr] = c
	r.hashRing.Add(addr)
	return nil
}

func (r *Router) getClientsForKey(key string) ([]*store.Client, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	names, err := r.hashRing.GetN(key, r.Replicas)
	if err != nil {
		return nil, err
	}
	cs := make([]*store.Client, 0, len(names))
	for _, value := range names {
		c, _ := r.clients[value]
		cs = append(cs, c)
	}
	return cs, nil
}

func (r *Router) getClientForKey(key string) (*store.Client, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	s, err := r.hashRing.Get(key)
	if err != nil {
		return nil, err
	}
	c, _ := r.clients[s]
	return c, nil
}
