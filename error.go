package stream

import "errors"

//StreamError is an error wrapper that is used to set the context of
//the error.
type StreamError struct {
	Error error
	Type  int
}

//Error contexts
const (
	STREAM_EOF = iota
	STREAM_HANDLER
	STREAM_MARSHALLING
	STREAM_WRITE
	STREAM_READ
)

//NewStreamErr is the constructor for the StreamError
func NewStreamErr(errType int, errStr string) StreamError {
	return StreamError{Error: errors.New(errStr), Type: errType}
}
