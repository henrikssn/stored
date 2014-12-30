package endpoint

import (
	"bytes"
	"github.com/henrikssn/stored/router"
	"github.com/henrikssn/stored/store"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"strconv"
	"testing"
	"time"
)

var (
	c   *Client
	r   *router.Router
	s   *store.Store
	e   *Endpoint
	err error

	dsn        = "localhost:9700"
	httpAddr   = "http://" + dsn
	storeAddr  = "localhost:9800"
	routerAddr = "localhost:9900"
	item       = &store.StoreItem{Key: "some key", Value: []byte{42}}
	key        = Key{Namespace: "default", Group: "default", Id: "foo"}
)

func init() {
	if err != nil {
		log.Fatal(err)
	}
	e = New()
	r = router.New()
	s = store.New()
	startup()
	var ok bool
	r.AddStore(storeAddr, &ok)
	e.internal.AddRouter(routerAddr, &ok)
	time.Sleep(1)
	c, err = NewClient(httpAddr)
}

func startup() {
	go startEndpoint(dsn)
	go startStore(storeAddr)
	go startRouter(routerAddr)
}

func startStore(addr string) {
	rpc.Register(s)

	l, e := net.Listen("tcp", addr)
	if e != nil {
		log.Fatal("listen error:", e)
	}
	go func() {
		for {
			conn, _ := l.Accept()
			go rpc.ServeConn(conn)
		}
	}()
}

func startRouter(addr string) {
	rpc.Register(r)

	l, e := net.Listen("tcp", addr)
	if e != nil {
		log.Fatal("listen error:", e)
	}
	go func() {
		for {
			conn, _ := l.Accept()
			go rpc.ServeConn(conn)
		}
	}()
}

func startEndpoint(addr string) {
	http.HandleFunc("/", e.StoreHandler)
	http.ListenAndServe(addr, nil)
}

func TestGetEmpty(t *testing.T) {
	val, err := c.Get(key)
	if err != nil {
		t.Error(err)
	}
	if !bytes.Equal(val, []byte("Error: Key not found\n")) {
		t.Errorf("Expected: <Error: Key not found\\n> got: %s", val)
	}
}

func TestPutGet(t *testing.T) {
	data := []byte{42}
	err := c.Put(key, data)
	if err != nil {
		t.Error(err)
	}
	actual, err := c.Get(key)
	if err != nil {
		t.Error(err)
	}
	if !bytes.Equal(actual, data) {
		t.Errorf("Expected: %s got %s", data, actual)
	}
}

func TestDelete(t *testing.T) {
	err := c.Delete(key)
	if err != nil {
		t.Error(err)
	}
}

func TestGetAfterDelete(t *testing.T) {
	val, err := c.Get(key)
	if err != nil {
		t.Error(err)
	}
	if !bytes.Equal(val, []byte("Error: Key not found\n")) {
		t.Errorf("Expected: <Error: Key not found\\n> got: %s", val)
	}
}

func TestHttpErrors(t *testing.T) {
	data := []byte{42}

	c.Get(key)
	if c.Response.StatusCode != http.StatusNotFound {
		t.Errorf("Wrong error code: %s", c.Response.Status)
	}

	c.Put(key, data)
	if c.Response.StatusCode != http.StatusCreated {
		t.Errorf("Wrong error code: %s", c.Response.Status)
	}

	c.Put(key, data)
	if c.Response.StatusCode != http.StatusOK {
		t.Errorf("Wrong error code: %s", c.Response.Status)
	}

	c.Get(key)
	if c.Response.StatusCode != http.StatusOK {
		t.Errorf("Wrong error code: %s", c.Response.Status)
	}

	c.Delete(key)
	if c.Response.StatusCode != http.StatusOK {
		t.Errorf("Wrong error code: %s", c.Response.Status)
	}

	c.Delete(key)
	if c.Response.StatusCode != http.StatusNotFound {
		t.Errorf("Wrong error code: %s", c.Response.Status)
	}

	c.Get(key)
	if c.Response.StatusCode != http.StatusNotFound {
		t.Errorf("Wrong error code: %s", c.Response.Status)
	}
}

func BenchmarkPut(b *testing.B) {
	for i := 0; i < b.N; i++ {
		c.Put(Key{Id: strconv.Itoa(i)}, []byte{})
	}
}

func BenchmarkGet(b *testing.B) {
	for i := 0; i < b.N; i++ {
		c.Get(Key{Id: strconv.Itoa(i)})
	}
}
