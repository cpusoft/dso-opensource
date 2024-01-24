package packet

import (
	"errors"

	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/convert"
	"github.com/cpusoft/goutil/jsonutil"
)

type NsModel struct {
	NsDNamePacketDomain *PacketDomain           `json:"nsDNamePacketDomain"`
	NsDName             jsonutil.PrintableBytes `json:"cName"`
}

func NewNsModel(nsDName string, offsetFromStart uint16) (nsModel *NsModel, newOffsetFromStart uint16, err error) {
	belogs.Debug("NewNsModel(): nsDName:", nsDName, "  offsetFromStart:", offsetFromStart)

	nsDNamePacketDomain, newOffsetFromStart, err := NewPacketDomainNoCompression([]byte(nsDName), offsetFromStart)
	if err != nil {
		belogs.Error("NewNsModel(): NewPacketDomainNoCompression fail, nsDName:", nsDName, err)
		return nil, 0, err
	}
	c := &NsModel{
		NsDNamePacketDomain: nsDNamePacketDomain,
		NsDName:             []byte(nsDName),
	}
	belogs.Debug("NewNsModel(): nsDName:", nsDName, "   nsModel:", jsonutil.MarshalJson(c), "  newOffsetFromStart:", newOffsetFromStart)
	return c, newOffsetFromStart, nil
}

func ParseBytesToNsModel(bytess []byte, offsetFromStart uint16,
	packetDecompressionLabel *PacketDecompressionLabel) (nsModel *NsModel, newOffsetFromStart uint16, err error) {
	belogs.Debug("ParseBytesToNsModel(): bytess:", convert.PrintBytesOneLine(bytess), "   offsetFromStart:", offsetFromStart)
	if len(bytess) < 2 {
		belogs.Error("ParseBytesToNsModel(): len(bytess) < 2, fail:", convert.PrintBytesOneLine(bytess), err)
		return nil, 0, errors.New("the length of ns is too short")
	}

	nsDNamePacketDomain, newOffsetFromStart, err := ParseBytesToPacketDomain(bytess, offsetFromStart, packetDecompressionLabel)
	if err != nil {
		belogs.Error("ParseBytesToNsModel(): ParseBytesToPacketDomain fail, bytess:", convert.PrintBytesOneLine(bytess), err)
		return nil, 0, err
	}
	c := &NsModel{
		NsDNamePacketDomain: nsDNamePacketDomain,
		NsDName:             bytess,
	}
	belogs.Debug("ParseBytesToNsModel(): bytess:", convert.PrintBytesOneLine(bytess),
		"   nsModel:", jsonutil.MarshalJson(c),
		"   newOffsetFromStart:", newOffsetFromStart)
	return c, newOffsetFromStart, nil
}

func (c *NsModel) Bytes() []byte {
	return c.NsDNamePacketDomain.Bytes()
}
func (c *NsModel) Length() uint16 {
	return c.NsDNamePacketDomain.Length()
}
func (c *NsModel) ToRrData() string {
	return string(c.NsDNamePacketDomain.FullDomain)
}
