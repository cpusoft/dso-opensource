package packet

import (
	"fmt"
	"testing"

	"github.com/cpusoft/goutil/jsonutil"
)

func TestParseBytesToPacketModelA(t *testing.T) {
	b := []byte{0xc0, 0x0c, 0x00, 0x01, 0x00, 0x01, 0x00, 0x00, 0x01, 0xf2, 0x00, 0x04, 0x0e, 0x12, 0xb4, 0x71}
	ms, nl, err := ParseBytesToPacketModels(b, 0, 100, 0, false)
	fmt.Println(jsonutil.MarshalJson(ms), nl, err)
}

func TestParseBytesToPacketModelsNS(t *testing.T) {
	b := []byte{0xc0, 0x0f, 0x00, 0x02, 0x00, 0x01, 0x00, 0x00, 0x0e, 0xc0, 0x00, 0x0a, 0x07, 0x6e, 0x73, 0x2d, 0x63, 0x6e, 0x63, 0x31, 0xc0, 0x15, 0xc0, 0x0f, 0x00, 0x02, 0x00, 0x01, 0x00, 0x00, 0x0e, 0xc0, 0x00, 0x0a, 0x07, 0x6e, 0x73, 0x2d, 0x63, 0x6d, 0x6e, 0x31, 0xc0, 0x15, 0xc0, 0x0f, 0x00, 0x02, 0x00, 0x01, 0x00, 0x00, 0x0e, 0xc0, 0x00, 0x0a, 0x07, 0x6e, 0x73, 0x2d, 0x74, 0x65, 0x6c, 0x31, 0xc0, 0x15, 0xc0, 0x0f, 0x00, 0x02, 0x00, 0x01, 0x00, 0x00, 0x0e, 0xc0, 0x00, 0x09, 0x06, 0x6e, 0x73, 0x2d, 0x6f, 0x73, 0x31, 0xc0, 0x15}
	ms, nl, err := ParseBytesToPacketModels(b, 0, 100, 0, false)
	fmt.Println(jsonutil.MarshalJson(ms), nl, err)
}
