package model

import (
	"bytes"
	"encoding/binary"
	"errors"

	"dns-model/common"
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

type UpdateModel struct {
	HeaderForUpdateModel common.HeaderForUpdateModel `json:"headerForUpdateModel"`
	CountZPUAModel       common.CountZPUAModel       `json:"countZPUAModel"`
	UpdateDataModel      UpdateDataModel             `json:"updateDataModel"`
}

func (c UpdateModel) GetHeaderModel() common.HeaderModel {
	return c.HeaderForUpdateModel
}
func (c UpdateModel) GetCountModel() common.CountModel {
	return c.CountZPUAModel
}
func (c UpdateModel) GetDataModel() interface{} {
	return c.UpdateDataModel
}
func (c UpdateModel) GetDnsModelType() string {
	return "packet"
}

type UpdateDataModel struct {
	ZoneModel            *ZoneModel            `json:"zoneModel"`
	PrerequisiteModels   []*packet.PacketModel `json:"prerequisiteModel"`
	UpdateModels         []*packet.PacketModel `json:"updateModel"`
	AdditionalDataModels []*packet.PacketModel `json:"additionalDataModel"`
}

func NewUpdateModelByHeaderAndCount(headerModel common.HeaderModel, countModel common.CountModel) (updateModel *UpdateModel, err error) {
	belogs.Debug("NewUpdateModelByHeaderAndCount(): headerModel:", jsonutil.MarshalJson(headerModel), " countModel:", jsonutil.MarshalJson(countModel))

	c := &UpdateModel{}
	headerJson := jsonutil.MarshalJson(headerModel)
	countJson := jsonutil.MarshalJson(countModel)
	belogs.Debug("NewUpdateModelByHeaderAndCount(): headerJson:", headerJson, "   countJson:", countJson)

	err = jsonutil.UnmarshalJson(headerJson, &c.HeaderForUpdateModel)
	if err != nil {
		belogs.Error("NewUpdateModelByHeaderAndCount():UnmarshalJson HeaderForUpdateModel fail:", headerJson, err)
		return nil, err
	}
	err = jsonutil.UnmarshalJson(countJson, &c.CountZPUAModel)
	if err != nil {
		belogs.Error("NewUpdateModelByHeaderAndCount():UnmarshalJson CountZPUAModel fail:", countJson, err)
		return nil, err
	}

	c.UpdateDataModel.PrerequisiteModels = make([]*packet.PacketModel, 0)
	c.UpdateDataModel.UpdateModels = make([]*packet.PacketModel, 0)
	c.UpdateDataModel.AdditionalDataModels = make([]*packet.PacketModel, 0)
	return c, nil
}

func NewUpdateModelByParameters(id uint16, qr, rCode uint8) (updateModel *UpdateModel, err error) {
	belogs.Debug("NewUpdateModelByParameters(): id:", id, " qr:", qr, "  rCode:", rCode)
	parameter := uint16(common.ComposeQrOpCodeZRCode(qr, dnsutil.DNS_OPCODE_UPDATE, rCode))
	headerModel, _ := common.NewHeaderModel(id, parameter, common.DNS_HEADER_TYPE_UPDATE)
	countModel, _ := common.NewCountModel(0, 0, 0, 0, common.DNS_COUNT_TYPE_ZPUA)
	return NewUpdateModelByHeaderAndCount(headerModel, countModel)
}

// id ,qr, rCode -> header/count
// zName/zTtl(as origin) -> zoneModel
// zName and prerequisiteClassIntType --> prerequisite
// rrModels -> updateModels
func NewUpdateModelByParametersAndRrModels(id uint16,
	zName string, zTtl int64, prerequisiteClassIntType uint16,
	rrModels []*rr.RrModel) (updateModel *UpdateModel, err error) {
	belogs.Debug("NewUpdateModelByParametersAndRrModels(): id:", id,
		"  zName:", zName, "   prerequisiteClassIntType:", prerequisiteClassIntType,
		"  rrModels:", jsonutil.MarshalJson(rrModels))

	// header/count --> updateModel
	qr := dnsutil.DNS_QR_REQUEST
	rCode := dnsutil.DNS_RCODE_NOERROR
	updateModel, err = NewUpdateModelByParameters(id, qr, rCode)
	if err != nil {
		belogs.Error("NewUpdateModelByParametersAndRrModels(): NewUpdateModelByParameters fail:", err)
		return nil, err
	}
	belogs.Debug("NewUpdateModelByParametersAndRrModels(): updateModel:", jsonutil.MarshalJson(updateModel))

	// zoneModel
	zoneModel, _, err := NewZoneModel(zName, 0)
	if err != nil {
		belogs.Error("NewUpdateModelByParametersAndRrModels(): NewZoneModel fail,  zName:", zName, err)
		return nil, err
	}
	updateModel.SetZoneModel(zoneModel)
	belogs.Debug("NewUpdateModelByParametersAndRrModels():zoneModel:", jsonutil.MarshalJson(zoneModel))

	// prerequisite
	prerequisitePacketDomain, _, err := packet.NewPacketDomainNoCompression([]byte(zName), 0)
	if err != nil {
		belogs.Error("NewUpdateModelByParametersAndRrModels(): NewPacketDomainNoCompression fail,  zName:", zName, err)
		return nil, err
	}
	belogs.Debug("NewUpdateModelByParametersAndRrModels():prerequisitePacketDomain:", jsonutil.MarshalJson(prerequisitePacketDomain),
		"   bytes:", convert.PrintBytesOneLine(prerequisitePacketDomain.Bytes()))
	prerequisiteModel := packet.NewPacketModel(prerequisitePacketDomain, prerequisitePacketDomain.Bytes(),
		dnsutil.DNS_TYPE_INT_ANY, prerequisiteClassIntType, 0, 0,
		nil, nil, packet.DNS_PACKET_UPDATA_PREREQUISITE)
	belogs.Debug("NewUpdateModelByParametersAndRrModels():prerequisiteModel:", jsonutil.MarshalJson(prerequisiteModel))
	updateModel.AddPrerequisiteModel(prerequisiteModel)

	// update
	for i := range rrModels {
		packetModel, _, err := dnsconvert.ConvertRrToPacket(null.IntFrom(zTtl), rrModels[i], 0, packet.DNS_PACKET_NAME_TYPE_CLASS_TTL_RDLENGTH_RDATA)
		if err != nil {
			belogs.Error("NewUpdateModelByParametersAndRrModels(): ConvertRrToPacket fail,  zTtl:", zTtl,
				"  rrModels[i]:", jsonutil.MarshalJson(rrModels[i]), err)
			return nil, err
		}
		updateModel.AddUpdateModel(packetModel)
		belogs.Info("#生成SRP的'数据更新'类型数据包: " + osutil.GetNewLineSep() +
			"{'域名':'" + rrModels[i].RrFullDomain +
			"','Type':'" + rrModels[i].RrType +
			"','Class':'" + rrModels[i].RrClass +
			"','Ttl':" + convert.ToString(rrModels[i].RrTtl.ValueOrZero()) +
			",'Data':'" + rrModels[i].RrData + "'}")
	}
	belogs.Debug("NewUpdateModelByParametersAndRrModels(): ok updateModel:", jsonutil.MarshalJson(updateModel))
	return updateModel, nil
}

// receiveData: receiveBytes[offsetFromStart:]
func ParseBytesToUpdateModel(headerModel common.HeaderModel, countModel common.CountModel,
	receiveData []byte, offsetFromStart uint16) (receiveUpdateModel *UpdateModel, newOffsetFromStart uint16, err error) {
	belogs.Debug("ParseBytesToUpdateModel(): headerModel:", jsonutil.MarshalJson(headerModel),
		"  countModel:", jsonutil.MarshalJson(headerModel),
		"  receiveData:", convert.PrintBytesOneLine(receiveData), "   offsetFromStart:", offsetFromStart)

	packetDecompressionLabel := packet.NewPacketDecompressionLabel()
	receiveUpdateModel, err = NewUpdateModelByHeaderAndCount(headerModel, countModel)
	if err != nil {
		belogs.Error("ParseBytesToUpdateModel(): NewUpdateModelByHeaderAndCount fail:",
			" headerModel:", jsonutil.MarshalJson(headerModel), " countModel:", jsonutil.MarshalJson(countModel), err)
		return nil, 0, err
	}
	belogs.Info("ParseBytesToUpdateModel(): receiveUpdateModel:", jsonutil.MarshalJson(*receiveUpdateModel))
	// when response to client,may alltobe zero: rfc2136 3.8 - Response
	zoCount := receiveUpdateModel.CountZPUAModel.GetCount(0)
	prCount := receiveUpdateModel.CountZPUAModel.GetCount(1)
	upCount := receiveUpdateModel.CountZPUAModel.GetCount(2)
	adCount := receiveUpdateModel.CountZPUAModel.GetCount(3)

	// when **Count is zero, len(receiveData) will be 0
	if zoCount == 0 && prCount == 0 && upCount == 0 && adCount == 0 {
		belogs.Debug("ParseBytesToUpdateModel(): count all are zero:", jsonutil.MarshalJson(receiveUpdateModel), "  newOffsetFromStart:", newOffsetFromStart)
		return receiveUpdateModel, newOffsetFromStart, nil
	}
	belogs.Debug("ParseBytesToUpdateModel(): zoCount:", zoCount, "  prCount:", prCount,
		"   upCount:", upCount, " adCount:", adCount, "  newOffsetFromStart:", newOffsetFromStart,
		"   len(receiveData):", len(receiveData))

	// when **Count is not zero, len(receiveData) should > 0: rfc2136 3.8 - Response
	if len(receiveData) == 0 {
		belogs.Debug("ParseBytesToUpdateModel(): recv byte's may be zero for **Count:",
			"  len(receiveData):", len(receiveData))
		return nil, 0, errors.New("Received bytes is too small for legal UPDATE format")
	}

	var tmpStart, tmpStart1 uint16
	tmpStart = offsetFromStart
	tmpBytess := make([]byte, len(receiveData))
	copy(tmpBytess, receiveData)

	zoneModel, tmpStart1, err := ParseBytesToZoneModel(tmpBytess, tmpStart, packetDecompressionLabel)
	if err != nil {
		belogs.Error("ParseBytesToUpdateModel(): ParseBytesToZoneModel fail:",
			" tmpBytess:", convert.PrintBytesOneLine(tmpBytess), "  tmpStart1:", tmpStart1, err)
		return nil, 0, errors.New("Received bytes is failed to ParseBytesToZoneModel")
	}
	receiveUpdateModel.UpdateDataModel.ZoneModel = zoneModel
	belogs.Debug("ParseBytesToUpdateModel(): zoneModel:", jsonutil.MarshalJson(zoneModel))

	zoneLen := tmpStart1 - tmpStart
	tmpBytess = tmpBytess[zoneLen:]
	packetModels, newOffsetFromStart, err := packet.ParseBytesToPacketModels(tmpBytess, 0, offsetFromStart, packet.DNS_PACKET_NAME_TYPE_CLASS_TTL_RDLENGTH_RDATA, 0, packetDecompressionLabel)
	if err != nil {
		belogs.Error("ParseBytesToUpdateModel(): ParseBytesToPacketModels fail:",
			" tmpBytess:", convert.PrintBytesOneLine(tmpBytess), "  tmpStart1:", tmpStart1, err)
		return nil, 0, errors.New("Received bytes is failed to ParseBytesToZoneModel")
	}
	belogs.Debug("ParseBytesToUpdateModel(): ParseBytesToPacketModels,len(packetModels):", len(packetModels),
		"   packetModels:", jsonutil.MarshalJson(packetModels),
		"   zoCount:", zoCount, "  prCount:", prCount,
		"   upCount:", upCount, " adCount:", adCount, "  newOffsetFromStart:", newOffsetFromStart)
	if zoCount != 1 {
		belogs.Error("ParseBytesToUpdateModel():zoCount is not one, fail: zoCount:", zoCount)
		return nil, 0, errors.New("zoCount should be one")
	}
	if int(prCount+upCount+adCount) != len(packetModels) {
		belogs.Error("ParseBytesToUpdateModel(): counts sum is not equal to len(packetModels), fail:",
			" prCount:", prCount, " upCount:", upCount, "  adCount:", adCount, "  len(packetModels):", len(packetModels))
		return nil, 0, errors.New("Received bytes is failed to ParseBytesZPUAModel because count is not equal")
	}
	if int(prCount+upCount+adCount) == 0 {
		belogs.Debug("ParseBytesToUpdateModel(): int(prCount+upCount+adCount) == 0")
	} else {
		if prCount > 0 && int(prCount) <= len(packetModels) {
			receiveUpdateModel.UpdateDataModel.PrerequisiteModels = packetModels[:prCount]
		}
		if upCount > 0 && int(prCount+upCount) <= len(packetModels) {
			receiveUpdateModel.UpdateDataModel.UpdateModels = packetModels[prCount : prCount+upCount]
			for i := range receiveUpdateModel.UpdateDataModel.UpdateModels {
				updateModel := receiveUpdateModel.UpdateDataModel.UpdateModels[i]
				updateRrModel, _ := dnsconvert.ConvertPacketToRr("", updateModel)
				belogs.Info("#解析得到SRP的'数据更新'类型数据包: " + osutil.GetNewLineSep() +
					"{'域名':'" + updateRrModel.RrFullDomain +
					"','Type':'" + updateRrModel.RrType +
					"','Class':'" + updateRrModel.RrClass +
					"','Ttl':" + convert.ToString(updateRrModel.RrTtl.ValueOrZero()) +
					",'Data':'" + updateRrModel.RrData + "'}")
			}

		}
		if adCount > 0 && int(prCount+upCount+adCount) <= len(packetModels) {
			receiveUpdateModel.UpdateDataModel.AdditionalDataModels = packetModels[prCount+upCount : prCount : prCount+upCount+adCount]
		}
	}
	belogs.Debug("ParseBytesToUpdateModel(): receiveUpdateModel:", jsonutil.MarshalJson(receiveUpdateModel), "  newOffsetFromStart:", newOffsetFromStart)
	return receiveUpdateModel, newOffsetFromStart, nil
}

func (c *UpdateModel) Bytes() []byte {
	wr := bytes.NewBuffer([]byte{})
	binary.Write(wr, binary.BigEndian, c.HeaderForUpdateModel.Bytes())
	binary.Write(wr, binary.BigEndian, c.CountZPUAModel.Bytes())
	if c.UpdateDataModel.ZoneModel != nil {
		binary.Write(wr, binary.BigEndian, c.UpdateDataModel.ZoneModel.Bytes())
	}
	for i := range c.UpdateDataModel.PrerequisiteModels {
		binary.Write(wr, binary.BigEndian, c.UpdateDataModel.PrerequisiteModels[i].Bytes())
	}
	for i := range c.UpdateDataModel.UpdateModels {
		binary.Write(wr, binary.BigEndian, c.UpdateDataModel.UpdateModels[i].Bytes())
	}
	for i := range c.UpdateDataModel.AdditionalDataModels {
		binary.Write(wr, binary.BigEndian, c.UpdateDataModel.AdditionalDataModels[i].Bytes())
	}
	return wr.Bytes()
}
func (c *UpdateModel) SetZoneModel(zoneModel *ZoneModel) {
	c.UpdateDataModel.ZoneModel = zoneModel
	c.CountZPUAModel.ZoCount = 1
}

func (c *UpdateModel) AddPrerequisiteModel(prerequisiteModel *packet.PacketModel) {
	c.UpdateDataModel.PrerequisiteModels = append(c.UpdateDataModel.PrerequisiteModels, prerequisiteModel)
	c.CountZPUAModel.PrCount += 1
}
func (c *UpdateModel) AddUpdateModel(updateModel *packet.PacketModel) {
	c.UpdateDataModel.UpdateModels = append(c.UpdateDataModel.UpdateModels, updateModel)
	c.CountZPUAModel.UpCount += 1
}
func (c *UpdateModel) AddAdditionalDataModel(additionalDataModel *packet.PacketModel) {
	c.UpdateDataModel.AdditionalDataModels = append(c.UpdateDataModel.AdditionalDataModels, additionalDataModel)
	c.CountZPUAModel.AdCount += 1
}
