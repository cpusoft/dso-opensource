package packet

import (
	"bytes"
	"encoding/binary"
	"errors"
	"strconv"

	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/convert"
	"github.com/cpusoft/goutil/jsonutil"
)

type MxModel struct {
	Preference           uint16                  `json:"preference"`
	ExchangePacketDomain *PacketDomain           `json:"exchangePacketDomain"`
	Exchange             jsonutil.PrintableBytes `json:"exchange"`
}

func NewMxModel(preferenceStr string, exchangeStr string, offsetFromStart uint16) (mxModel *MxModel, newOffsetFromStart uint16, err error) {
	belogs.Debug("NewMxModel(): preferenceStr:", preferenceStr, "   exchangeStr:", exchangeStr, "   offsetFromStart:", offsetFromStart)
	preference, err := strconv.Atoi(preferenceStr)
	if err != nil {
		belogs.Error("NewMxModel(): Atoi preferenceStr fail, preferenceStr:", preferenceStr, err)
		return nil, 0, err
	}

	exchangePacketDomain, newOffsetFromStart, err := NewPacketDomainNoCompression([]byte(exchangeStr), offsetFromStart)
	if err != nil {
		belogs.Error("NewMxModel(): NewPacketDomainNoCompression fail, exchangeStr:", exchangeStr, err)
		return nil, 0, err
	}
	c := &MxModel{
		Preference:           uint16(preference),
		ExchangePacketDomain: exchangePacketDomain,
		Exchange:             []byte(exchangeStr),
	}
	belogs.Debug("NewMxModel(): preferenceStr:", preferenceStr, "  exchangeStr:", exchangeStr,
		"   mxModel:", jsonutil.MarshalJson(c), "   newOffsetFromStart:", newOffsetFromStart)
	return c, newOffsetFromStart, nil
}
func ParseBytesToMxModel(bytess []byte, offsetFromStart uint16, packetDecompressionLabel *PacketDecompressionLabel) (mxModel *MxModel, newOffsetFromStart uint16, err error) {
	belogs.Debug("ParseBytesToMxModel(): bytess:", convert.PrintBytesOneLine(bytess), "   offsetFromStart:", offsetFromStart)
	if len(bytess) < 3 {
		belogs.Error("ParseBytesToMxModel(): len(bytess) < 3, fail:", convert.PrintBytesOneLine(bytess), err)
		return nil, 0, errors.New("the length of mx is too short")
	}

	preference := binary.BigEndian.Uint16(bytess[:2])
	newOffsetFromStart = offsetFromStart + 2
	belogs.Debug("ParseBytesToMxModel(): preference:", preference)

	exchange := bytess[2:]
	exchangePacketDomain, newOffsetFromStart, err := ParseBytesToPacketDomain(exchange, newOffsetFromStart, packetDecompressionLabel)
	if err != nil {
		belogs.Error("ParseBytesToCNameModel(): ParseBytesToPacketDomain fail, bytess:", convert.PrintBytesOneLine(bytess), err)
		return nil, 0, err
	}
	c := &MxModel{
		Preference:           preference,
		ExchangePacketDomain: exchangePacketDomain,
		Exchange:             exchange,
	}

	belogs.Debug("ParseBytesToMxModel(): bytes:", convert.PrintBytesOneLine(bytess),
		"   mxModel:", jsonutil.MarshalJson(c),
		"   newOffsetFromStart:", newOffsetFromStart)
	return c, newOffsetFromStart, nil
}

func (c *MxModel) Bytes() []byte {
	wr := bytes.NewBuffer([]byte{})
	binary.Write(wr, binary.BigEndian, c.Preference)
	binary.Write(wr, binary.BigEndian, c.ExchangePacketDomain.Bytes())
	return wr.Bytes()
}
func (c *MxModel) Length() uint16 {
	return uint16(2 + c.ExchangePacketDomain.Length())
}
func (c *MxModel) ToRrData() string {
	return convert.ToString(c.Preference) + " " + string(c.ExchangePacketDomain.FullDomain)
}
