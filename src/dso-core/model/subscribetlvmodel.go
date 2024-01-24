package model

import (
	"bytes"
	"encoding/binary"
	"errors"
	"strings"

	packet "dns-model/packet"
	"dns-model/rr"
	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/convert"
	"github.com/cpusoft/goutil/dnsutil"
	"github.com/cpusoft/goutil/jsonutil"
	"github.com/cpusoft/goutil/osutil"
)

// /////////////////////////////
// subscribe
type SubscribeTlvModel struct {
	DsoType             uint16                  `json:"dsoType"`
	DsoLength           uint16                  `json:"dsoLength"`
	DnsNamePacketDomain *packet.PacketDomain    `json:"dnsNamePacketDomain"`
	DnsName             jsonutil.PrintableBytes `json:"dnsName"`  //  []byte
	DnsType             uint16                  `json:"dnsType"`  // 'type' is keyword in golang, so use dnsType
	DnsClass            uint16                  `json:"dnsClass"` //
}

func NewSubscribeTlvModel(dnsNamePacketDomain *packet.PacketDomain, dnsName []byte,
	dnsType uint16, dnsClass uint16) *SubscribeTlvModel {
	c := &SubscribeTlvModel{
		DsoType:             dnsutil.DSO_TYPE_SUBSCRIBE,
		DsoLength:           uint16(dnsNamePacketDomain.Length() + 2 + 2),
		DnsNamePacketDomain: dnsNamePacketDomain,
		DnsName:             dnsName,
		DnsType:             dnsType,
		DnsClass:            dnsClass,
	}
	return c
}

func NewSubscribeTlvModelByRrModel(rrModel *rr.RrModel) (*SubscribeTlvModel, error) {
	belogs.Debug("NewSubscribeTlvModelByRrModel(): rrModel:", jsonutil.MarshalJson(rrModel))

	offsetFromStart := uint16(0)
	fullDomain := rrModel.RrFullDomain
	if len(fullDomain) == 0 {
		fullDomain = rrModel.RrName + "." + strings.TrimSuffix(rrModel.Origin, ".")
	}
	belogs.Debug("NewSubscribeTlvModelByRrModel(): fullDomain:", fullDomain)

	packetDomain, newOffsetFromStart, err := packet.NewPacketDomainNoCompression([]byte(fullDomain), offsetFromStart)
	if err != nil {
		belogs.Error("NewSubscribeTlvModelByRrModel(): NewPacketDomainNoCompression fail, rrModel:", jsonutil.MarshalJson(rrModel), err)
		return nil, err
	}
	belogs.Debug("NewSubscribeTlvModelByRrModel(): packetDomain:", jsonutil.MarshalJson(packetDomain), "   newOffsetFromStart:", newOffsetFromStart)

	packetType, ok := dnsutil.DnsStrTypes[rrModel.RrType]
	if !ok {
		belogs.Error("NewSubscribeTlvModelByRrModel(): DnsStrTypes fail, RrType:", rrModel.RrType)
		return nil, errors.New("RrType is illegal")
	}
	newOffsetFromStart += 2
	belogs.Debug("NewSubscribeTlvModelByRrModel(): packetType:", packetType, "   newOffsetFromStart:", newOffsetFromStart)

	packetClass, ok := dnsutil.DnsStrClasses[rrModel.RrClass]
	if !ok {
		belogs.Error("NewSubscribeTlvModelByRrModel(): DnsStrClasses fail, RrClass:", rrModel.RrClass)
		return nil, errors.New("RrClass is illegal")
	}
	newOffsetFromStart += 2
	belogs.Debug("NewSubscribeTlvModelByRrModel(): packetClass:", packetClass, "   newOffsetFromStart:", newOffsetFromStart)

	return NewSubscribeTlvModel(packetDomain, []byte(fullDomain),
		packetType, packetClass), nil
}

func NewDsoModelWithSubscribeTlvModel(messageId uint16, rrModel *rr.RrModel) (dsoModel *DsoModel, err error) {
	belogs.Debug("NewDsoModelWithSubscribeTlvModel():messageId:", messageId, " rrModel:", jsonutil.MarshalJson(rrModel))

	// dso header/count:
	dsoModel, err = NewDsoModelByParameters(messageId, dnsutil.DNS_QR_REQUEST, dnsutil.DNS_RCODE_NOERROR)
	if err != nil {
		belogs.Error("NewDsoModelWithSubscribeTlvModel(): NewDsoModelByParameters fail:", err)
		return nil, err
	}
	belogs.Debug("NewDsoModelWithSubscribeTlvModel(): NewDsoModelByParameters, dsoModel:", jsonutil.MarshalJson(dsoModel))

	subscirbeTlvModel, err := NewSubscribeTlvModelByRrModel(rrModel)
	if err != nil {
		belogs.Error("NewDsoModelWithSubscribeTlvModel(): NewSubscribeTlvModelByRrModel fail, rrModel:", jsonutil.MarshalJson(rrModel), err)
		return nil, err
	}
	dsoModel.AddTlvModel(subscirbeTlvModel)
	belogs.Debug("NewDsoModelWithSubscribeTlvModel(): dsoModel:", jsonutil.MarshalJson(dsoModel))
	belogs.Info("#生成DSO的'订阅'类型数据包: " + osutil.GetNewLineSep() +
		"{'域名':'" + rrModel.RrFullDomain + "','Type':'" + rrModel.RrType +
		"','Class':'" + rrModel.RrClass + "'}")
	return dsoModel, nil
}

func ParseBytesToSubscribeTlvModel(dsoLength uint16, subscribeBytes []byte,
	offsetFromStart uint16, packetDecompressionLabel *packet.PacketDecompressionLabel) (tlvModel TlvModel, newOffsetFromStart uint16, err error) {

	if len(subscribeBytes) < dnsutil.DSO_TYPE_SUBSCRIBE_MIN_LENGTH {
		belogs.Error("ParseBytesToSubscribeTlvModel(): recv byte's length is too small, fail: ",
			"   len(subscribeBytes):", len(subscribeBytes), "   DSO_TYPE_SUBSCRIBE_MIN_LENGTH:", dnsutil.DSO_TYPE_SUBSCRIBE_MIN_LENGTH)
		return nil, 0, errors.New("Received packet is too small for legal DSO Subscribe format")
	}
	if len(subscribeBytes) < int(dsoLength) {
		belogs.Error("ParseBytesToSubscribeTlvModel(): recv byte's length is too small, fail: ",
			"   len(subscribeBytes):", len(subscribeBytes), "  dsoLength:", dsoLength)
		return nil, 0, errors.New("Received packet is too small for legal DSO Subscribe format")
	}

	dnsNameLength := dsoLength - 4 // length of type + class
	dnsNameBytes := subscribeBytes[:dnsNameLength]
	dnsNamePacketDomain, newOffsetFromStart, err := packet.ParseBytesToPacketDomain(dnsNameBytes, offsetFromStart, packetDecompressionLabel)
	if err != nil {
		belogs.Error("ParseBytesToNsModel(): ParseBytesToPacketDomain fail, subscribeBytes:", convert.PrintBytesOneLine(subscribeBytes), err)
		return nil, 0, err
	}
	belogs.Debug("ParseBytesToSubscribeTlvModel(): dnsNameLength:", dnsNameLength, "  dnsNameBytes:", convert.PrintBytesOneLine(dnsNameBytes),
		"  dnsNamePacketDomain:", jsonutil.MarshalJson(dnsNamePacketDomain))

	typeClassBytes := subscribeBytes[dnsNameLength : dnsNameLength+4]
	dnsType := binary.BigEndian.Uint16(typeClassBytes[:2])
	rrType := dnsutil.DnsIntTypes[dnsType]
	newOffsetFromStart += 2
	belogs.Debug("ParseBytesToSubscribeTlvModel(): dnsType:", dnsType, " rrType:", rrType, "  newOffsetFromStart:", newOffsetFromStart)

	dnsClass := binary.BigEndian.Uint16(typeClassBytes[2:])
	rrClass := dnsutil.DnsIntClasses[dnsClass]
	newOffsetFromStart += 2
	belogs.Info("ParseBytesToSubscribeTlvModel():  dnsNamePacketDomain:", jsonutil.MarshalJson(dnsNamePacketDomain), "  dnsNameBytes:", string(dnsNameBytes),
		"  dnsType:", dnsType, " rrType:", rrType, " dnsClass:", dnsClass, "  rrClass:", rrClass,
		"  newOffsetFromStart:", newOffsetFromStart)

	belogs.Info("#解析得到DSO的'订阅'类型数据包: " + osutil.GetNewLineSep() +
		"{'域名':'" + string(dnsNamePacketDomain.FullDomain) +
		"','Type':'" + rrType +
		"','Class':'" + rrClass + "'}")
	return NewSubscribeTlvModel(dnsNamePacketDomain, dnsNameBytes, dnsType, dnsClass), newOffsetFromStart, nil
}

func (c SubscribeTlvModel) Bytes() []byte {
	wr := bytes.NewBuffer([]byte{})
	binary.Write(wr, binary.BigEndian, c.DsoType)
	binary.Write(wr, binary.BigEndian, c.DsoLength)
	binary.Write(wr, binary.BigEndian, c.DnsNamePacketDomain.Bytes())
	binary.Write(wr, binary.BigEndian, c.DnsType)
	binary.Write(wr, binary.BigEndian, c.DnsClass)
	return wr.Bytes()
}

func (c SubscribeTlvModel) PrintBytes() string {
	return convert.PrintBytes(c.Bytes(), 8)
}
func (c SubscribeTlvModel) GetDsoType() uint16 {
	return dnsutil.DSO_TYPE_SUBSCRIBE
}
