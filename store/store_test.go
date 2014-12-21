package store

import (
	"bytes"
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

	dsn       = "localhost:9876"
	storeItem = &StoreItem{Key: "some key", Value: []byte{42}}
)

func init() {
	if err != nil {
		log.Fatal(err)
	}
	startStore()
	c, err = NewClient(dsn, 500*time.Millisecond)
}

func startStore() {
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
}

func TestGetEmptyMap(t *testing.T) {
	item, _ := c.Get(storeItem.Key)
	if item != nil {
		t.Errorf("Store key should not exist: %s\n", storeItem.Key)
	}
}

func TestPut(t *testing.T) {
	_, err := c.Put(storeItem)
	if err != nil {
		t.Error(err)
	}
}

func TestGet(t *testing.T) {
	item, _ := c.Get(storeItem.Key)
	if item == nil {
		t.Errorf("Key should exist: %s\n", storeItem.Key)
	}
	if !bytes.Equal(item.Value, storeItem.Value) {
		t.Errorf("Item expected %s got %s\n", storeItem, item)
	}
}

func TestDelete(t *testing.T) {
	_, err := c.Delete(storeItem.Key)
	if err != nil {
		t.Error(err)
	}
}

func BenchmarkPut(b *testing.B) {
	for i := 0; i < b.N; i++ {
		c.Put(&StoreItem{Key: strconv.Itoa(i)})
	}
}

func BenchmarkGet(b *testing.B) {
	for i := 0; i < b.N; i++ {
		c.Get(strconv.Itoa(i))
	}
}
