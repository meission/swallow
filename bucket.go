package swallow

type Message struct {
	key string
	b   []byte
}

type Bucket struct {
	// key:hash(ip:port:XXX)
	connPool  map[string]*Conn
	writeChan chan *Conn
	readChan  chan *Message
}

func NewBucket(len int) *Bucket {
	return &Bucket{
		connPool:  make(map[string]*Conn, len),
		writeChan: make(chan *Conn, len),
		readChan:  make(chan *Message, len),
	}
}

func (b *Bucket) Set(c *Conn) {
	b.connPool[c.RemoteAddr().String()] = c
}
func (b *Bucket) Write(key string, bs []byte) {
	c, ok := b.connPool[key]
	if ok {
		c.buffer = bs
	}
	b.writeChan <- c
}
