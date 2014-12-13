package server

import (
	"github.com/henrikssn/stored/overlay"
	"io"
	"log"
)

type txn struct {
	conn *conn
	o    *overlay.Overlay
	req  Request
	resp Response
}

var ops = map[Operation]func(*txn){
	Operation_GET: (*txn).get,
	Operation_PUT: (*txn).put,
	Operation_DEL: (*txn).del,
	Operation_NOP: (*txn).nop,
}

func (t *txn) run() {
	op := t.req.GetOp()
	if f, ok := ops[op]; ok {
		f(t)
	}
}

func (t *txn) get() {
	resp := t.o.Get(t.req.GetKey(), t.req.GetSubKey())
	switch resp.GetStatus() {
	case overlay.GetResponse_OK:
		t.resp.Status = Response_OK.Enum()
	case overlay.GetResponse_NOT_FOUND:
		t.resp.Status = Response_NOT_FOUND.Enum()
		t.respond()
		return
	}
	t.resp.Key = t.req.Key
	t.resp.Value = resp.Value
	t.respond()
}

func (t *txn) put() {
	resp := t.o.Put(t.req.GetKey(), t.req.GetSubKey(), t.req.Value)
	switch resp.GetStatus() {
	case overlay.PutResponse_OK:
		t.resp.Status = Response_OK.Enum()
	case overlay.PutResponse_NOT_FOUND:
		t.resp.Status = Response_NOT_FOUND.Enum()
	}
	t.respond()
}

func (t *txn) del() {
}

func (t *txn) nop() {
	t.respond()
}

func (t *txn) respond() {
	t.resp.Tag = t.req.Tag
	err := t.conn.write(&t.resp)
	if err != nil && err != io.EOF {
		log.Println(err)
	}
}
