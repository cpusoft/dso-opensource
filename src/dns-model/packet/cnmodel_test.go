package packet

import (
	"fmt"
	"testing"

	"github.com/cpusoft/goutil/jsonutil"
)

func TestParseBytesToCNameModel(t *testing.T) {
	b := []byte{0x03, 0x77, 0x77, 0x77, 0x01, 0x61, 0x06, 0x73, 0x68, 0x69, 0x66, 0x65, 0x6e, 0xc0, 0x16}
	aaaa, newOffsetFromStart, err := ParseBytesToCNameModel(b, 100)
	fmt.Println(jsonutil.MarshalJson(aaaa), newOffsetFromStart, err)
}
