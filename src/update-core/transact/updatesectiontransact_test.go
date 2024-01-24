package transact

import (
	"testing"

	"dns-model/message"
	"dns-model/packet"
	"github.com/cpusoft/goutil/belogs"
	_ "github.com/cpusoft/goutil/conf"
	"github.com/cpusoft/goutil/dnsutil"
	"github.com/cpusoft/goutil/jsonutil"
	_ "github.com/cpusoft/goutil/logs"
	"github.com/cpusoft/goutil/xormdb"
	updatemodel "update-core/model"
)

func TestPerformUpdateTransact(t *testing.T) {
	// id uint16, qr, rCode uint8
	id := uint16(888)
	qr := uint8(dnsutil.DNS_QR_REQUEST)
	rCode := uint8(dnsutil.DNS_RCODE_NOERROR)
	updateModel, _ := updatemodel.NewUpdateModelByParameters(id, qr, rCode)
	belogs.Debug("TestPerformUpdateTransact():updateModel:", jsonutil.MarshalJson(updateModel))

	// zone
	zName := "example.com"
	offsetFromStart := uint16(16)
	zoneModel, newOffsetFromStart, _ := updatemodel.NewZoneModel(zName, offsetFromStart)
	updateModel.SetZoneModel(zoneModel)
	belogs.Debug("TestPerformUpdateTransact():zoneModel:", jsonutil.MarshalJson(zoneModel), newOffsetFromStart)

	// prerequisite
	packetDecompressionLabel := packet.NewPacketDecompressionLabel()
	l1, _ := packet.NewPacketLabel([]byte("example"), newOffsetFromStart)
	newOffsetFromStart += uint16(len("example") + 1)
	l2, _ := packet.NewPacketLabel([]byte("com"), newOffsetFromStart)
	newOffsetFromStart += uint16(len("com") + 1)
	ls := make([]*packet.PacketLabel, 0)
	ls = append(ls, l1)
	ls = append(ls, l2)
	packetDomain, err := packet.NewPacketDomainByAddPacketLabels(ls, packetDecompressionLabel)
	belogs.Debug("TestPerformUpdateTransact():prerequisiteModel packetDomain:", jsonutil.MarshalJson(packetDomain), newOffsetFromStart, err)
	//(packetDomain *PacketDomain, packetDomainBytes []byte, packetType uint16, packetClass uint16,
	// packetTtl uint32, packetDataLength uint16,
	// packetData interface{}, packetDataBytes []byte, packetModelType int)
	prerequisiteModel := packet.NewPacketModel(packetDomain, packetDomain.Bytes(), dnsutil.DNS_TYPE_INT_ANY, dnsutil.DNS_CLASS_INT_IN,
		0, 0,
		nil, nil, packet.DNS_PACKET_NAME_TYPE_CLASS)
	belogs.Debug("TestPerformUpdateTransact():prerequisiteModel:", jsonutil.MarshalJson(prerequisiteModel))
	updateModel.AddPrerequisiteModel(prerequisiteModel)

	// update
	// reset offset
	newOffsetFromStart = offsetFromStart + prerequisiteModel.Length()
	l1, _ = packet.NewPacketLabel([]byte("example"), newOffsetFromStart)
	newOffsetFromStart += uint16(len("example") + 1)
	l2, _ = packet.NewPacketLabel([]byte("com"), newOffsetFromStart)
	newOffsetFromStart += uint16(len("com") + 1)
	ls = make([]*packet.PacketLabel, 0)
	ls = append(ls, l1)
	ls = append(ls, l2)
	packetDomain, err = packet.NewPacketDomainByAddPacketLabels(ls, packetDecompressionLabel)
	belogs.Debug("TestPerformUpdateTransact(): updateModel:packetDomain:", jsonutil.MarshalJson(packetDomain), newOffsetFromStart, err)

	aModel, newOffsetFromStart, _ := packet.NewAModel("100.100.100.101", newOffsetFromStart)
	belogs.Debug("TestPerformUpdateTransact(): aModel:", jsonutil.MarshalJson(aModel))
	updateSectionModel := packet.NewPacketModel(packetDomain, packetDomain.Bytes(), dnsutil.DNS_TYPE_INT_A, dnsutil.DNS_CLASS_INT_IN,
		10000, aModel.Length(),
		aModel, aModel.Bytes(), packet.DNS_PACKET_NAME_TYPE_CLASS_TTL_RDLENGTH_RDATA)
	belogs.Debug("TestPerformUpdateTransact():updateSectionModel:", jsonutil.MarshalJson(updateSectionModel))
	updateModel.AddUpdateModel(updateSectionModel)

	// start mysql
	err = xormdb.InitMySql()
	if err != nil {
		belogs.Error("TestPerformUpdateTransact(): start InitMySql failed:", err)
		belogs.Debug("TestPerformUpdateTransact(): dns-server failed to start, ", err)
		return
	}
	defer xormdb.XormEngine.Close()

	responseDnsModel, err := PerformUpdateTransact(updateModel, message.DNS_TRANSACT_SIDE_SERVER)
	belogs.Debug("TestPerformUpdateTransact():responseDnsModel:", responseDnsModel, err)
}
