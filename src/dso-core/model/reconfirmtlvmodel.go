package model

import (
	"bytes"
	"encoding/binary"
	"errors"

	"dns-model/packet"
	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/convert"
	"github.com/cpusoft/goutil/dnsutil"
	"github.com/cpusoft/goutil/jsonutil"
)

///////////////////////////////
// reconfirm
type ReconfirmTlvModel struct {
	DsoType   uint16 `json:"dsoType"`
	DsoLength uint16 `json:"dsoLength"`

	// packetModelType
	DnsPacketModel *packet.PacketModel `json:"dnsPacketModel"`
}

func NewReconfirmTlvModel(dnsPacketModel *packet.PacketModel) *ReconfirmTlvModel {
	c := &ReconfirmTlvModel{
		DsoType:        dnsutil.DSO_TYPE_SUBSCRIBE,
		DsoLength:      dnsPacketModel.Length(),
		DnsPacketModel: dnsPacketModel,
	}
	belogs.Info("#生成DSO的'再次确认'类型数据包:")

	return c
}

func ParseBytesToReconfirmTlvModel(dsoLength uint16, reconfirmBytes []byte,
	offsetFromStart uint16, packetDecompressionLabel *packet.PacketDecompressionLabel) (tlvModel TlvModel, newOffsetFromStart uint16, err error) {
	if len(reconfirmBytes) < dnsutil.DSO_TYPE_RECONFIRM_MIN_LENGTH {
		belogs.Error("ParseBytesToReconfirmTlvModel(): recv byte's length is too small, fail:",
			" len(reconfirmBytes):", len(reconfirmBytes), "   DSO_TYPE_RECONFIRM_MIN_LENGTH:", dnsutil.DSO_TYPE_RECONFIRM_MIN_LENGTH)
		return nil, 0, errors.New("Received packet is too small for legal DSO Reconfirm format")
	}
	if len(reconfirmBytes) < int(dsoLength) {
		belogs.Error("ParseBytesToReconfirmTlvModel(): recv byte's length is too small, fail:",
			" len(reconfirmBytes):", len(reconfirmBytes), "  dsoLength:", dsoLength)
		return nil, 0, errors.New("Received packet is too small for legal DSO Push format")
	}
	belogs.Debug("ParseBytesToReconfirmTlvModel():dsoLength:", dsoLength, "   len(reconfirmBytes):", len(reconfirmBytes))

	// packetmodeltype = true
	dnsPacketModels, newOffsetFromStart, err := packet.ParseBytesToPacketModels(reconfirmBytes, dsoLength, offsetFromStart, packet.DNS_PACKET_NAME_TYPE_CLASS_RDATA, 0, packetDecompressionLabel)
	if err != nil {
		belogs.Error("ParseBytesToPushTlvModel(): ParseBytesToPacketModels fail,len(reconfirmBytes): ", len(reconfirmBytes))
		return nil, 0, errors.New("Received packet is illegal for legal DSO Push format:" + err.Error())
	}

	belogs.Info("ParseBytesToReconfirmTlvModel():dnsPacketModels:", jsonutil.MarshalJson(dnsPacketModels),
		"  newOffsetFromStart:", newOffsetFromStart)
	if len(dnsPacketModels) == 0 {
		belogs.Error("ParseBytesToPushTlvModel(): len(dnsPacketModels) == 0 fail: reconfirmBytes:", convert.PrintBytesOneLine(reconfirmBytes))
		return nil, 0, errors.New("Received packet is illegal for legal DSO Push format:" + err.Error())
	}
	belogs.Info("#解析得到DSO的'再次确认'类型数据包:")
	return NewReconfirmTlvModel(dnsPacketModels[0]), newOffsetFromStart, nil

}

func (c ReconfirmTlvModel) Bytes() []byte {
	wr := bytes.NewBuffer([]byte{})
	binary.Write(wr, binary.BigEndian, c.DsoType)
	binary.Write(wr, binary.BigEndian, c.DsoLength)
	binary.Write(wr, binary.BigEndian, c.DnsPacketModel.Bytes())
	return wr.Bytes()
}

func (c ReconfirmTlvModel) PrintBytes() string {
	return convert.PrintBytes(c.Bytes(), 8)
}
func (c ReconfirmTlvModel) GetDsoType() uint16 {
	return dnsutil.DSO_TYPE_RECONFIRM
}
