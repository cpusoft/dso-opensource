package packet

import (
	"fmt"
	"testing"

	"github.com/cpusoft/goutil/jsonutil"
)

func TestParseBytesToNsModel(t *testing.T) {
	b := []byte{0x04, 0x6e, 0x73, 0x31, 0x35, 0xc0, 0x0c}
	aaaa, newOffsetFromStart, err := ParseBytesToNsModel(b, 100)
	fmt.Println(jsonutil.MarshalJson(aaaa), newOffsetFromStart, err)
}
