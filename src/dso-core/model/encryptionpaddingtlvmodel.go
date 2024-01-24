package model

import (
	"bytes"
	"encoding/binary"
	"errors"

	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/convert"
	"github.com/cpusoft/goutil/dnsutil"
	"github.com/cpusoft/goutil/osutil"
)

// /////////////////////////////
// encryption padding
type EncryptionPaddingTlvModel struct {
	DsoType           uint16 `json:"dsoType"`
	DsoLength         uint16 `json:"dsoLength"`
	EncryptionPadding []byte `json:"encryptionPadding"`
}

func NewEncryptionPaddingTlvModel(dsoLength uint16, encryptionPadding []byte) *EncryptionPaddingTlvModel {
	c := &EncryptionPaddingTlvModel{
		DsoType:           dnsutil.DSO_TYPE_ENCRYPTION_PADDING,
		DsoLength:         dsoLength,
		EncryptionPadding: encryptionPadding,
	}
	return c
}

func NewDsoAndEncryptionPaddingTlvModel(messageId uint16, qr uint8, rCode uint8,
	dsoLength uint16, encryptionPadding []byte) *DsoModel {
	dsoModel, _ := NewDsoModelByParameters(messageId, qr, rCode)
	ecryptionPaddingTlvModel := NewEncryptionPaddingTlvModel(dsoLength, encryptionPadding)
	dsoModel.AddTlvModel(ecryptionPaddingTlvModel)
	belogs.Info("#生成DSO的'加密填充'类型数据包: " + osutil.GetNewLineSep() + "{'EncryptionPadding':'" + convert.PrintBytesOneLine(encryptionPadding) + "'}")
	return dsoModel
}

func ParseBytesToEncryptionPaddingTlvModel(dsoLength uint16, encryptionPaddingBytes []byte,
	offsetFromStart uint16) (tlvModel TlvModel, newOffsetFromStart uint16, err error) {

	if len(encryptionPaddingBytes) < int(dsoLength) {
		belogs.Error("ParseBytesToEncryptionPaddingTlvModel(): recv byte's length is too small, fail:",
			"  len(encryptionPaddingBytes):", len(encryptionPaddingBytes), "  dsoLength:", dsoLength)
		return nil, 0, errors.New("Received packet is too small for legal DSO EncryptionPaddingTlv format")
	}
	encryptionPadding := encryptionPaddingBytes
	newOffsetFromStart = offsetFromStart + uint16(len(encryptionPadding))
	belogs.Info("ParseBytesToEncryptionPaddingTlvModel():dsoLength:", dsoLength, "  encryptionPadding:", convert.PrintBytesOneLine(encryptionPadding))
	belogs.Info("#解析得到DSO的'加密填充'类型数据包: " + osutil.GetNewLineSep() + "{'EncryptionPadding':'" + convert.PrintBytesOneLine(encryptionPadding) + "'}")
	return NewEncryptionPaddingTlvModel(dsoLength, encryptionPadding), newOffsetFromStart, nil
}

func (c *EncryptionPaddingTlvModel) Bytes() []byte {
	wr := bytes.NewBuffer([]byte{})
	binary.Write(wr, binary.BigEndian, c.DsoType)
	binary.Write(wr, binary.BigEndian, c.DsoLength)
	binary.Write(wr, binary.BigEndian, c.EncryptionPadding)
	return wr.Bytes()

}
func (c *EncryptionPaddingTlvModel) PrintBytes() string {
	return convert.PrintBytes(c.Bytes(), 8)
}
func (c *EncryptionPaddingTlvModel) GetDsoType() uint16 {
	return dnsutil.DSO_TYPE_ENCRYPTION_PADDING
}
