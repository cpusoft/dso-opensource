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

// TYPE_A
type AModel struct {
	// [4]byte
	Address    jsonutil.HexBytes `json:"address"`
	AddressStr string            `json:"addressStr"`
}

func NewAModel(addressStr string, offsetFromStart uint16) (aModel *AModel, newOffsetFromStart uint16, err error) {
	belogs.Debug("NewAModel(): addressStr:", addressStr, "   offsetFromStart:", offsetFromStart)
	address := iputil.IpToDnsFormatByte(addressStr)
	if address == nil {
		return nil, 0, errors.New("Illegal format")
	}
	c := &AModel{
		Address:    address,
		AddressStr: addressStr,
	}
	newOffsetFromStart = offsetFromStart + uint16(len(address))
	belogs.Debug("NewAModel(): addressStr:", addressStr, "   aModel:", jsonutil.MarshalJson(c), "  newOffsetFromStart:", newOffsetFromStart)
	return c, newOffsetFromStart, nil
}

func ParseBytesToAModel(bytess []byte, offsetFromStart uint16) (aModel *AModel, newOffsetFromStart uint16, err error) {
	belogs.Debug("ParseBytesToAModel(): bytess:", convert.PrintBytesOneLine(bytess), "   offsetFromStart:", offsetFromStart)
	if len(bytess) == 0 {
		belogs.Error("ParseBytesToAModel(): bytess is empty, fail:", convert.PrintBytesOneLine(bytess), err)
		return nil, 0, errors.New("the length of address(ipv4) is too short")
	}

	c := &AModel{
		Address:    bytess,
		AddressStr: iputil.DnsFormatToIp(bytess, false),
	}
	newOffsetFromStart = offsetFromStart + uint16(len(bytess))
	belogs.Debug("ParseBytesToAModel(): bytes:", convert.PrintBytesOneLine(bytess),
		"   aModel:", jsonutil.MarshalJson(c), "  newOffsetFromStart:", newOffsetFromStart)
	return c, newOffsetFromStart, nil
}

func (c *AModel) Bytes() []byte {
	wr := bytes.NewBuffer([]byte{})
	binary.Write(wr, binary.BigEndian, c.Address)
	return wr.Bytes()
}
func (c *AModel) Length() uint16 {
	return uint16(len(c.Address))
}
func (c *AModel) ToRrData() string {
	return iputil.DnsFormatToIp(c.Address, true)
}
