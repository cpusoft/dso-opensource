package packet

import (
	"encoding/binary"
	"errors"

	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/convert"
	"github.com/cpusoft/goutil/dnsutil"
	"github.com/cpusoft/goutil/jsonutil"
)

type PacketLabel struct {
	IsCompression bool `json:"isCompression"`
	// if isCompression=true
	Pointer uint16 `json:"pointer"`

	// if isCompression=false
	LabelLength uint8                   `json:"labelLength"`
	Label       jsonutil.PrintableBytes `json:"label"`

	OffsetFromStart uint16 `json:"offsetFromStart"` // from all dnsbytes offset
}

func NewPacketLabel(label []byte, offsetFromStart uint16) (*PacketLabel, error) {
	belogs.Debug("NewPacketLabel(): label:", convert.PrintBytesOneLine(label),
		"   offsetFromStart:", offsetFromStart)
	if len(label) == 0 || uint8(len(label)) > dnsutil.DNS_DOMAIN_ONE_LABEL_MAXLENGTH {
		belogs.Error("NewPacketLabel(): length of label is empty or too long, fail, len(label):", len(label))
		return nil, errors.New("length of label is empty or too long")
	}

	c := &PacketLabel{
		IsCompression:   false,
		Pointer:         0,
		LabelLength:     uint8(len(label)),
		Label:           label,
		OffsetFromStart: offsetFromStart,
	}
	belogs.Debug("NewPacketLabel(): packetLabel:", jsonutil.MarshalJson(c), "   bytes:", convert.PrintBytesOneLine(c.Bytes()))
	return c, nil
}

func ParseBytesToPacketLabel(bytess []byte, offsetFromStart uint16) (packetLable *PacketLabel,
	newOffsetFromStart uint16, err error) {
	isCompression, pointer, labelLength, label, newOffsetFromStart, err :=
		dnsutil.CheckDomainCompressionPointer(bytess, offsetFromStart)
	if err != nil {
		belogs.Error("ParseBytesToPacketLabel(): CheckDomainCompressionPointer fail, bytess:",
			convert.PrintBytesOneLine(bytess))
		return nil, 0, err
	}
	belogs.Debug("ParseBytesToPacketLabel(): isCompression :", isCompression, "  pointer:", pointer,
		"   labelLength:", labelLength, "   label:", convert.PrintBytesOneLine(label),
		"   newOffsetFromStart:", newOffsetFromStart)

	c := &PacketLabel{
		IsCompression:   isCompression,
		Pointer:         pointer,
		LabelLength:     labelLength,
		Label:           label,
		OffsetFromStart: offsetFromStart,
	}
	belogs.Debug("ParseBytesToPacketLabel():packetLabel:", jsonutil.MarshalJson(c),
		"   newOffsetFromStart:", newOffsetFromStart, "   bytes:", convert.PrintBytesOneLine(c.Bytes()))
	return c, newOffsetFromStart, nil
}
func (c *PacketLabel) Bytes() []byte {
	if c.IsCompression {
		offset := uint16(c.Pointer | dnsutil.DNS_DOMAIN_COMPRESSION_POINTER)
		buf := make([]byte, 2)
		binary.BigEndian.PutUint16(buf, offset)
		belogs.Debug("PacketLabel.Bytes():IsCompression PacketLabel, buf:", convert.PrintBytesOneLine(buf))
		return buf
	} else {
		buf := make([]byte, (1 + len(c.Label)))
		buf[0] = c.LabelLength
		copy(buf[1:], c.Label)
		belogs.Debug("PacketLabel.Bytes():!IsCompression PacketLabel, buf:", convert.PrintBytesOneLine(buf))
		return buf
	}
}
func (c *PacketLabel) Length() uint16 {
	if c.IsCompression {
		return 2
	} else {
		return 1 + uint16(c.LabelLength)
	}
}
