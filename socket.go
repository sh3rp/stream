package stream

import (
	"fmt"
	"net"
)

//DEFAULT_BUFFER_SIZE is the size of the byte packet to read off the stream
var DEFAULT_BUFFER_SIZE = 65535

//socketStream is the TCP socket implementation of the Stream interface
type socketStream struct {
	conn         net.Conn
	bufferSize   int // size of the byte array to read into
	handler      func([]byte)
	errorHandler func(StreamError)
	headerCodec  StreamHeaderCodec
	packetQueue  chan []byte
	readCtl      chan bool // used to cancel the read loop
	writeCtl     chan bool // used to cancel the write loop
	readBytes    int64
	writeBytes   int64
}

//NewSocketStream creates a new stream using the passed net.Conn as the stream source
//and sets the appropriate options passed to the constructor
func NewSocketStream(conn net.Conn, option ...func(*socketStream)) Stream {
	c := &socketStream{
		conn:        conn,
		bufferSize:  DEFAULT_BUFFER_SIZE,
		packetQueue: make(chan []byte, 1), // default the queue to 1
		readCtl:     make(chan bool, 1),
		writeCtl:    make(chan bool, 1),
	}

	for _, opt := range option {
		opt(c)
	}

	if c.errorHandler == nil {
		c.errorHandler = func(e StreamError) {
			fmt.Printf("Error: %+v", e.Error)
		}
	}

	if c.headerCodec == nil {
		c.headerCodec = MarkerLengthCodec{}
	}

	go c.read()
	go c.write()

	return c
}

func (c *socketStream) IsConnected() bool {
	return c.conn != nil
}

func (c *socketStream) Send(pkt []byte) {
	c.packetQueue <- pkt
}

func (c *socketStream) Close() {
	c.writeCtl <- true
	c.readCtl <- true
}

func (c *socketStream) ReadBytes() int64 {
	return c.readBytes
}

func (c *socketStream) WriteBytes() int64 {
	return c.writeBytes
}

func (c *socketStream) write() {
	for {
		select {
		case pkt := <-c.packetQueue:

			if c.conn == nil {
				c.errorHandler(NewStreamErr(STREAM_HANDLER, "Connection not set, cannot write packet"))
				continue
			}

			data := c.headerCodec.Encode(pkt)
			n, err := c.conn.Write(data)

			if err != nil {
				c.errorHandler(NewStreamErr(STREAM_WRITE, fmt.Sprintf("Writing: %v, closing connection", err)))
				c.conn.Close()
				c.writeCtl <- true
			} else {
				c.writeBytes = c.writeBytes + int64(n)
			}

		case <-c.writeCtl:

			if c.conn != nil {
				c.conn.Close()
				c.conn = nil
			}

			return
		}
	}
}

func (c *socketStream) read() {

	if c.conn == nil {
		c.errorHandler(NewStreamErr(STREAM_WRITE, "Connection is nil"))
		return
	}

	if c.handler == nil {
		c.errorHandler(NewStreamErr(STREAM_HANDLER, "No handler set for inbound data"))
		return
	}

	buffer := make([]byte, c.bufferSize)

	done := false

	for {
		select {
		default:
			l, err := c.conn.Read(buffer)

			c.readBytes = c.readBytes + int64(l)

			data := buffer[:l]

			if err != nil {
				if err.Error() != "EOF" {
					c.errorHandler(NewStreamErr(STREAM_READ, fmt.Sprintf("Connection read error: %v", err)))
				} else {
					c.errorHandler(NewStreamErr(STREAM_EOF, fmt.Sprintf("Connection EOF: %v", err)))
				}
				done = true
				break
			}

			pkts, err := c.headerCodec.Decode(data)

			if err != nil {
				c.errorHandler(NewStreamErr(STREAM_MARSHALLING, fmt.Sprintf("Packet marshalling: %+v", err)))
			}

			if c.handler == nil {
				c.errorHandler(NewStreamErr(STREAM_HANDLER, "No handler set for inbound data"))
			} else {
				for _, pkt := range pkts {
					c.handler(pkt)
				}
			}

		case <-c.readCtl:

			c.conn.Close()
			done = true
			c.conn = nil

			break
		}

		if done {
			break
		}
	}
}
