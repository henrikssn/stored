package store

import (
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

func (c *Client) Get(key string) (*StoreItem, error) {
	var item *StoreItem
	err := c.connection.Call("Store.Get", key, &item)
	return item, err
}

func (c *Client) Put(item *StoreItem) (bool, error) {
	var added bool
	err := c.connection.Call("Store.Put", item, &added)
	return added, err
}

func (c *Client) Delete(key string) (bool, error) {
	var deleted bool
	err := c.connection.Call("Store.Delete", key, &deleted)
	return deleted, err
}

func (c *Client) Clear() (bool, error) {
	var cleared bool
	err := c.connection.Call("Store.Clear", true, &cleared)
	return cleared, err
}
