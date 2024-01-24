package model

import (
	"bytes"
	"encoding/binary"
	"errors"

	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/convert"
	"github.com/cpusoft/goutil/dnsutil"
	"github.com/cpusoft/goutil/jsonutil"
	"github.com/cpusoft/goutil/osutil"
)

// /////////////////////////////
// unsubscribe
type UnsubscribeTlvModel struct {
	DsoType            uint16 `json:"dsoType"`
	DsoLength          uint16 `json:"dsoLength"`
	SubscribeMessageId uint16 `json:"subscribeMessageId"`
}

func NewUnsubscribeTlvModel(subscribeMessageId uint16) *UnsubscribeTlvModel {
	c := &UnsubscribeTlvModel{
		DsoType:            dnsutil.DSO_TYPE_UNSUBSCRIBE,
		DsoLength:          dnsutil.DSO_TYPE_UNSUBSCRIBE_LENGTH,
		SubscribeMessageId: subscribeMessageId,
	}
	return c
}

func ParseBytesToUnsubscribeTlvModel(dsoLength uint16, unsubscribeByte []byte,
	offsetFromStart uint16) (tlvModel TlvModel, newOffsetFromStart uint16, err error) {

	if len(unsubscribeByte) < dnsutil.DSO_TYPE_UNSUBSCRIBE_LENGTH {
		belogs.Error("ParseBytesToUnsubscribeTlvModel(): recv byte's length is too small, fail: ",
			"   len(unsubscribeByte):", len(unsubscribeByte), "   DSO_TYPE_UNSUBSCRIBE_LENGTH:", dnsutil.DSO_TYPE_UNSUBSCRIBE_LENGTH)
		return nil, 0, errors.New("Received packet is too small for legal DSO Unsubscribe format")
	}
	if len(unsubscribeByte) < int(dsoLength) {
		belogs.Error("ParseBytesToUnsubscribeTlvModel(): recv byte's length is too small, fail: ",
			"   len(unsubscribeByte):", len(unsubscribeByte), "   dsoLength:", dsoLength)
		return nil, 0, errors.New("Received packet is too small for legal DSO Unsubscribe format")
	}

	subscribeMessageId := binary.BigEndian.Uint16(unsubscribeByte[:2])
	newOffsetFromStart = offsetFromStart + 2
	belogs.Info("ParseBytesToUnsubscribeTlvModel(): subscribeMessageId:", subscribeMessageId, "   newOffsetFromStart:", newOffsetFromStart)
	belogs.Info("#解析得到DSO的'取消订阅'类型数据包: " + osutil.GetNewLineSep() +
		"{'subscribeMessageId':" + convert.ToString(subscribeMessageId) + "}")
	return NewUnsubscribeTlvModel(subscribeMessageId), newOffsetFromStart, nil

}
func (c UnsubscribeTlvModel) Bytes() []byte {
	wr := bytes.NewBuffer([]byte{})
	binary.Write(wr, binary.BigEndian, c.DsoType)
	binary.Write(wr, binary.BigEndian, c.DsoLength)
	binary.Write(wr, binary.BigEndian, c.SubscribeMessageId)
	return wr.Bytes()
}

func (c UnsubscribeTlvModel) PrintBytes() string {
	return convert.PrintBytes(c.Bytes(), 8)
}

func (c UnsubscribeTlvModel) GetDsoType() uint16 {
	return dnsutil.DSO_TYPE_UNSUBSCRIBE
}
func NewDsoModelWithUnsubscribeTlvModel(subscribeMessageId uint16) (dsoModel *DsoModel, err error) {
	belogs.Debug("NewDsoModelWithUnsubscribeTlvModel(): subscribeMessageId:", subscribeMessageId)

	// messageId must be zero
	/* rfc8765
	In accordance with the definition of DSO unidirectional messages, the MESSAGE ID field MUST be zero.  There is no server response to an UNSUBSCRIBE message.
	*/
	dsoModel, err = NewDsoModelByParameters(0, dnsutil.DNS_QR_REQUEST, dnsutil.DNS_RCODE_NOERROR)
	if err != nil {
		belogs.Error("NewDsoModelWithUnsubscribeTlvModel(): NewDsoModelByParameters fail:", err)
		return nil, err
	}
	belogs.Debug("NewDsoModelWithUnsubscribeTlvModel(): NewDsoModelByParameters, dsoModel:", jsonutil.MarshalJson(dsoModel))

	unsubscirbeTlvModel := NewUnsubscribeTlvModel(subscribeMessageId)
	dsoModel.AddTlvModel(unsubscirbeTlvModel)
	belogs.Debug("NewDsoModelWithUnsubscribeTlvModel(): dsoModel:", jsonutil.MarshalJson(dsoModel))
	belogs.Info("#生成DSO的'取消订阅'类型数据包: " + osutil.GetNewLineSep() +
		"{'subscribeMessageId':" + convert.ToString(subscribeMessageId) + "}")
	return dsoModel, nil
}
