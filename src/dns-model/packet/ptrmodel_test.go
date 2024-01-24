package packet

import (
	"fmt"
	"testing"

	"github.com/cpusoft/goutil/jsonutil"
)

func TestParseBytesToPtrModel(t *testing.T) {
	b := []byte{0x29, 0x20, 0x2a, 0x20, 0x59, 0x61, 0x68, 0x6f, 0x6f, 0x2c, 0x20, 0x6d, 0x61, 0x70, 0x73, 0x2c, 0x20, 0x77, 0x65, 0x61, 0x74, 0x68, 0x65, 0x72, 0x2c, 0x20, 0x61, 0x6e, 0x64, 0x20, 0x73, 0x74, 0x6f, 0x63, 0x6b, 0x20, 0x71, 0x75, 0x6f, 0x74, 0x65, 0x73, 0xc0, 0x0c}
	aaaa, newOffsetFromStart, err := ParseBytesToPtrModel(b, 100)
	fmt.Println(jsonutil.MarshalJson(aaaa), newOffsetFromStart, err)
}
