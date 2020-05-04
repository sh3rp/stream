package stream

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMarkerLengthCodec(t *testing.T) {

}
func TestLengthFieldCodec63(t *testing.T) {
	for i := 0; i < int(math.Pow(2, 6))-1; i++ {
		bytes := getHeaderFromLen(i)
		assert.Equal(t, 1, len(bytes))
		len := getLenFromHeader(bytes)
		assert.Equal(t, i, len)
	}
}

func TestLengthFieldCodec16838(t *testing.T) {
	for i := int(math.Pow(2, 6)); i < int(math.Pow(2, 14))-1; i++ {
		bytes := getHeaderFromLen(i)
		assert.Equal(t, 2, len(bytes))
		len := getLenFromHeader(bytes)
		assert.Equal(t, i, len)
	}
}

func TestLengthFieldCodec4194304(t *testing.T) {
	a := float64(14)
	for i := int(math.Pow(2, a)); i < int(math.Pow(2, 22))-1; i = int(math.Pow(2, a)) {
		bytes := getHeaderFromLen(i)
		assert.Equal(t, 3, len(bytes))
		length := getLenFromHeader(bytes)
		assert.Equal(t, i, length)

		bytes = getHeaderFromLen(i + 1)
		assert.Equal(t, 3, len(bytes))
		length = getLenFromHeader(bytes)
		assert.Equal(t, i+1, length)
		a++
	}
}
