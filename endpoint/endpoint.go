package endpoint

import (
	"github.com/henrikssn/stored/router"
	"github.com/henrikssn/stored/store"
	"github.com/stathat/consistent"
	"io/ioutil"
	"log"
	"net/http"
	"net/rpc"
	"sync"
	"time"
)

type (
	Endpoint struct {
		internal *EndpointInternal
	}
	EndpointInternal struct {
		routers  map[string]*router.Client
		hashRing *consistent.Consistent
		mu       *sync.RWMutex
	}
)

func New() *Endpoint {
	i := &EndpointInternal{
		routers:  make(map[string]*router.Client),
		hashRing: consistent.New(),
		mu:       &sync.RWMutex{},
	}
	return &Endpoint{internal: i}
}

func (e *Endpoint) RegisterInternalRPC() {
	rpc.Register(e.internal)
}

func (e *Endpoint) Listen(httpAddr string) {
	http.HandleFunc("/", e.StoreHandler)
	http.ListenAndServe(httpAddr, nil)
}

func (e *Endpoint) StoreHandler(w http.ResponseWriter, req *http.Request) {
	key := req.URL.RequestURI()
	if key == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	var resp []byte
	var err error
	switch req.Method {
	case "GET":
		resp, err = e.Get(key)
	case "PUT":
		data, err := ioutil.ReadAll(req.Body)
		if err != nil {
			break
		}
		resp, err = e.Put(key, data)
	case "DELETE":
		resp, err = e.Delete(key)
	}
	if err != nil {
		w.WriteHeader(400)
		return
	}
	w.Write(resp)
}

func (e *Endpoint) Get(key string) ([]byte, error) {
	r, err := e.internal.getRouterForKey(key)
	if err != nil {
		return nil, err
	}
	item, err := r.Get(key)
	if err != nil {
		return nil, err
	}
	log.Printf("Endpoint.Get(%s)=%s", key, item.Value)
	return item.Value, err
}

func (e *Endpoint) Put(key string, data []byte) ([]byte, error) {
	r, err := e.internal.getRouterForKey(key)
	if err != nil {
		return nil, err
	}
	_, err = r.Put(&store.StoreItem{Key: key, Value: data})
	log.Printf("Endpoint.Put(%s, %s)", key, data)
	return nil, err
}

func (e *Endpoint) Delete(key string) ([]byte, error) {
	r, err := e.internal.getRouterForKey(key)
	if err != nil {
		return nil, err
	}
	_, err = r.Delete(key)
	log.Printf("Endpoint.Delete(%s)", key)
	return nil, err
}

func (e *EndpointInternal) AddRouter(addr string, ok *bool) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	c, err := router.NewClient(addr, 500*time.Millisecond)
	if err != nil {
		return err
	}
	e.routers[addr] = c
	e.hashRing.Add(addr)
	return nil
}

func (e *EndpointInternal) getRouterForKey(key string) (*router.Client, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	s, err := e.hashRing.Get(key)
	if err != nil {
		return nil, err
	}
	c, _ := e.routers[s]
	return c, nil
}
