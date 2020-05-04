package stream

//WithReadBufferSize sets the size of the data buffer used to read/write to the wire; if this
//option is not set, the default is set to 65535
func WithReadBufferSize(size int) func(*socketStream) {
	return func(sc *socketStream) {
		sc.bufferSize = size
	}
}

//WithPacketBufferSize is the size of the write queue buffer; if this option is not set, the
//default is set to 1
func WithPacketBufferSize(size int) func(*socketStream) {
	return func(sc *socketStream) {
		sc.packetQueue = make(chan []byte, size)
	}
}

//WithPacketHandler sets the handler used to pass parsed bytes off the read queue; if this option
//is not set, an error is logged when a packet is read off the stream
func WithPacketHandler(handler func([]byte)) func(*socketStream) {
	return func(sc *socketStream) {
		sc.handler = handler
	}
}

//WithErrorHandler sets the handler used to report errors; if this option is not set, a simple
//fmt.Printf handler is applied to ensure errors can be reported
func WithErrorHandler(f func(StreamError)) func(*socketStream) {
	return func(sc *socketStream) {
		sc.errorHandler = f
	}
}

//WithHeaderCodec sets the StreamHeaderCodec for the stream; if this option is not set, the
//MarkerLengthCodec is used
func WithHeaderCodec(f StreamHeaderCodec) func(*socketStream) {
	return func(sc *socketStream) {
		sc.headerCodec = f
	}
}
