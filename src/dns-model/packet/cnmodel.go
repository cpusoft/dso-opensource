package packet

import (
	"errors"

	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/convert"
	"github.com/cpusoft/goutil/jsonutil"
)

type CNameModel struct {
	CNamePacketDomain *PacketDomain           `json:"cNamePacketDomain"`
	CName             jsonutil.PrintableBytes `json:"cName"`
}

func NewCNameModel(cName string, offsetFromStart uint16) (cNameModel *CNameModel, newOffsetFromStart uint16, err error) {
	belogs.Debug("NewCNameModel(): cName:", cName, "   offsetFromStart:", offsetFromStart)
	cNamePacketDomain, newOffsetFromStart, err := NewPacketDomainNoCompression([]byte(cName), offsetFromStart)
	if err != nil {
		belogs.Error("NewCNameModel(): NewPacketDomainNoCompression fail, cName:", cName, err)
		return nil, 0, err
	}
	c := &CNameModel{
		CNamePacketDomain: cNamePacketDomain,
		CName:             []byte(cName),
	}
	belogs.Debug("NewCNameModel(): cName:", cName, "   cNameModel:", jsonutil.MarshalJson(c),
		"   newOffsetFromStart:", newOffsetFromStart)
	return c, newOffsetFromStart, nil
}

func ParseBytesToCNameModel(bytess []byte, offsetFromStart uint16, packetDecompressionLabel *PacketDecompressionLabel) (cNameModel *CNameModel, newOffsetFromStart uint16, err error) {
	belogs.Debug("ParseBytesToCNameModel(): bytess:", convert.PrintBytesOneLine(bytess), "   offsetFromStart:", offsetFromStart)
	if len(bytess) < 2 {
		belogs.Error("ParseBytesToCNameModel(): len(bytess) < 2, fail:", convert.PrintBytesOneLine(bytess), err)
		return nil, 0, errors.New("the length of cname is too short")
	}

	cNamePacketDomain, newOffsetFromStart, err := ParseBytesToPacketDomain(bytess, offsetFromStart, packetDecompressionLabel)
	if err != nil {
		belogs.Error("ParseBytesToCNameModel():ParseBytesToPacketDomain fail, bytess:", convert.PrintBytesOneLine(bytess), err)
		return nil, 0, err
	}
	c := &CNameModel{
		CNamePacketDomain: cNamePacketDomain,
		CName:             bytess,
	}
	belogs.Debug("ParseBytesToCNameModel(): bytes:", convert.PrintBytesOneLine(bytess),
		"   cNameModel:", jsonutil.MarshalJson(c),
		"   newOffsetFromStart:", newOffsetFromStart)
	return c, newOffsetFromStart, nil
}

func (c *CNameModel) Bytes() []byte {
	return c.CNamePacketDomain.Bytes()
}
func (c *CNameModel) Length() uint16 {
	return c.CNamePacketDomain.Length()
}
func (c *CNameModel) ToRrData() string {
	return string(c.CNamePacketDomain.FullDomain)
}
