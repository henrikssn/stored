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
		Response   *http.Response
	}
	Key struct {
		Namespace string
		Group     string
		Id        string
	}
)

func (k *Key) Str() string {
	return k.Namespace + "/" + k.Group + "/" + k.Id
}

func NewClient(addr string) (*Client, error) {
	c := &http.Client{}
	url := addr + "/"
	return &Client{connection: c, url: url}, nil
}

func (c *Client) Get(key Key) ([]byte, error) {
	req, err := http.NewRequest("GET", c.url+key.Str(), nil)
	resp, err := c.connection.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	data, _ := ioutil.ReadAll(resp.Body)
	c.Response = resp
	return data, nil
}

func (c *Client) Put(key Key, data []byte) error {
	req, err := http.NewRequest("PUT", c.url+key.Str(), bytes.NewReader(data))
	c.Response, err = c.connection.Do(req)
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) Delete(key Key) error {
	req, err := http.NewRequest("DELETE", c.url+key.Str(), nil)
	c.Response, err = c.connection.Do(req)
	if err != nil {
		return err
	}
	return nil
}
