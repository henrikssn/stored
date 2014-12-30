package router

import (
	"bytes"
	"github.com/henrikssn/stored/store"
	"log"
	"net"
	"net/rpc"
	"strconv"
	"testing"
	"time"
)

var (
	c   *Client
	err error

	dsn    = "localhost:9876"
	stores = []string{"localhost:9877", "localhost:9878"}
	item   = &store.StoreItem{Key: "some key", Value: []byte{42}}
)

func init() {
	if err != nil {
		log.Fatal(err)
	}
	startup()
	c, err = NewClient(dsn, 500*time.Millisecond)
}

func startup() {
	rpc.Register(New())

	l, e := net.Listen("tcp", dsn)
	if e != nil {
		log.Fatal("listen error:", e)
	}
	go func() {
		for {
			conn, _ := l.Accept()
			go rpc.ServeConn(conn)
		}
	}()
	for _, addr := range stores {
		go startStore(addr)
	}
}

func startStore(addr string) {
	rpc.Register(store.New())

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

func TestAddStore(t *testing.T) {
	for _, addr := range stores {
		_, err := c.AddStore(addr)
		if err != nil {
			t.Error(err)
		}
	}
}

func TestGetEmptyMap(t *testing.T) {
	i, _ := c.Get(item.Key)
	if i != nil {
		t.Errorf("Store key should not exist: %s\n", item.Key)
	}
}

func TestPut(t *testing.T) {
	_, err := c.Put(item)
	if err != nil {
		t.Error(err)
	}
}

func TestGet(t *testing.T) {
	i, err := c.Get(item.Key)
	if err != nil {
		t.Error(err)
	}
	if i == nil {
		t.Errorf("Key should exist: %s\n", item.Key)
	}
	if !bytes.Equal(i.Value, item.Value) {
		t.Errorf("Item expected %s got %s\n", item, i)
	}
}

func BenchmarkPut(b *testing.B) {
	for i := 0; i < b.N; i++ {
		c.Put(&store.StoreItem{Key: strconv.Itoa(i)})
	}
}

func BenchmarkGet(b *testing.B) {
	for i := 0; i < b.N; i++ {
		c.Get(strconv.Itoa(i))
	}
}
