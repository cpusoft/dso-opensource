package packet

import (
	"fmt"
	"testing"

	"github.com/cpusoft/goutil/jsonutil"
)

func TestParseBytesToAModel(t *testing.T) {
	b := []byte{0x92, 0x70, 0x3c, 0x35}
	aaaa, newOffsetFromStart, err := ParseBytesToAModel(b, 100)
	fmt.Println(jsonutil.MarshalJson(aaaa), newOffsetFromStart, err)
}
