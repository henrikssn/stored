package route

import (
  "net"
)

type Packet struct {
	Addr *net.UDPAddr
	Data []byte
}
