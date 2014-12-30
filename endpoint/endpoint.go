package endpoint

import (
	"errors"
	"github.com/henrikssn/stored/router"
	"github.com/henrikssn/stored/store"
	"github.com/stathat/consistent"
	"io/ioutil"
	"log"
	"net/http"
	_ "net/http/pprof"
	"net/rpc"
	"strings"
	"sync"
	"time"
)

var _ = log.Printf

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
	log.Println(http.ListenAndServe(httpAddr, nil))
}

func (e *Endpoint) StoreHandler(w http.ResponseWriter, req *http.Request) {
	namespace, group, id, err := parseURI(req.URL.RequestURI())
	if err != nil {
		log.Printf("HTTP Action returned error: %s", err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Invalid URI: " + err.Error() + "\n"))
		return
	}

	// Temporal hack, will be replaced when store can properly handle gruops
	key := namespace + "/" + group + "/" + id

	var resp []byte
	switch req.Method {
	case "GET":
		resp, err = e.Get(key)
		if err != nil {
			break
		}
		w.Write(resp)
	case "PUT":
		data, err := ioutil.ReadAll(req.Body)
		if err != nil {
			break
		}
		added, err := e.Put(key, data)
		if added {
			w.WriteHeader(http.StatusCreated)
		}
	case "DELETE":
		_, err = e.Delete(key)
	}

	if err != nil {
		log.Printf("HTTP Action returned error: %s", err)
		if err.Error() == "Key not found" {
			w.WriteHeader(http.StatusNotFound)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
		w.Write([]byte("Error: " + err.Error() + "\n"))
	}
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
	return item.Value, err
}

func (e *Endpoint) Put(key string, data []byte) (bool, error) {
	r, err := e.internal.getRouterForKey(key)
	if err != nil {
		return false, err
	}
	added, err := r.Put(&store.StoreItem{Key: key, Value: data})
	return added, err
}

func (e *Endpoint) Delete(key string) ([]byte, error) {
	r, err := e.internal.getRouterForKey(key)
	if err != nil {
		return nil, err
	}
	_, err = r.Delete(key)
	return nil, err
}

func (e *Endpoint) AddRouter(addr string) error {
	var ok bool
	return e.internal.AddRouter(addr, &ok)
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

func parseURI(uri string) (string, string, string, error) {
	s := strings.Split(uri, "/")
	if len(s) != 4 {
		return "", "", "", errors.New("URI " + uri + " does not match /[namespace]/[key]/[id]")
	}
	return s[0], s[1], s[2], nil
}
