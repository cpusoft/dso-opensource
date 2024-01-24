package packet

import (
	"errors"

	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/convert"
	"github.com/cpusoft/goutil/jsonutil"
)

type PtrModel struct {
	PtrDNamePacketDomain *PacketDomain           `json:"ptrDNamePacketDomain"`
	PtrDName             jsonutil.PrintableBytes `json:"ptrDName"`
}

func NewPtrModel(ptrDName string, offsetFromStart uint16) (ptrModel *PtrModel, newOffsetFromStart uint16, err error) {
	belogs.Debug("NewCNameModel(): ptrDName:", ptrDName, "   offsetFromStart:", offsetFromStart)

	ptrDNamePacketDomain, newOffsetFromStart, err := NewPacketDomainNoCompression([]byte(ptrDName), offsetFromStart)
	if err != nil {
		belogs.Error("NewPtrModel(): ParseBytesToPacketDomain fail, ptrDName:", ptrDName, err)
		return nil, 0, err
	}
	c := &PtrModel{
		PtrDNamePacketDomain: ptrDNamePacketDomain,
		PtrDName:             []byte(ptrDName),
	}
	belogs.Debug("NewPtrModel(): ptrDName:", ptrDName, "   ptrModel:", jsonutil.MarshalJson(c), "   newOffsetFromStart:", newOffsetFromStart)
	return c, newOffsetFromStart, nil
}

func ParseBytesToPtrModel(bytess []byte, offsetFromStart uint16, packetDecompressionLabel *PacketDecompressionLabel) (ptrModel *PtrModel, newOffsetFromStart uint16, err error) {
	belogs.Debug("ParseBytesToPtrModel(): bytess:", convert.PrintBytesOneLine(bytess), "   offsetFromStart:", offsetFromStart)
	if len(bytess) < 2 {
		belogs.Error("ParseBytesToPtrModel(): len(bytess) < 2, fail:", convert.PrintBytesOneLine(bytess), err)
		return nil, 0, errors.New("the length of ptr is too short")
	}
	ptrDNamePacketDomain, newOffsetFromStart, err := ParseBytesToPacketDomain(bytess, offsetFromStart, packetDecompressionLabel)
	if err != nil {
		belogs.Error("ParseBytesToPtrModel():ParseBytesToPacketDomain fail, bytess:", convert.PrintBytesOneLine(bytess), err)
		return nil, 0, err
	}
	c := &PtrModel{
		PtrDNamePacketDomain: ptrDNamePacketDomain,
		PtrDName:             bytess,
	}
	belogs.Debug("ParseBytesToPtrModel(): bytess:", convert.PrintBytesOneLine(bytess),
		"   ptrModel:", jsonutil.MarshalJson(c),
		"   newOffsetFromStart:", newOffsetFromStart)
	return c, newOffsetFromStart, nil
}
func (c *PtrModel) Bytes() []byte {
	return c.PtrDNamePacketDomain.Bytes()
}
func (c *PtrModel) Length() uint16 {
	return c.PtrDNamePacketDomain.Length()
}
func (c *PtrModel) ToRrData() string {
	return string(c.PtrDNamePacketDomain.FullDomain)
}
