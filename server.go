package swallow

import (
	"io"
	"net"
	"strings"
	"sync/atomic"
	"time"
)

// Server implements tcp server.
type Server struct {
	// Default server name is used if left blank.
	Name []byte

	Concurrency int

	ReadBufferSize  int
	WriteBufferSize int

	ReadTimeout  time.Duration
	WriteTimeout time.Duration

	// Maximum number of concurrent client connections allowed per IP.
	MaxConnsPerIP int

	// Maximum number of requests served per connection.
	MaxPerConn int

	perIPConnCounter perIPConnCounter
	//
	buffer []byte
}

// DefaultConcurrency is the maximum number of concurrent connections
const DefaultConcurrency = 256 * 1024

func (s *Server) getConcurrency() int {
	n := s.Concurrency
	if n <= 0 {
		n = DefaultConcurrency
	}
	return n
}

// Serve serves incoming connections from the given listener.
func (s *Server) Serve(ln net.Listener) error {
	var (
		c   net.Conn
		err error
	)
	maxWorkersCount := s.getConcurrency()
	rp := &workerPool{
		WorkerFunc:      s.serveConn,
		MaxWorkersCount: maxWorkersCount,
	}
	rp.Start()
	go s.handle()
	for {
		if c, err = acceptConn(s, ln); err != nil {
			rp.Stop()
			if err == io.EOF {
				return nil
			}
			return err
		}
		if !rp.Serve(c) {
			c.Close()
			time.Sleep(100 * time.Millisecond)
		}
		c = nil
	}
}

var defaultServerName = []byte("swallow server")

func (s *Server) getServerName() []byte {
	if len(s.Name) == 0 {
		return defaultServerName
	}
	return s.Name
}

var globalConnID uint64

func nextConnID() uint64 {
	return atomic.AddUint64(&globalConnID, 1)
}

func (s *Server) serveConn(c net.Conn) error {
	//	serverName := s.getServerName()
	connNum := uint64(0)
	var (
		err error
	)
	for {
		connNum++
		//TODO
		if s.MaxPerConn > 0 && connNum >= uint64(s.MaxPerConn) {
			//	ctx.SetConnectionClose()
		}
	}
	return err
}

func acceptConn(s *Server, ln net.Listener) (net.Conn, error) {
	for {
		c, err := ln.Accept()
		if err != nil {
			if c != nil {
				panic("BUG: net.Listener returned non-nil conn and non-nil error")
			}
			if netErr, ok := err.(net.Error); ok && netErr.Temporary() {
				time.Sleep(time.Second)
				continue
			}
			if err != io.EOF && !strings.Contains(err.Error(), "use of closed network connection") {
				return nil, err
			}
			return nil, io.EOF
		}
		if c == nil {
			panic("BUG: net.Listener returned (nil, nil)")
		}
		if s.MaxConnsPerIP > 0 {
			pic := wrapPerIPConn(s, c)
			if pic == nil {
				continue
			}
			c = pic
		}
		return c, nil
	}
}

func wrapPerIPConn(s *Server, c net.Conn) net.Conn {
	ip := getUint32IP(c)
	if ip == 0 {
		return c
	}
	n := s.perIPConnCounter.Register(ip)
	if n > s.MaxConnsPerIP {
		s.perIPConnCounter.Unregister(ip)
		c.Close()
		return nil
	}
	return acquirePerIPConn(c, ip, &s.perIPConnCounter)
}

// Serve serves incoming connections from the given listener.
func (s *Server) handle() error {
	var (
		c   net.Conn
		err error
	)
	maxWorkersCount := s.getConcurrency()
	wp := &workerPool{
		WorkerFunc:      s.write,
		MaxWorkersCount: maxWorkersCount,
	}
	wp.Start()
	for i := 0; i < 100; i++ {
		b := NewBucket(10)
		go func(b *Bucket) {
			for {
				msg, ok := <-b.writeChan
				if !ok {
					wp.Stop()
					return
				}
				s.buffer = msg.b
				if !wp.Serve(msg.Conn) {
					c.Close()
					time.Sleep(100 * time.Millisecond)
				}
				c = nil
			}
		}(b)
	}
	return err
}
func (s *Server) write(c net.Conn) error {
	return nil
}
