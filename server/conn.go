package server

import (
	"code.google.com/p/goprotobuf/proto"
	"encoding/binary"
	"github.com/henrikssn/stored/overlay"
	"io"
	"log"
)

type conn struct {
	nc      io.ReadWriter
	overlay *overlay.Overlay
}

func (c *conn) serve() {
	for {
		var t txn
		t = txn{conn: c, o: c.overlay}
		err := c.read(&t.req)
		log.Printf("Got request: %s", t.req)
		if err != nil {
			if err != io.EOF {
				log.Println(err)
			}
			return
		}
		t.run()
	}
}

func (c *conn) read(r *Request) error {
	var size int32
	err := binary.Read(c.nc, binary.BigEndian, &size)
	if err != nil {
		return err
	}

	buf := make([]byte, size)
	_, err = io.ReadFull(c.nc, buf)
	if err != nil {
		return err
	}

	return proto.Unmarshal(buf, r)
}

func (c *conn) write(r *Response) error {
	buf, err := proto.Marshal(r)
	if err != nil {
		return err
	}
	log.Printf("Sending response: %s", r)

	err = binary.Write(c.nc, binary.BigEndian, int32(len(buf)))
	if err != nil {
		return err
	}

	_, err = c.nc.Write(buf)
	return err
}
