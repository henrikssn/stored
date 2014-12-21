package endpoint

import (
	"bytes"
	"io/ioutil"
	"net/http"
)

type (
	Client struct {
		connection *http.Client
		url        string
	}
)

func NewClient(addr string) (*Client, error) {
	c := &http.Client{}
	url := addr + "/"
	return &Client{connection: c, url: url}, nil
}

func (c *Client) Get(key string) ([]byte, error) {
	req, err := http.NewRequest("GET", c.url+key, nil)
	resp, err := c.connection.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	data, _ := ioutil.ReadAll(resp.Body)
	return data, nil
}

func (c *Client) Put(key string, data []byte) error {
	req, err := http.NewRequest("PUT", c.url+key, bytes.NewReader(data))
	_, err = c.connection.Do(req)
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) Delete(key string) error {
	req, err := http.NewRequest("DELETE", c.url+key, nil)
	_, err = c.connection.Do(req)
	if err != nil {
		return err
	}
	return nil
}
