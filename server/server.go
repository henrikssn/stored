package server

import (
	"github.com/henrikssn/stored/overlay"
	"log"
	"net"
)

type Server struct {
	overlay *overlay.Overlay
}

func New(o *overlay.Overlay) *Server {
	return &Server{overlay: o}
}

func (s *Server) ListenAndServe(l net.Listener) {
	for {
		conn, err := l.Accept()
		if err != nil {
			log.Println(err)
			continue
		}
		go s.serve(conn)
	}
}

func (s *Server) serve(nc net.Conn) {
	c := &conn{
		nc:      nc,
		overlay: s.overlay,
	}
	c.serve()
	nc.Close()
}
