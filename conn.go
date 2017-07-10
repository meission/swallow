package swallow

import "net"

type Conn struct {
	net.Conn
	able bool
}

func NewConn(c net.Conn, len int) *Conn {
	return &Conn{
		Conn: c,
		able: true,
	}
}
