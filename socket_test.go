package stream

import (
	"bytes"
	"math/rand"
	"net"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

var bytes10 = generateBytes(10)
var bytes100 = generateBytes(100)
var bytes1000 = generateBytes(1000)

func setup(sHandler, cHandler *testHandler) (Stream, Stream) {
	s, c := net.Pipe()
	sConn := NewSocketStream(s, WithPacketHandler(sHandler.handle))
	cConn := NewSocketStream(c, WithPacketHandler(cHandler.handle))
	return sConn, cConn
}

func setupWithLengthEnc(sHandler, cHandler *testHandler, errHandler *errorHandler) (Stream, Stream) {
	s, c := net.Pipe()
	sConn := NewSocketStream(s, WithHeaderCodec(LengthFieldCodec{}), WithPacketHandler(sHandler.handle), WithErrorHandler(errHandler.handle))
	cConn := NewSocketStream(c, WithHeaderCodec(LengthFieldCodec{}), WithPacketHandler(cHandler.handle), WithErrorHandler(errHandler.handle))
	return sConn, cConn
}

func generateBytes(numBytes int) []byte {
	bytes := make([]byte, numBytes)
	for i := 0; i < numBytes; i++ {
		bytes[i] = uint8(rand.Intn(254))
	}
	return bytes
}

func TestSendRecv1(t *testing.T) {
	// setup
	clientHandler := &testHandler{t: t}
	errHandler := &errorHandler{t: t}
	wg := &sync.WaitGroup{}
	wg.Add(1)
	clientHandler.wg = wg
	sConn, _ := setupWithLengthEnc(&testHandler{}, clientHandler, errHandler)

	// end setup
	data := generateBytes(10)
	sConn.Send(data)
	wg.Wait()
	assert.Equal(t, 1, len(clientHandler.packets))
	assert.True(t, bytes.Equal(data, clientHandler.packets[0]))
}

func TestSendRecv10000(t *testing.T) {
	iterations := 10000
	// setup
	clientHandler := &testHandler{t: t}
	errHandler := &errorHandler{t: t}
	wg := &sync.WaitGroup{}
	wg.Add(iterations)
	clientHandler.wg = wg
	sConn, _ := setupWithLengthEnc(&testHandler{}, clientHandler, errHandler)

	// end setup
	for i := 0; i < iterations; i++ {
		data := generateBytes(10)
		sConn.Send(data)
	}

	wg.Wait()

	assert.Equal(t, iterations, len(clientHandler.packets))
}

func BenchmarkMetricMarkerLength10BytePacket(b *testing.B) {
	wg := &sync.WaitGroup{}
	sConn, _ := setup(&testHandler{}, &testHandler{wg: wg})
	for i := 0; i < b.N; i++ {
		wg.Add(1)
		sConn.Send(bytes10)
		wg.Wait()
	}
}
func BenchmarkMetricMarkerLength100BytePacket(b *testing.B) {
	wg := &sync.WaitGroup{}
	sConn, _ := setup(&testHandler{}, &testHandler{wg: wg})
	for i := 0; i < b.N; i++ {
		wg.Add(1)
		sConn.Send(bytes100)
		wg.Wait()
	}
}
func BenchmarkMetricMarkerLength1000BytePacket(b *testing.B) {
	wg := &sync.WaitGroup{}
	sConn, _ := setup(&testHandler{}, &testHandler{wg: wg})
	for i := 0; i < b.N; i++ {
		wg.Add(1)
		sConn.Send(bytes1000)
		wg.Wait()
	}
}
func BenchmarkMetricLength10BytePacket(b *testing.B) {
	wg := &sync.WaitGroup{}
	sConn, _ := setupWithLengthEnc(&testHandler{}, &testHandler{wg: wg}, &errorHandler{})
	for i := 0; i < b.N; i++ {
		wg.Add(1)
		sConn.Send(bytes10)
		wg.Wait()
	}
}
func BenchmarkMetricLength100BytePacket(b *testing.B) {
	wg := &sync.WaitGroup{}
	sConn, _ := setupWithLengthEnc(&testHandler{}, &testHandler{wg: wg}, &errorHandler{})
	for i := 0; i < b.N; i++ {
		wg.Add(1)
		sConn.Send(bytes100)
		wg.Wait()
	}
}
func BenchmarkMetricLength1000BytePacket(b *testing.B) {
	wg := &sync.WaitGroup{}
	sConn, _ := setupWithLengthEnc(&testHandler{}, &testHandler{wg: wg}, &errorHandler{})
	for i := 0; i < b.N; i++ {
		wg.Add(1)
		sConn.Send(bytes1000)
		wg.Wait()
	}
}

type testHandler struct {
	packets [][]byte
	t       *testing.T
	b       *testing.B
	wg      *sync.WaitGroup
}

func (th *testHandler) handle(bytes []byte) {
	th.packets = append(th.packets, bytes)
	th.wg.Done()
}

type errorHandler struct {
	errors []StreamError
	t      *testing.T
}

func (eh *errorHandler) handle(err StreamError) {
	eh.t.Logf("Error: %+v", err)
	eh.errors = append(eh.errors, err)
}
