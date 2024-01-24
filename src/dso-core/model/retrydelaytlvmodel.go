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
// retry delay
type RetryDelayTlvModel struct {
	DsoType    uint16 `json:"dsoType"`
	DsoLength  uint16 `json:"dsoLength"`
	RetryDelay uint32 `json:"retryDelay"`
}

func NewRetryDelayTlvModel(retryDelay uint32) *RetryDelayTlvModel {
	c := &RetryDelayTlvModel{
		DsoType:    dnsutil.DSO_TYPE_RETRY_DELAY,
		DsoLength:  dnsutil.DSO_TYPE_RETRY_DELAY_LENGTH,
		RetryDelay: retryDelay,
	}
	return c
}

func NewMessageAndRetryDelayTlvModel(messageId uint16, qr uint8, rCode uint8,
	retryDelay uint32) *DsoModel {
	dsoModel, _ := NewDsoModelByParameters(messageId, qr, rCode)
	retryDelayTlvModel := NewRetryDelayTlvModel(retryDelay)
	dsoModel.AddTlvModel(retryDelayTlvModel)
	belogs.Info("#生成DSO的'主动关闭连接'类型数据包: " + osutil.GetNewLineSep() +
		"{'RetryDelay':" + convert.ToString(retryDelay) + "}")

	return dsoModel
}

func ParseBytesToRetryDelayTlvModel(dsoLength uint16, retryDelayBytes []byte,
	offsetFromStart uint16) (tlvModel TlvModel, newOffsetFromStart uint16, err error) {
	if len(retryDelayBytes) < dnsutil.DSO_TYPE_RETRY_DELAY_LENGTH {
		belogs.Error("ParseBytesToRetryDelayTlvModel(): recv byte's length is too small, fail:",
			" len(retryDelayBytes): ", len(retryDelayBytes), "   DSO_TYPE_RETRY_DELAY_LENGTH:", dnsutil.DSO_TYPE_RETRY_DELAY_LENGTH)
		return nil, 0, errors.New("Received packet is too small for legal DSO RetryDelay format")
	}
	if len(retryDelayBytes) < int(dsoLength) {
		belogs.Error("ParseBytesToRetryDelayTlvModel(): recv byte's length is too small, fail:",
			" len(retryDelayBytes): ", len(retryDelayBytes), "   dsoLength:", dsoLength)
		return nil, 0, errors.New("Received packet is too small for legal DSO RetryDelay format")
	}
	var retryDelay = binary.BigEndian.Uint32(retryDelayBytes[:4])
	newOffsetFromStart = offsetFromStart + 4
	belogs.Info("ParseBytesToRetryDelayTlvModel(): retryDelay:", retryDelay, "  newOffsetFromStart:", newOffsetFromStart)
	belogs.Info("#解析得到DSO的'主动关闭连接'类型数据包: " + osutil.GetNewLineSep() +
		"{'RetryDelay':" + convert.ToString(retryDelay) + "}")
	return NewRetryDelayTlvModel(retryDelay), newOffsetFromStart, nil

}

func (c RetryDelayTlvModel) Bytes() []byte {
	wr := bytes.NewBuffer([]byte{})
	binary.Write(wr, binary.BigEndian, c.DsoType)
	binary.Write(wr, binary.BigEndian, c.DsoLength)
	binary.Write(wr, binary.BigEndian, c.RetryDelay)
	return wr.Bytes()
}
func (c RetryDelayTlvModel) PrintBytes() string {
	return convert.PrintBytes(c.Bytes(), 8)
}
func (c RetryDelayTlvModel) GetDsoType() uint16 {
	return dnsutil.DSO_TYPE_RETRY_DELAY
}
