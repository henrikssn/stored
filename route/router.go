package route

import (
	"log"
	"net"
)

type Router struct {
	conn *net.UDPConn
	in   chan Packet
	out  chan Packet
}

func NewRouter(conn *net.UDPConn) *Router {
	r := &Router{
		conn: conn,
		in:   make(chan Packet, 100),
		out:  make(chan Packet, 100),
	}
	return r
}

func (r *Router) Start() *Router {
	go r.handleInbound()
	go r.handleOutbound()
	return r
}

func (r *Router) Send(packet Packet) {
	r.out <- packet
}

func (r *Router) GetReply() chan Packet {
	return r.in
}

func (r *Router) handleOutbound() {
	for packet := range r.out {
		n, err := r.conn.WriteTo(packet.Data, packet.Addr)
		log.Printf("Sent %d bytes to %s", n, packet.Addr.String())
		if err != nil {
			log.Println(err)
			continue
		}
		if n != len(packet.Data) {
			log.Println("packet len too long:", len(packet.Data))
			continue
		}
	}
}

func (r *Router) handleInbound() {
	for {
		buf := make([]byte, 1024)
		n, addr, err := r.conn.ReadFromUDP(buf)
		log.Printf("Received %d bytes from %s", n, addr.String())
		if err != nil {
			log.Println(err)
			continue
		}
		buf = buf[:n]
		r.in <- Packet{addr, buf}
	}
}
