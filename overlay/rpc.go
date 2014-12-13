package overlay

import (
	"code.google.com/p/goprotobuf/proto"
	"github.com/henrikssn/stored/route"
	"log"
	"net"
)

type rpc struct {
	ch  chan []byte
	msg *Msg
}

type RpcHandler struct {
	seq         int64
	reqs        map[int64]rpc
	action      map[Msg_Cmd]func([]byte) proto.Message
	info_action map[Msg_Cmd]func([]byte)
}

func NewRpcHandler() *RpcHandler {
	r := new(RpcHandler)
	r.seq = 0
	r.reqs = make(map[int64]rpc)
	r.action = make(map[Msg_Cmd]func([]byte) proto.Message)
	r.info_action = make(map[Msg_Cmd]func([]byte))
	return r
}

func (rh *RpcHandler) nextSeqn() int64 {
	rh.seq += 1
	return rh.seq
}

// Handles incoming packets
func (rh *RpcHandler) receive(r *route.Router) {
	for packet := range r.GetReply() {
		msg := new(Msg)
		proto.Unmarshal(packet.Data, msg)
		log.Printf("Processing message: %s", msg)
		switch msg.GetType() {
		case Msg_REQUEST:
			rh.handleRequest(msg, packet.Addr, r)
		case Msg_RESPONSE:
			rh.handleResponse(msg)
		case Msg_INFO:
			rh.handleInfo(msg)
		}
	}
	log.Println("No more packets")
}

func (rh *RpcHandler) Send(cmd Msg_Cmd, req proto.Message, addr *net.UDPAddr, r *route.Router) chan []byte {
	data, _ := proto.Marshal(req)
	msg := new(Msg)
	msg.Cmd = cmd.Enum()
	msg.Type = Msg_REQUEST.Enum()
	seqn := rh.nextSeqn()
	msg.Seqn = &seqn
	msg.Data = data
	b, _ := proto.Marshal(msg)
	r.Send(route.Packet{addr, b})
	rh.reqs[seqn] = rpc{ch: make(chan []byte, 1), msg: msg}
	return rh.reqs[seqn].ch
}

func (rh *RpcHandler) SendInfo(cmd Msg_Cmd, req proto.Message, addr *net.UDPAddr, r *route.Router) {
	data, _ := proto.Marshal(req)
	msg := new(Msg)
	msg.Cmd = cmd.Enum()
	msg.Type = Msg_INFO.Enum()
	msg.Data = data
	b, _ := proto.Marshal(msg)
	r.Send(route.Packet{addr, b})
}

func (rh *RpcHandler) RegisterAction(cmd Msg_Cmd, action func([]byte) proto.Message) {
	rh.action[cmd] = action
}

func (rh *RpcHandler) RegisterInfoAction(cmd Msg_Cmd, action func([]byte)) {
	rh.info_action[cmd] = action
}

// Invokes the rpc action and sends back the response.
func (rh *RpcHandler) handleRequest(msg *Msg, addr *net.UDPAddr, r *route.Router) {
	resp := rh.action[msg.GetCmd()](msg.GetData())
	data, err := proto.Marshal(resp)
	if err != nil {
		log.Println(err)
	}
	msg.Type = Msg_RESPONSE.Enum()
	msg.Data = data
	b, _ := proto.Marshal(msg)
	r.Send(route.Packet{addr, b})
}

func (rh *RpcHandler) handleResponse(msg *Msg) {
	ch := rh.reqs[msg.GetSeqn()].ch
	delete(rh.reqs, msg.GetSeqn())
	ch <- msg.Data
}

func (rh *RpcHandler) handleInfo(msg *Msg) {
	rh.info_action[msg.GetCmd()](msg.GetData())
}
