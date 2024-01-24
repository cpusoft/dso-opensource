package packet

import "github.com/cpusoft/goutil/jsonutil"

type PacketDecompressionLabel struct {
	PointerLabels map[uint16]jsonutil.PrintableBytes `json:"pointerLabels"`
}

func NewPacketDecompressionLabel() *PacketDecompressionLabel {
	c := &PacketDecompressionLabel{}
	c.PointerLabels = make(map[uint16]jsonutil.PrintableBytes, 0)
	return c
}
func (c *PacketDecompressionLabel) Add(offsetFromStart uint16, label jsonutil.PrintableBytes) {
	c.PointerLabels[offsetFromStart] = label
}
func (c *PacketDecompressionLabel) Find(pointer uint16) jsonutil.PrintableBytes {
	v, ok := c.PointerLabels[pointer]
	if ok {
		return v
	}
	return nil
}
