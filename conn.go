package swallow

import "net"

type Conn struct {
	net.Conn
	name   string
	able   bool
	buffer []byte
}

func NewConn(c net.Conn, len int) *Conn {
	return &Conn{
		Conn: c,
		able: true,
	}
}
