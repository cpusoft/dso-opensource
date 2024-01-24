package common

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"

	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/convert"
	"github.com/cpusoft/goutil/dnsutil"
	"github.com/cpusoft/goutil/jsonutil"
)

const (
	DNS_HEADER_LENGTH      = 4
	DNS_HEADER_TYPE_QUERY  = 1
	DNS_HEADER_TYPE_UPDATE = 2
	DNS_HEADER_TYPE_DSO    = 3
)

type HeaderModel interface {
	Bytes() []byte
	GetIdOrMessageId() uint16
	GetQr() uint8
	GetOpCode() uint8
	GetZ() uint8
	GetRCode() uint8

	GetHeaderType() int
}
type HeaderForQueryModel struct {
	Id uint16 `json:"id"`

	QrOpCodeAaTcRdRaZRCode uint16 `json:"qrOpCodeAaTcRdRaZRCode"`
	Qr                     uint8  `json:"qr"`
	OpCode                 uint8  `json:"opCode"`
	Aa                     uint8  `json:"aa"`
	Tc                     uint8  `json:"tc"`
	Rd                     uint8  `json:"rd"`
	Ra                     uint8  `json:"ra"`
	Z                      uint8  `json:"z"`
	RCode                  uint8  `json:"rCode"`
}

type HeaderForUpdateModel struct {
	Id uint16 `json:"id"`

	QrOpCodeZRCode uint16 `json:"qrOpCodeZRCode"`
	Qr             uint8  `json:"qr"`
	OpCode         uint8  `json:"opCode"`
	Z              uint8  `json:"z"`
	RCode          uint8  `json:"rCode"`
}

type HeaderForDsoModel struct {
	MessageId uint16 `json:"messageId"`

	QrOpCodeZRCode uint16 `json:"qrOpCodeZRCode"`
	Qr             uint8  `json:"qr"`
	OpCode         uint8  `json:"opCode"`
	Z              uint8  `json:"z"`
	RCode          uint8  `json:"rCode"`
}

func NewHeaderModel(id uint16, parameters uint16, headerType int) (headerModel HeaderModel, err error) {

	if headerType == DNS_HEADER_TYPE_QUERY {
		qr, opCode, aa, tc, rd, ra, z, rCode, err := ParseQrOpCodeAaTcRdRaZRCode(QrOpCodeAaTcRdRaZRCode(parameters))
		c := &HeaderForQueryModel{
			Id:     id,
			Qr:     qr,
			OpCode: opCode,
			Aa:     aa,
			Tc:     tc,
			Rd:     rd,
			Ra:     ra,
			Z:      z,
			RCode:  rCode,
		}
		if err != nil {
			belogs.Error("NewHeaderModel(): DNS_HEADER_TYPE_QUERY ParseQrOpCodeAaTcRdRaZRCode fail,err:", err)
			return nil, err
		}
		c.QrOpCodeAaTcRdRaZRCode = parameters
		return c, nil
	} else if headerType == DNS_HEADER_TYPE_UPDATE {
		qr, opCode, z, rCode, err := ParseQrOpCodeZRCode(QrOpCodeZRCode(parameters))
		c := &HeaderForUpdateModel{
			Id:     id,
			Qr:     qr,
			OpCode: opCode,
			Z:      z,
			RCode:  rCode,
		}
		if err != nil {
			belogs.Error("NewHeaderModel():DNS_HEADER_TYPE_UPDATE ParseQrOpCodeZRCode fail,err:", err)
			return nil, err
		}
		c.QrOpCodeZRCode = parameters
		return c, nil
	} else if headerType == DNS_HEADER_TYPE_DSO {
		qr, opCode, z, rCode, err := ParseQrOpCodeZRCode(QrOpCodeZRCode(parameters))
		c := &HeaderForDsoModel{
			MessageId: id,
			Qr:        qr,
			OpCode:    opCode,
			Z:         z,
			RCode:     rCode,
		}
		if err != nil {
			belogs.Error("NewHeaderModel():DNS_HEADER_TYPE_DSO ParseQrOpCodeZRCode fail,err:", err)
			return nil, err
		}
		c.QrOpCodeZRCode = parameters
		return c, nil
	}
	return nil, errors.New("not support header type")
}

func ParseBytesToHeaderModel(headerBytes []byte) (headerType int, headerModel HeaderModel, err error) {
	if len(headerBytes) != DNS_HEADER_LENGTH {
		belogs.Error("ParseBytesToHeaderModel():")
		return 0, nil, errors.New("Header Length is illegal")
	}
	belogs.Debug("ParseBytesToHeaderModel():headerBytes: ", convert.PrintBytesOneLine(headerBytes))

	// get id/messageId
	id := binary.BigEndian.Uint16(headerBytes[:2])
	belogs.Debug("ParseBytesToHeaderModel(): id:", id, "0x:", fmt.Sprintf("%0x", id))

	parameters := binary.BigEndian.Uint16(headerBytes[2:4])
	opCode := uint8((parameters >> 11) & 0x0f) // 0111 1000 0000 0000 --> 1111
	belogs.Debug("ParseBytesToHeaderModel(): parameters:", parameters, " 0x:", fmt.Sprintf("%0x", parameters),
		" opCode:", opCode, "0x:", fmt.Sprintf("%0x", opCode))

	// check opCode
	if opCode != dnsutil.DNS_OPCODE_QUERY && opCode != dnsutil.DNS_OPCODE_UPDATE && opCode != dnsutil.DNS_OPCODE_DSO {
		belogs.Error("ParseBytesToHeaderModel():opCode is neither 0(query), nor 5(UPDATE) ,nor 6(DSO),  opCode:", opCode)
		return 0, nil, errors.New("OpCode is neither 5(UPDATE) , nor 5(UPDATE) ,nor 6(DSO)")
	}

	if opCode == dnsutil.DNS_OPCODE_QUERY {
		headerType = DNS_HEADER_TYPE_QUERY
	} else if opCode == dnsutil.DNS_OPCODE_UPDATE {
		headerType = DNS_HEADER_TYPE_UPDATE
	} else if opCode == dnsutil.DNS_OPCODE_DSO {
		headerType = DNS_HEADER_TYPE_DSO
	} else {
		belogs.Error("ParseBytesToHeaderModel():opCode is neither 5(UPDATE) nor 6(DSO), fail, opCode:", opCode)
		return 0, nil, errors.New("OpCode is neither 5(UPDATE) nor 6(DSO)")
	}
	belogs.Debug("ParseBytesToHeaderModel(): opCode:", opCode, "   headerType:", headerType)

	c, err := NewHeaderModel(id, parameters, headerType)
	if err != nil {
		belogs.Error("ParseBytesToHeaderModel():NewHeaderModel fail,id:", id, " parameters:", parameters,
			"  headerType:", headerType)
		return 0, nil, errors.New("NewHeaderModel fail")
	}

	belogs.Info("ParseBytesToHeaderModel(): NewHeaderModel:", jsonutil.MarshalJson(c))
	return headerType, c, nil
}
func (c HeaderForQueryModel) Bytes() []byte {
	wr := bytes.NewBuffer([]byte{})
	binary.Write(wr, binary.BigEndian, c.Id)
	binary.Write(wr, binary.BigEndian, c.QrOpCodeAaTcRdRaZRCode)
	return wr.Bytes()
}
func (c HeaderForQueryModel) GetIdOrMessageId() uint16 {
	return c.Id
}
func (c HeaderForQueryModel) GetQr() uint8 {
	return c.Qr
}
func (c HeaderForQueryModel) GetOpCode() uint8 {
	return c.OpCode
}
func (c HeaderForQueryModel) GetZ() uint8 {
	return c.Z
}
func (c HeaderForQueryModel) GetRCode() uint8 {
	return c.RCode
}
func (c HeaderForQueryModel) GetHeaderType() int {
	return DNS_HEADER_TYPE_QUERY
}

func (c HeaderForUpdateModel) Bytes() []byte {
	wr := bytes.NewBuffer([]byte{})
	binary.Write(wr, binary.BigEndian, c.Id)
	binary.Write(wr, binary.BigEndian, c.QrOpCodeZRCode)
	return wr.Bytes()
}
func (c HeaderForUpdateModel) GetIdOrMessageId() uint16 {
	return c.Id
}
func (c HeaderForUpdateModel) GetQr() uint8 {
	return c.Qr
}
func (c HeaderForUpdateModel) GetOpCode() uint8 {
	return c.OpCode
}
func (c HeaderForUpdateModel) GetZ() uint8 {
	return c.Z
}
func (c HeaderForUpdateModel) GetRCode() uint8 {
	return c.RCode
}
func (c HeaderForUpdateModel) GetHeaderType() int {
	return DNS_HEADER_TYPE_UPDATE
}
func (c HeaderForDsoModel) Bytes() []byte {
	wr := bytes.NewBuffer([]byte{})
	binary.Write(wr, binary.BigEndian, c.MessageId)
	binary.Write(wr, binary.BigEndian, c.QrOpCodeZRCode)
	return wr.Bytes()
}
func (c HeaderForDsoModel) GetIdOrMessageId() uint16 {
	return c.MessageId
}
func (c HeaderForDsoModel) GetQr() uint8 {
	return c.Qr
}
func (c HeaderForDsoModel) GetOpCode() uint8 {
	return c.OpCode
}
func (c HeaderForDsoModel) GetZ() uint8 {
	return c.Z
}
func (c HeaderForDsoModel) GetRCode() uint8 {
	return c.RCode
}
func (c HeaderForDsoModel) GetHeaderType() int {
	return DNS_HEADER_TYPE_DSO
}
