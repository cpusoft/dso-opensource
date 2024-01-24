package model

import (
	"bytes"
	"encoding/binary"
	"errors"

	"dns-model/common"
	"dns-model/packet"
	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/convert"
	"github.com/cpusoft/goutil/dnsutil"
	"github.com/cpusoft/goutil/jsonutil"
)

/////////////////////////
//

type DsoModel struct {
	HeaderForDsoModel common.HeaderForDsoModel `json:"headerForDsoModel"`
	CountQANAModel    common.CountQANAModel    `json:"countQANAModel"`
	DsoDataModel      DsoDataModel             `json:"dsoDataModel"`
}

func (c DsoModel) GetHeaderModel() common.HeaderModel {
	return c.HeaderForDsoModel
}
func (c DsoModel) GetCountModel() common.CountModel {
	return c.CountQANAModel
}
func (c DsoModel) GetDataModel() interface{} {
	return c.DsoDataModel
}
func (c DsoModel) GetDnsModelType() string {
	return "packet"
}

type DsoDataModel struct {
	TlvModels []TlvModel `json:"tlvModels"`
}

func NewDsoModelByHeaderAndCount(headerModel common.HeaderModel, countModel common.CountModel) (dsoModel *DsoModel, err error) {
	c := &DsoModel{}
	headerJson := jsonutil.MarshalJson(headerModel)
	countJson := jsonutil.MarshalJson(countModel)
	belogs.Debug("NewDsoModelByHeaderAndCount(): headerJson:", headerJson, "   countJson:", countJson)

	err = jsonutil.UnmarshalJson(headerJson, &c.HeaderForDsoModel)
	if err != nil {
		belogs.Error("NewDsoModelByHeaderAndCount():UnmarshalJson HeaderForDsoModel fail:", headerJson, err)
		return nil, err
	}
	err = jsonutil.UnmarshalJson(countJson, &c.CountQANAModel)
	if err != nil {
		belogs.Error("NewDsoModelByHeaderAndCount():UnmarshalJson CountQANAModel fail:", countJson, err)
		return nil, err
	}

	c.DsoDataModel.TlvModels = make([]TlvModel, 0)
	return c, nil
}

func NewDsoModelByParameters(id uint16, qr, rCode uint8) (dsoModel *DsoModel, err error) {
	parameter := common.ComposeQrOpCodeZRCode(qr, dnsutil.DNS_OPCODE_DSO, rCode)
	headerModel, _ := common.NewHeaderModel(id, uint16(parameter), common.DNS_HEADER_TYPE_DSO)
	countModel, _ := common.NewCountModel(0, 0, 0, 0, common.DNS_COUNT_TYPE_QANA)
	return NewDsoModelByHeaderAndCount(headerModel, countModel)
}

func (c *DsoModel) AddTlvModel(tlvModel TlvModel) {
	c.DsoDataModel.TlvModels = append(c.DsoDataModel.TlvModels, tlvModel)
}
func (c *DsoModel) AddTlvModels(tlvModels []TlvModel) {
	c.DsoDataModel.TlvModels = append(c.DsoDataModel.TlvModels, tlvModels...)
}

func (c DsoModel) Bytes() []byte {
	wr := bytes.NewBuffer([]byte{})
	binary.Write(wr, binary.BigEndian, c.HeaderForDsoModel.Bytes())
	binary.Write(wr, binary.BigEndian, c.CountQANAModel.Bytes())
	for i := range c.DsoDataModel.TlvModels {
		binary.Write(wr, binary.BigEndian, (c.DsoDataModel.TlvModels[i]).Bytes())
	}
	return wr.Bytes()
}

func ParseBytesToDsoModel(headerModel common.HeaderModel, countModel common.CountModel,
	receiveData []byte, offsetFromStart uint16) (receiveDsoModel *DsoModel, newOffsetFromStart uint16, err error) {
	belogs.Debug("ParseBytesToDsoModel(): headerModel:", jsonutil.MarshalJson(headerModel),
		"  countModel:", jsonutil.MarshalJson(headerModel),
		"  receiveData:", convert.PrintBytesOneLine(receiveData), "   offsetFromStart:", offsetFromStart)

	receiveDsoModel, err = NewDsoModelByHeaderAndCount(headerModel, countModel)
	if err != nil {
		belogs.Error("ParseBytesToDsoModel():NewDsoModelByHeaderAndCount fail,headerModel: ", jsonutil.MarshalJson(headerModel),
			"   countModel:", jsonutil.MarshalJson(countModel), err)
		return nil, 0, err
	}
	belogs.Info("ParseBytesToDsoModel(): receiveDsoModel:", jsonutil.MarshalJson(*receiveDsoModel))
	if len(receiveData) < dnsutil.DSO_LENGTH_MIN {
		qr := receiveDsoModel.GetHeaderModel().GetQr()
		if qr == dnsutil.DNS_QR_RESPONSE {
			// client's receiveData may be smaller than DSO_LENGTH_MIN
			belogs.Debug("ParseBytesToDsoModel(): recv byte's length is too small, qr is QR_RESPONSE:", qr,
				" len(receiveData):", len(receiveData), " can be smaller than DSO_LENGTH_MIN:", dnsutil.DSO_LENGTH_MIN)
			return receiveDsoModel, 0, nil
		} else {
			// server's receiveData should not be smaller than DSO_LENGTH_MIN
			belogs.Error("ParseBytesToDsoModel(): recv byte's length is too small, qr is QR_REQUEST:", qr,
				" len(receiveData):", len(receiveData), " cannot be smaller than DSO_LENGTH_MIN:", dnsutil.DSO_LENGTH_MIN)
			return nil, 0, errors.New("Received bytes is too small for legal DSO format")
		}
	}

	packetDecompressionLabel := packet.NewPacketDecompressionLabel()
	var tmpStart, loopStart uint16
	loopStart = offsetFromStart
	tmpStart = offsetFromStart
	var tmpBytess, loopBytess []byte
	tmpBytess = make([]byte, len(receiveData))
	copy(tmpBytess, receiveData)

	for {
		loopBytess = tmpBytess
		loopStart = tmpStart
		if len(loopBytess) < dnsutil.DSO_TLV_LENGTH_MIN {
			belogs.Debug("ParseBytesToDsoModel(): result receiveDsoModel:", jsonutil.MarshalJson(receiveDsoModel),
				"   loopBytess:", convert.PrintBytesOneLine(loopBytess))
			belogs.Info("ParseBytesToDsoModel(): result receiveDsoModel:", jsonutil.MarshalJson(receiveDsoModel),
				"   len(loopBytess):", len(loopBytess))
			return receiveDsoModel, loopStart, nil
		}

		// err is dnsError
		dsoType, dsoLength, tmpStart, err := parseToTypeAndLength(loopBytess, loopStart)
		if err != nil {
			belogs.Error("ParseBytesToDsoModel():parseToTypeAndLength fail,loopBytess: ", convert.PrintBytesOneLine(loopBytess), err)
			return nil, 0, err
		}
		tmpBytess = loopBytess[4:]
		belogs.Debug("ParseBytesToDsoModel(): dsoType:", dsoType, "   dsoLength:", dsoLength, "  tmpStart:", tmpStart)

		tmpStart1 := tmpStart
		// err is dnsError
		var tlvModel TlvModel
		switch dsoType {
		case dnsutil.DSO_TYPE_KEEPALIVE:
			tlvModel, tmpStart, err = ParseBytesToKeepaliveTlvModel(dsoLength, tmpBytess, tmpStart1)
		case dnsutil.DSO_TYPE_RETRY_DELAY:
			tlvModel, tmpStart, err = ParseBytesToRetryDelayTlvModel(dsoLength, tmpBytess, tmpStart1)
		case dnsutil.DSO_TYPE_ENCRYPTION_PADDING:
			tlvModel, tmpStart, err = ParseBytesToEncryptionPaddingTlvModel(dsoLength, tmpBytess, tmpStart1)
		case dnsutil.DSO_TYPE_SUBSCRIBE:
			tlvModel, tmpStart, err = ParseBytesToSubscribeTlvModel(dsoLength, tmpBytess, tmpStart1, packetDecompressionLabel)
		case dnsutil.DSO_TYPE_PUSH:
			tlvModel, tmpStart, err = ParseBytesToPushTlvModel(dsoLength, tmpBytess, tmpStart1, packetDecompressionLabel)
		case dnsutil.DSO_TYPE_UNSUBSCRIBE:
			tlvModel, tmpStart, err = ParseBytesToUnsubscribeTlvModel(dsoLength, tmpBytess, tmpStart1)
		case dnsutil.DSO_TYPE_RECONFIRM:
			tlvModel, tmpStart, err = ParseBytesToReconfirmTlvModel(dsoLength, tmpBytess, tmpStart1, packetDecompressionLabel)
		default:
			belogs.Error("ParseBytesToDsoModel():switch dsoType fail, dsoType is error: ", dsoType)
			return nil, 0, errors.New("not support dsoType")
		}
		if err != nil {
			belogs.Error("ParseBytesToDsoModel(): ParseTo***TlvModel fail, dsoLength:", dsoLength,
				"   tmpBytess:", convert.PrintBytesOneLine(tmpBytess), "   tmpStart1:", tmpStart1, err)
			return nil, 0, err
		}

		receiveDsoModel.AddTlvModel(tlvModel)
		belogs.Debug("ParseBytesToDsoModel():new loop, tlvModel:", jsonutil.MarshalJson(tlvModel),
			"  tmpStart:", tmpStart)

		// get new leftBytess
		tlvModelLen := tmpStart - tmpStart1
		leftLen := len(tmpBytess) - int(tlvModelLen)
		if leftLen == 0 {
			belogs.Info("ParseBytesToDsoModel(): leftLen == 0, receiveDsoModel:", jsonutil.MarshalJson(receiveDsoModel))
			return receiveDsoModel, tmpStart, nil
		}
		loopStart = tmpStart
		tmpBytess1 := make([]byte, leftLen)
		copy(tmpBytess1, tmpBytess[tlvModelLen:])
		tmpBytess = tmpBytess1
		belogs.Debug("ParseBytesToDsoModel():tmpBytess:", convert.PrintBytesOneLine(tmpBytess),
			"  newOffsetFromStart:", newOffsetFromStart)
	}

}

// err is dnsError
func parseToTypeAndLength(bytess []byte, offsetFromStart uint16) (dsoType, dsoLength uint16,
	newOffsetFromStart uint16, err error) {
	belogs.Debug("parseToTypeAndLength(): bytess:", convert.PrintBytesOneLine(bytess), "  offsetFromStart:", offsetFromStart)
	if len(bytess) < 4 {
		belogs.Error("parseToTypeAndLength(): bytess:", convert.PrintBytesOneLine(bytess))
		return 0, 0, 0, errors.New("Length of dos is too short")
	}
	dsoType = binary.BigEndian.Uint16(bytess[:2])
	offsetFromStart += 2
	belogs.Debug("parseToTypeAndLength(): dsoType:", dsoType, "  offsetFromStart:", offsetFromStart)

	dsoLength = binary.BigEndian.Uint16(bytess[2:4])
	offsetFromStart += 2
	belogs.Debug("parseToTypeAndLength(): dsoLength:", dsoLength, "  offsetFromStart:", offsetFromStart)
	return dsoType, dsoLength, offsetFromStart, nil

}

// for callTransactTlvFunc as parameter
type SimpleDsoModel struct {
	MessageId uint16   `json:"messageId"`
	Qr        uint8    `json:"qr"`
	RCode     uint8    `json:"rCode"`
	TlvIndex  uint64   `json:"tlvIndex"`
	TlvModel  TlvModel `json:"tlvModel"`
}

func GetPrimaryTlvDsoType(dsoModel *DsoModel) (exist bool, dsoType uint16) {
	if len(dsoModel.DsoDataModel.TlvModels) == 0 {
		return false, 0
	}
	return true, dsoModel.DsoDataModel.TlvModels[0].GetDsoType()
}
