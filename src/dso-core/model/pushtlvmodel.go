package model

import (
	"bytes"
	"encoding/binary"
	"errors"

	dnsconvert "dns-model/convert"
	packet "dns-model/packet"
	"dns-model/rr"
	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/convert"
	"github.com/cpusoft/goutil/dnsutil"
	"github.com/cpusoft/goutil/jsonutil"
	"github.com/cpusoft/goutil/osutil"
	"github.com/guregu/null"
)

// /////////////////////////////
// push
type PushTlvModel struct {
	DsoType         uint16                `json:"dsoType"`
	DsoLength       uint16                `json:"dsoLength"`
	DnsPacketModels []*packet.PacketModel `json:"dnsPacketModels"`
	DnsRrModels     []*rr.RrModel         `json:"dnsRrModels"`
}

func NewPushTlvModel() *PushTlvModel {
	c := &PushTlvModel{
		DsoType:         dnsutil.DSO_TYPE_PUSH,
		DsoLength:       0,
		DnsPacketModels: make([]*packet.PacketModel, 0),
	}
	return c
}

func NewDsoModelWithPushTlvModel(rrModels []*rr.RrModel) (dsoModel *DsoModel, err error) {
	belogs.Debug("NewDsoModelWithPushTlvModel(): rrModels:", jsonutil.MarshalJson(rrModels))

	// dso header/count:
	// QR_RESPONSE:
	// messageId is 0
	dsoModel, err = NewDsoModelByParameters(0, dnsutil.DNS_QR_RESPONSE, dnsutil.DNS_RCODE_NOERROR)
	if err != nil {
		belogs.Error("NewDsoModelWithPushTlvModel(): NewDsoModelByParameters fail:", err)
		return nil, err
	}
	belogs.Debug("NewDsoModelWithPushTlvModel(): NewDsoModelByParameters, dsoModel:", jsonutil.MarshalJson(dsoModel))

	pushTlvModel := NewPushTlvModel()
	for i := range rrModels {
		//convert.ConvertRrToPacket(originModel *rr.OriginModel, rrModel *rr.RrModel, offsetFromStart uint16,
		//	packetModelType int) (packetModel *packet.PacketModel, newOffsetFromStart uint16, err error)
		packetModel, _, err := dnsconvert.ConvertRrToPacket(null.NewInt(0, false), rrModels[i], 0,
			packet.DNS_PACKET_NAME_TYPE_CLASS_TTL_RDLENGTH_RDATA)
		if err != nil {
			belogs.Error("NewDsoModelWithPushTlvModel(): ConvertRrToPacket fail, rrModels[i]:", jsonutil.MarshalJson(rrModels[i]), err)
			return nil, err
		}
		belogs.Debug("NewDsoModelWithPushTlvModel(): ConvertRrToPacket ok, packetModel:", jsonutil.MarshalJson(packetModel))
		pushTlvModel.DnsPacketModels = append(pushTlvModel.DnsPacketModels, packetModel)
		belogs.Info("#生成DSO的'数据推送'类型数据包: " + osutil.GetNewLineSep() +
			"{'域名':'" + rrModels[i].RrFullDomain +
			"','Type':'" + rrModels[i].RrType +
			"','Class':'" + rrModels[i].RrClass +
			"','Ttl':" + convert.ToString(rrModels[i].RrTtl.ValueOrZero()) +
			",'Data':'" + rrModels[i].RrData + "'}")
	}
	dsoModel.AddTlvModel(pushTlvModel)
	belogs.Debug("NewDsoModelWithPushTlvModel(): dsoModel:", jsonutil.MarshalJson(dsoModel))
	return dsoModel, nil
}

func ParseBytesToPushTlvModel(dsoLength uint16, pushBytes []byte, offsetFromStart uint16,
	packetDecompressionLabel *packet.PacketDecompressionLabel) (tlvModel TlvModel,
	newOffsetFromStart uint16, err error) {
	if len(pushBytes) < dnsutil.DSO_TYPE_PUSH_MIN_LENGTH {
		belogs.Error("ParseBytesToPushTlvModel(): recv byte's length is too small, fail, len(pushBytes):", len(pushBytes),
			"  DSO_TYPE_PUSH_MIN_LENGTH:", dnsutil.DSO_TYPE_PUSH_MIN_LENGTH)
		return nil, 0, errors.New("Received packet is too small for legal DSO Push format")
	}
	if len(pushBytes) < int(dsoLength) {
		belogs.Error("ParseBytesToPushTlvModel(): recv byte's length is too small, fail:",
			" len(pushBytes):", len(pushBytes), "  dsoLength:", dsoLength)
		return nil, 0, errors.New("Received packet is too small for legal DSO Push format")
	}
	belogs.Debug("ParseBytesToPushTlvModel():dsoLength:", dsoLength, "    len(pushBytes):", len(pushBytes))

	dsoDataModels, newOffsetFromStart, err := packet.ParseBytesToPacketModels(pushBytes, dsoLength, offsetFromStart, packet.DNS_PACKET_NAME_TYPE_CLASS_TTL_RDLENGTH_RDATA, 0, packetDecompressionLabel)
	if err != nil {
		belogs.Error("ParseBytesToPushTlvModel(): ParseBytesToPacketModels fail: pushBytes: ", convert.PrintBytesOneLine(pushBytes),
			"  dsoLength:", dsoLength, "   offsetFromStart:", offsetFromStart)
		return nil, 0, errors.New("Received packet is illegal for legal DSO Push format:" + err.Error())
	}
	pushTlvModel := NewPushTlvModel()
	for i := range dsoDataModels {
		pushTlvModel.AddDsoDataModel(dsoDataModels[i])
	}
	belogs.Info("ParseBytesToPushTlvModel(): pushTlvModel:", jsonutil.MarshalJson(pushTlvModel), "  newOffsetFromStart:", newOffsetFromStart)

	// packetModels --> rrModels
	err = pushTlvModel.ConvertPacketModelsToRrModels()
	if err != nil {
		belogs.Error("ParseBytesToPushTlvModel(): ConvertPacketModelsToRrModels fail: ", err)
		return nil, 0, errors.New("Received packet is illegal for legal DSO Push RrModel format:" + err.Error())
	}
	for i := range pushTlvModel.DnsRrModels {
		belogs.Info("#解析得到DSO的'数据推送'类型数据包: " + osutil.GetNewLineSep() +
			"{'域名':'" + pushTlvModel.DnsRrModels[i].RrFullDomain + "','Type':'" + pushTlvModel.DnsRrModels[i].RrType +
			"','Class':'" + pushTlvModel.DnsRrModels[i].RrClass +
			"','Ttl':" + convert.ToString(pushTlvModel.DnsRrModels[i].RrTtl.ValueOrZero()) +
			",'Data':'" + pushTlvModel.DnsRrModels[i].RrData + "'}")
	}
	return pushTlvModel, newOffsetFromStart, nil
}

func (c *PushTlvModel) AddDsoDataModel(dnsPacketModel *packet.PacketModel) {
	c.DnsPacketModels = append(c.DnsPacketModels, dnsPacketModel)
	c.DsoLength += dnsPacketModel.Length()
}

func (c PushTlvModel) Bytes() []byte {
	wr := bytes.NewBuffer([]byte{})
	binary.Write(wr, binary.BigEndian, c.DsoType)
	binary.Write(wr, binary.BigEndian, c.DsoLength)
	for i := range c.DnsPacketModels {
		binary.Write(wr, binary.BigEndian, (c.DnsPacketModels[i]).Bytes())
	}
	return wr.Bytes()
}

func (c PushTlvModel) PrintBytes() string {
	return convert.PrintBytes(c.Bytes(), 8)
}
func (c PushTlvModel) GetDsoType() uint16 {
	return dnsutil.DSO_TYPE_PUSH
}
func (c *PushTlvModel) ConvertPacketModelsToRrModels() error {
	c.DnsRrModels = make([]*rr.RrModel, 0)
	if len(c.DnsPacketModels) == 0 {
		return nil
	}
	for i := range c.DnsPacketModels {
		dnsRrModel, err := dnsconvert.ConvertPacketToRr("", c.DnsPacketModels[i])
		if err != nil {
			belogs.Error("ConvertPacketModelsToRrModels(): ConvertPacketToRr fail, c.DnsPacketModels[i]:", jsonutil.MarshalJson(c.DnsPacketModels[i]))
			return err
		}
		c.DnsRrModels = append(c.DnsRrModels, dnsRrModel)
	}
	belogs.Debug("ConvertPacketModelsToRrModels(): c.DnsRrModels:", jsonutil.MarshalJson(c.DnsRrModels))
	belogs.Info("ConvertPacketModelsToRrModels(): len(c.DnsRrModels):", len(c.DnsRrModels))
	return nil
}
