package packet

import (
	"fmt"
	"testing"

	"github.com/cpusoft/goutil/jsonutil"
)

func TestParseBytesToAaaaModel(t *testing.T) {
	b := []byte{0x2a, 0x04, 0xe4, 0xc0, 0x00, 0x53, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x53}
	aaaa, newOffsetFromStart, err := ParseBytesToAaaaModel(b, 100)
	fmt.Println(jsonutil.MarshalJson(aaaa), newOffsetFromStart, err)
}
