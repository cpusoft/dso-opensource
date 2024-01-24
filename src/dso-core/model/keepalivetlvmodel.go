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
// keepalive
type KeepaliveTlvModel struct {
	DsoType           uint16 `json:"dsoType"`
	DsoLength         uint16 `json:"dsoLength"`
	InactivityTimeout uint32 `json:"inactivityTimeout"`
	KeepaliveInterval uint32 `json:"keepaliveInterval"`
}

func NewKeepaliveTlvModel(inactivityTimeout, keepaliveInterval uint32) *KeepaliveTlvModel {
	c := &KeepaliveTlvModel{
		DsoType:           dnsutil.DSO_TYPE_KEEPALIVE,
		DsoLength:         dnsutil.DSO_TYPE_KEEPALIVE_LENGTH,
		InactivityTimeout: inactivityTimeout,
		KeepaliveInterval: keepaliveInterval,
	}

	return c
}

func NewDsoModelWithKeepaliveTlvModel(messageId uint16, qr uint8, rCode uint8,
	inactivityTimeout, keepaliveInterval uint32) *DsoModel {
	belogs.Debug("NewDsoModelWithKeepaliveTlvModel():messageId:", messageId, "  qr:", qr, "  rCode:", rCode,
		"  inactivityTimeout:", inactivityTimeout, "   keepaliveInterval:", keepaliveInterval)

	dsoModel, _ := NewDsoModelByParameters(messageId, qr, rCode)
	keepaliveModel := NewKeepaliveTlvModel(inactivityTimeout, keepaliveInterval)
	dsoModel.AddTlvModel(keepaliveModel)
	belogs.Info("#生成DSO的'建立会话'和'会话保持'类型数据包: " + osutil.GetNewLineSep() + "{'InactivityTimeout':" + convert.ToString(inactivityTimeout) +
		",'KeepaliveInterval':" + convert.ToString(keepaliveInterval) + "}")
	return dsoModel
}

func ParseBytesToKeepaliveTlvModel(dsoLength uint16, keepaliveBytes []byte,
	offsetFromStart uint16) (tlvModel TlvModel, newOffsetFromStart uint16, err error) {

	if len(keepaliveBytes) < dnsutil.DSO_TYPE_KEEPALIVE_LENGTH {
		belogs.Error("ParseBytesToKeepaliveTlvModel(): recv byte's length is too small, fail:",
			"  len(keepaliveBytes):", len(keepaliveBytes), "  DSO_TYPE_KEEPALIVE_LENGTH:", dnsutil.DSO_TYPE_KEEPALIVE_LENGTH)
		return nil, 0, errors.New("Received packet is too small for legal DSO Keepalive format")
	}
	if len(keepaliveBytes) < int(dsoLength) {
		belogs.Error("ParseBytesToKeepaliveTlvModel(): recv byte's length is too small, fail:",
			"  len(keepaliveBytes):", len(keepaliveBytes), "  dsoLength:", dsoLength)
		return nil, 0, errors.New("Received packet is too small for legal DSO Keepalive format")
	}

	inactivityTimeout := binary.BigEndian.Uint32(keepaliveBytes[:4])
	newOffsetFromStart = offsetFromStart + 4
	belogs.Debug("ParseBytesToKeepaliveTlvModel(): inactivityTimeout:", inactivityTimeout, "  newOffsetFromStart:", newOffsetFromStart)

	keepaliveInterval := binary.BigEndian.Uint32(keepaliveBytes[4:8])
	newOffsetFromStart += 4
	belogs.Info("ParseBytesToKeepaliveTlvModel(): inactivityTimeout:", inactivityTimeout, "  keepaliveInterval:", keepaliveInterval, "  newOffsetFromStart:", newOffsetFromStart)
	belogs.Info("#解析得到DSO的'建立会话'和'会话保持'类型数据包: " + osutil.GetNewLineSep() + "{'InactivityTimeout':" + convert.ToString(inactivityTimeout) +
		",'KeepaliveInterval':" + convert.ToString(keepaliveInterval) + "}")
	return NewKeepaliveTlvModel(inactivityTimeout, keepaliveInterval), newOffsetFromStart, nil
}

func (c KeepaliveTlvModel) Bytes() []byte {
	wr := bytes.NewBuffer([]byte{})
	binary.Write(wr, binary.BigEndian, c.DsoType)
	binary.Write(wr, binary.BigEndian, c.DsoLength)
	binary.Write(wr, binary.BigEndian, c.InactivityTimeout)
	binary.Write(wr, binary.BigEndian, c.KeepaliveInterval)
	return wr.Bytes()
}

func (c KeepaliveTlvModel) PrintBytes() string {
	return convert.PrintBytes(c.Bytes(), 8)
}
func (c KeepaliveTlvModel) GetDsoType() uint16 {
	return dnsutil.DSO_TYPE_KEEPALIVE
}
