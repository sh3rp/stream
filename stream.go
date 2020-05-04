package stream

//Stream is the contract used for send/receiving packets across a stream
type Stream interface {

	//Send takes a byte array and sends it across the stream
	Send([]byte)

	//Close closes the stream
	Close()

	//IsConnected checks to see if the stream is available for send/receive operations
	//across the stream
	IsConnected() bool

	//ReadBytes is a counter that counts the bytes read from the stream
	ReadBytes() int64

	//WriteBytes is a counter that counts the bytes written to the stream
	WriteBytes() int64
}
