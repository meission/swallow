package swallow

type Message struct {
	b []byte
	*Conn
}

type Bucket struct {
	connPool  map[string]*Conn
	writeChan chan *Message
}

func NewBucket(len int) *Bucket {
	return &Bucket{
		connPool: make(map[string]*Conn, len),
	}
}

func (b *Bucket) Set(c *Conn) {
	b.connPool[c.RemoteAddr().String()] = c
}
func (b *Bucket) Write(c *Conn, bs []byte) {
	b.writeChan <- &Message{bs, c}
}
