package model

import (
	"bytes"
	"encoding/binary"
	"errors"

	packet "dns-model/packet"
	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/convert"
	"github.com/cpusoft/goutil/dnsutil"
	"github.com/cpusoft/goutil/jsonutil"
)

type ZoneModel struct {
	ZNamePacketDomain *packet.PacketDomain `json:"zNamePacketDomain"`
	ZName             jsonutil.HexBytes    `json:"zName"`
	ZType             uint16               `json:"zType"`  // SOA
	ZClass            uint16               `json:"zClass"` // IN
}

func NewZoneModel(zName string, offsetFromStart uint16) (zoneModel *ZoneModel, newOffsetFromStart uint16, err error) {
	zNamePacketDomain, newOffsetFromStart, err := packet.NewPacketDomainNoCompression([]byte(zName), offsetFromStart)
	if err != nil {
		belogs.Error("NewZoneModel(): NewPacketDomainNoCompression fail, zName:", zName, err)
		return nil, 0, err
	}
	c := &ZoneModel{
		ZNamePacketDomain: zNamePacketDomain,
		ZName:             []byte(zName),
		ZType:             dnsutil.DNS_TYPE_INT_SOA,
		ZClass:            dnsutil.DNS_CLASS_INT_IN,
	}
	belogs.Debug("NewZoneModel(): zName:", zName, "   zoneModel:", jsonutil.MarshalJson(c), "  newOffsetFromStart:", newOffsetFromStart)
	return c, newOffsetFromStart, nil
}

// currently not support compression format
func ParseBytesToZoneModel(bytess []byte, offsetFromStart uint16,
	packetDecompressionLabel *packet.PacketDecompressionLabel) (zoneModel *ZoneModel, newOffsetFromStart uint16, err error) {
	numbersLen := 2 * 2 // 2*int16
	if len(bytess) < numbersLen {
		belogs.Error("ParseBytesToZoneModel(): len(bytess):", len(bytess), "   numbersLen:", numbersLen)
		return nil, 0, errors.New("Illegal Zone format")
	}

	zNamePacketDomain, newOffsetFromStart, err := packet.ParseBytesToPacketDomain(bytess, offsetFromStart, packetDecompressionLabel)
	if err != nil {
		belogs.Error("ParseBytesToZoneModel(): ParseBytesToPacketDomain fail, bytess:", convert.PrintBytesOneLine(bytess), err)
		return nil, 0, err
	}

	//
	zNameLen := newOffsetFromStart - offsetFromStart
	numberBytes := bytess[zNameLen : zNameLen+4] // 2*uint16
	belogs.Debug("ParseBytesToZoneModel(): zNameLen:", zNameLen, "  numberBytes:", convert.PrintBytesOneLine(numberBytes))

	// numbers
	zType := binary.BigEndian.Uint16(numberBytes[:2])
	if zType != dnsutil.DNS_TYPE_INT_SOA {
		belogs.Error("ParseBytesToZoneModel():zType is not SOA fail, zType:", zType)
		return nil, 0, errors.New("ZType is not SOA")
	}
	newOffsetFromStart += 2
	belogs.Debug("ParseBytesToZoneModel(): zType:", zType, "  newOffsetFromStart:", newOffsetFromStart)

	zClass := binary.BigEndian.Uint16(numberBytes[2:])
	if zClass != dnsutil.DNS_CLASS_INT_IN {
		belogs.Error("ParseBytesToZoneModel():zClass is not IN fail, zClass:", zClass)
		return nil, 0, errors.New("ZClass is not IN")
	}
	newOffsetFromStart += 2
	belogs.Debug("ParseBytesToZoneModel(): zClass:", zClass, "  newOffsetFromStart:", newOffsetFromStart)

	c := &ZoneModel{
		ZNamePacketDomain: zNamePacketDomain,
		ZName:             bytess[:zNameLen],
		ZType:             zType,
		ZClass:            zClass,
	}

	belogs.Debug("ParseBytesToZoneModel(): bytess:", convert.PrintBytesOneLine(bytess),
		"   zoneModel:", jsonutil.MarshalJson(c), "  newOffsetFromStart:", newOffsetFromStart)
	return c, newOffsetFromStart, nil
}

func (c *ZoneModel) Length() uint16 {
	// type(2)+class(2)+ttl(4)+rdlen(2)
	return uint16(len(c.ZName) + 2 + 2)
}
func (c *ZoneModel) Bytes() []byte {
	wr := bytes.NewBuffer([]byte{})
	binary.Write(wr, binary.BigEndian, c.ZNamePacketDomain.Bytes())
	binary.Write(wr, binary.BigEndian, c.ZType)
	binary.Write(wr, binary.BigEndian, c.ZClass)
	return wr.Bytes()
}
