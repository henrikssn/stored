package router

import (
	"github.com/henrikssn/stored/store"
	"net"
	"net/rpc"
	"time"
)

type (
	Client struct {
		connection *rpc.Client
	}
)

func NewClient(addr string, timeout time.Duration) (*Client, error) {
	connection, err := net.DialTimeout("tcp", addr, timeout)
	if err != nil {
		return nil, err
	}
	return &Client{connection: rpc.NewClient(connection)}, nil
}

func (c *Client) Get(key string) (*store.StoreItem, error) {
	var item *store.StoreItem
	err := c.connection.Call("Router.Get", key, &item)
	return item, err
}

func (c *Client) Put(item *store.StoreItem) (bool, error) {
	var added bool
	err := c.connection.Call("Router.Put", item, &added)
	return added, err
}

func (c *Client) Delete(key string) (bool, error) {
	var deleted bool
	err := c.connection.Call("Router.Delete", key, &deleted)
	return deleted, err
}

func (c *Client) Clear() (bool, error) {
	var cleared bool
	err := c.connection.Call("Router.Clear", true, &cleared)
	return cleared, err
}

func (c *Client) AddStore(addr string) (bool, error) {
	var ok bool
	err := c.connection.Call("Router.AddStore", addr, &ok)
	return ok, err
}
