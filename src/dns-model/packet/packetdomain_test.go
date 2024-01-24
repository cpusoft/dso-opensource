package packet

import (
	"fmt"
	"testing"

	"dns-model/packet"
	"github.com/cpusoft/goutil/jsonutil"
)

func TestParseBytesToPacketLabel(t *testing.T) {
	oneLen := []byte{0x05, 0x5f, 0x68, 0x74, 0x74, 0x70, 0x04, 0x5f, 0x74,
		0x63, 0x70, 0x06, 0x64, 0x6e, 0x73, 0x2d, 0x73, 0x64, 0x03, 0x6f, 0x72, 0x67} //, 0x00, 0x05, 0x5f, 0x68, 0x74, 0x74, 0x70}
	pakcetDomain, newOffsetFromStart, err := ParseBytesToPacketDomain(oneLen, 10)
	fmt.Println(jsonutil.MarshalJson(pakcetDomain), newOffsetFromStart, err)
}

func TestNewPacketDomainByAddPacketLabels1(t *testing.T) {
	packetDecompressionLabel := packet.NewPacketDecompressionLabel()
	l1, _ := NewPacketLabel([]byte("www"), 100)
	l2, _ := NewPacketLabel([]byte("baidu"), 104)
	l3, _ := NewPacketLabel([]byte("com"), 110)
	ls := make([]*PacketLabel, 0)
	ls = append(ls, l1)
	ls = append(ls, l2)
	ls = append(ls, l3)
	c, err := NewPacketDomainByAddPacketLabels(ls, packetDecompressionLabel)
	fmt.Println(c, err)
}
func TestParseBytesToPacketDomain(t *testing.T) {
	b := []byte{0x03, 0x6e, 0x73, 0x31, 0xc0, 0x14}
	c, o, er := ParseBytesToPacketDomain(b, 100)
	fmt.Println(c, o, er)
}
