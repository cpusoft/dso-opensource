package packet

import (
	"errors"

	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/convert"
	"github.com/cpusoft/goutil/jsonutil"
)

type TxtModel struct {
	TxtData jsonutil.PrintableBytes `json:"txtData"`
}

func NewTxtModel(txtData string, offsetFromStart uint16) (txtModel *TxtModel, newOffsetFromStart uint16, err error) {
	belogs.Debug("NewTxtModel(): txtData:", txtData, "   offsetFromStart:", offsetFromStart)
	c := &TxtModel{
		TxtData: []byte(txtData),
	}
	newOffsetFromStart = offsetFromStart + uint16(len(txtData))
	belogs.Debug("NewPtrModel():  txtModel:", jsonutil.MarshalJson(c), "  newOffsetFromStart:", newOffsetFromStart)
	return c, newOffsetFromStart, nil
}
func ParseBytesToTxtModel(bytess []byte, offsetFromStart uint16) (txtModel *TxtModel, newOffsetFromStart uint16, err error) {
	belogs.Debug("ParseBytesToTxtModel(): bytess:", convert.PrintBytesOneLine(bytess), "   offsetFromStart:", offsetFromStart)
	if len(bytess) == 0 {
		belogs.Error("ParseBytesToTxtModel(): len(bytess) ==0, fail:", convert.PrintBytesOneLine(bytess), err)
		return nil, 0, errors.New("the length of txt is too short")
	}
	c := &TxtModel{
		TxtData: bytess,
	}
	newOffsetFromStart = offsetFromStart + uint16(len(bytess))
	belogs.Debug("ParseBytesToTxtModel(): bytes:", convert.PrintBytesOneLine(bytess),
		"   txtModel:", jsonutil.MarshalJson(c),
		"   newOffsetFromStart:", newOffsetFromStart)
	return c, newOffsetFromStart, nil
}

func (c *TxtModel) Bytes() []byte {
	return c.TxtData
}
func (c *TxtModel) Length() uint16 {
	return uint16(len(c.TxtData))
}
func (c *TxtModel) ToRrData() string {
	return string(c.TxtData)
}
