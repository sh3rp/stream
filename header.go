package stream

import (
	"encoding/binary"
	"errors"
)

//StreamHeaderCodec defines how the header for packets is generated
//and read off the stream.  Since the stream header can be computed
//in a variety of different ways, with each bearing their own merit,
//the StreamHeaderCodec is used for various implementations of the
//header encoding and decoding.
type StreamHeaderCodec interface {

	//Decode takes bytes, either as a single packet or multiple
	//packets packed together and decodes them into an array of
	//byte arrays containing data; if decoding problems occur, an
	// error may be returned
	Decode([]byte) ([][]byte, error)

	//Encode takes a single packet and attaches a header for the
	//purposes of placing the entire payload onto the stream
	Encode([]byte) []byte
}

var MARKER = []byte{0xff, 0xff}

//MarkerLengthCodec used a combination of marker field + length field
//to define packet demarcation.  The initial marker is read first, to
//signal the beginning of a packet; if the marker is not detected, then
//the buffer is assumed to not have any packets in it.  If the marker
//is detected, the length is read as a 16-bit in, with a max packet
//size of 65535.  Since fields are statically defined and no header
//calculations must be made, this is slightly faster than the
//LengthFieldCodec.
type MarkerLengthCodec struct{}

func (ml MarkerLengthCodec) Decode(bytes []byte) ([][]byte, error) {
	var packets [][]byte
	var idx uint16
	idx = 0
	for idx < uint16(len(bytes)) {
		for i := 0; i < len(MARKER); i++ {
			if bytes[idx+uint16(i)] != MARKER[i] {
				return packets, errors.New("Malformed packet")
			}
		}
		idx = idx + 2
		length := binary.LittleEndian.Uint16(bytes[idx : idx+2])
		idx = idx + 2
		data := bytes[idx : idx+length]
		idx = idx + length
		packets = append(packets, data)
	}
	return packets, nil
}

func (ml MarkerLengthCodec) Encode(pkt []byte) []byte {
	var bytes []byte

	bytes = append(bytes, MARKER...)
	lenBytes := make([]byte, 2)
	binary.LittleEndian.PutUint16(lenBytes, uint16(len(pkt)))
	bytes = append(bytes, lenBytes...)
	bytes = append(bytes, pkt...)
	return bytes
}

//LengthFieldCodec use the first byte to pack as much information as
//possible into the header.  The top 2 bits are used to determine
//how many bytes are required to determine the length:
//
//  [00xxxxxx] = low 6 bits are the data length
//  [01xxxxxx] = low 6 bits are the high, byte index 1 is the low
//  [10xxxxxx] = low 6 bits are the high, byte index 1 and 2 are the
//               low
//
//This encoding provides for a maximum packet size of 4,194,304 bytes,
//a number far greater than will likely be necessary.  The size of
//the header makes for a nicely compact message.
type LengthFieldCodec struct{}

func (ml LengthFieldCodec) Decode(bytes []byte) ([][]byte, error) {
	var packets [][]byte

	idx := 0
	for idx < len(bytes) {
		len := getLenFromHeader(bytes[idx:])
		pad := 0
		if len <= 63 {
			pad = idx + 1
		} else if len <= 16383 {
			pad = idx + 2
		} else {
			pad = idx + 3
		}
		packets = append(packets, bytes[pad:pad+len])
		idx = pad + len
	}

	return packets, nil
}
func (ml LengthFieldCodec) Encode(pkt []byte) []byte {
	var data []byte
	data = append(data, getHeaderFromLen(len(pkt))...)
	data = append(data, pkt...)
	return data
}

//getHeaderFromLen gets the header bytes based on the specified length of the
//packet
func getHeaderFromLen(length int) []byte {
	var bytes []byte

	if length <= 63 {
		bytes = []byte{0x3f & uint8(length)}
	} else if length <= 16383 {
		bytes = []byte{64 | uint8(63&(length>>8)), uint8(255 & length)}
	} else {
		bytes = []byte{128 | uint8(63&(length>>16)), uint8(255 & (length >> 8)), uint8(255 & length)}
	}
	return bytes
}

//getLenFromHeader gets the length based on the size of the data packet
func getLenFromHeader(data []byte) int {
	var len uint32
	header := (data[0] & 0xc0) >> 6

	switch header {
	case 0:
		len = uint32(data[0] & 0x3f)
	case 1:
		len = uint32(uint32(data[0]&63)<<8) | uint32(255&data[1])
	case 2:
		len = uint32(uint32(data[0]&63)<<16) + uint32(uint32(data[1])<<8) + uint32(255&data[2])
	}

	return int(len)
}
