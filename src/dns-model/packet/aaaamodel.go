package packet

import (
	"bytes"
	"encoding/binary"
	"errors"

	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/convert"
	"github.com/cpusoft/goutil/iputil"
	"github.com/cpusoft/goutil/jsonutil"
)

type AaaaModel struct {
	// [16]byte
	Address    jsonutil.HexBytes `json:"address"`
	AddressStr string            `json:"addressStr"`
}

func NewAaaaModel(addressStr string, offsetFromStart uint16) (aaaaModel *AaaaModel, newOffsetFromStart uint16, err error) {
	belogs.Debug("NewAaaaModel(): addressStr:", addressStr, "   offsetFromStart:", offsetFromStart)
	address := iputil.IpToDnsFormatByte(addressStr)
	if address == nil {
		return nil, 0, errors.New("Illegal format")
	}
	c := &AaaaModel{
		Address:    address,
		AddressStr: addressStr,
	}
	newOffsetFromStart = offsetFromStart + uint16(len(address))
	belogs.Debug("NewAaaaModel(): addressStr:", addressStr, "   aaaaModel:", jsonutil.MarshalJson(c), "   newOffsetFromStart:", newOffsetFromStart)
	return c, newOffsetFromStart, nil
}

func ParseBytesToAaaaModel(bytess []byte, offsetFromStart uint16) (aaaaModel *AaaaModel, newOffsetFromStart uint16, err error) {
	belogs.Debug("ParseBytesToAaaaModel(): bytess:", convert.PrintBytesOneLine(bytess), "   offsetFromStart:", offsetFromStart)
	if len(bytess) == 0 {
		belogs.Error("ParseBytesToAaaaModel(): bytess is empty, fail:", convert.PrintBytesOneLine(bytess), err)
		return nil, 0, errors.New("the length of address(ipv6) is too short")
	}
	c := &AaaaModel{
		Address:    bytess,
		AddressStr: iputil.DnsFormatToIp(bytess, false),
	}
	newOffsetFromStart = offsetFromStart + uint16(len(bytess))
	belogs.Debug("ParseBytesToAaaaModel(): bytes:", convert.PrintBytesOneLine(bytess),
		"   aaaaModel:", jsonutil.MarshalJson(c), "  newOffsetFromStart:", newOffsetFromStart)
	return c, newOffsetFromStart, nil
}

func (c *AaaaModel) Bytes() []byte {
	wr := bytes.NewBuffer([]byte{})
	binary.Write(wr, binary.BigEndian, c.Address)
	return wr.Bytes()
}

func (c *AaaaModel) Length() uint16 {
	return uint16(len(c.Address))
}
func (c *AaaaModel) ToRrData() string {
	return iputil.DnsFormatToIp(c.Address, true)
}
