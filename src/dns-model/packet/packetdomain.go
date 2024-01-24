package packet

import (
	"bytes"
	"errors"

	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/convert"
	"github.com/cpusoft/goutil/jsonutil"
)

type PacketDomain struct {
	PacketLabels []*PacketLabel          `json:"packetLabels"`
	FullDomain   jsonutil.PrintableBytes `json:"fullDomain"`

	// all PacketLabels has no compression, then isCompression = false.
	IsCompression bool `json:"isCompression"`
}

func NewPacketDomainByAddPacketLabels(packetLabels []*PacketLabel,
	packetDecompressionLabel *PacketDecompressionLabel) (packetDomain *PacketDomain, err error) {
	belogs.Debug("NewPacketDomainByAddPacketLabels():packetLabels:", jsonutil.MarshalJson(packetLabels))
	isCompression := false
	fullDomain := make([]byte, 0)
	//offsetFromStartLabels := make(map[uint16]jsonutil.PrintableBytes, 0)

	// from back to front
	// if no compression, will get fullname, and each <offsetFromStart,lable> will save to packetDecompressionLabel
	// if is compression, will break first time.
	for i := len(packetLabels) - 1; i >= 0; i-- {
		if packetLabels[i].IsCompression {
			isCompression = true
			break
		}
		offsetFromStart := packetLabels[i].OffsetFromStart
		label := packetLabels[i].Label
		// append in 0
		labelTmp := make([]byte, 0)
		labelTmp = append(labelTmp, label...)
		if i < len(packetLabels)-1 {
			labelTmp = append(labelTmp, []byte(".")...)
		}
		fullDomain = append(labelTmp, fullDomain...)
		packetDecompressionLabel.Add(offsetFromStart, fullDomain)
		belogs.Debug("NewPacketDomainByAddPacketLabels():from back to front, i:", i,
			"  offsetFromStart:", offsetFromStart,
			"  fullDomain:", string(fullDomain),
			"  packetDecompressionLabel:", jsonutil.MarshalJson(packetDecompressionLabel))
	}

	if !isCompression {
		c := &PacketDomain{
			PacketLabels:  packetLabels,
			FullDomain:    fullDomain,
			IsCompression: isCompression,
			//OffsetFromStartLabels: offsetFromStartLabels,
		}
		belogs.Debug("NewPacketDomainByAddPacketLabels():not isCompression, c:", jsonutil.MarshalJson(c))
		return c, nil
	} else {
		// from front to back.
		// get from front name , then Find back name in packetDecompressionLabel to replace pointer
		belogs.Debug("NewPacketDomainByAddPacketLabels(): isCompression, len(packetLabels):", len(packetLabels))
		for i := 0; i < len(packetLabels); i++ {
			if packetLabels[i].IsCompression {
				decompressionLables := packetDecompressionLabel.Find(packetLabels[i].Pointer)
				belogs.Debug("NewPacketDomainByAddPacketLabels(): Pointer Find in packetDecompressionLabel, i:", i,
					"   pointer:", packetLabels[i].Pointer,
					"   decompressionLables:", string(decompressionLables),
					"   packetDecompressionLabel:", jsonutil.MarshalJson(packetDecompressionLabel))
				if len(decompressionLables) > 0 {
					fullDomain = append(fullDomain, decompressionLables...)
					belogs.Debug("NewPacketDomainByAddPacketLabels(): found in packetDecompressionLabel,i:", i, "  fullDomain:", string(fullDomain))
				} else {
					belogs.Error("NewPacketDomainByAddPacketLabels(): not found in packetDecompressionLabel fail,i:", i,
						"  packetLabels[i].Pointer:", packetLabels[i].Pointer, "  packetDecompressionLabel:", jsonutil.MarshalJson(packetDecompressionLabel))
					return nil, errors.New("new label is not found in packetDecompressionLabel")
				}
				break
			}
			label := packetLabels[i].Label
			fullDomain = append(fullDomain, label...)
			if i < len(packetLabels)-1 {
				fullDomain = append(fullDomain, []byte(".")...)
			}
			belogs.Debug("NewPacketDomainByAddPacketLabels():from front to back, i:", i, "  fullDomain:", string(fullDomain))
		}

		// will save whole fullDomain, save < Labels[0].OffsetFromStart:fullDomain>
		offsetFromStart := packetLabels[0].OffsetFromStart
		packetDecompressionLabel.Add(offsetFromStart, fullDomain)
		belogs.Debug("NewPacketDomainByAddPacketLabels():isCompression add packetDecompressionLabel, offsetFromStart:", offsetFromStart,
			"  fullDomain:", string(fullDomain), "  packetDecompressionLabel:", jsonutil.MarshalJson(packetDecompressionLabel))

		c := &PacketDomain{
			PacketLabels:  packetLabels,
			FullDomain:    fullDomain,
			IsCompression: isCompression,
		}
		belogs.Debug("NewPacketDomainByAddPacketLabels(): isCompression, c:", jsonutil.MarshalJson(c),
			"  packetDecompressionLabel:", jsonutil.MarshalJson(packetDecompressionLabel))

		return c, nil
	}

}

// no compression
func NewPacketDomainNoCompression(domain []byte, offsetFromStart uint16) (packetDomain *PacketDomain, newOffsetFromStart uint16, err error) {
	belogs.Debug("NewPacketDomainNoCompression(): domain :", convert.PrintBytesOneLine(domain), "  offsetFromStart:", offsetFromStart)
	if len(domain) == 0 {
		belogs.Error("NewPacketDomainNoCompression(): length of domain is empty, fail, len(domain):", len(domain))
		return nil, 0, errors.New("length of domain is empty")
	}
	packetLabels := make([]*PacketLabel, 0)
	sep := []byte(".")
	split := bytes.Split(domain, sep)
	belogs.Debug("NewPacketDomainNoCompression(): split:", split)
	newOffsetFromStart = offsetFromStart
	for i := range split {
		label := split[i]
		if len(label) == 0 {
			continue
		}
		packetLabel, err := NewPacketLabel(label, newOffsetFromStart)
		if err != nil {
			belogs.Error("NewPacketDomainNoCompression(): NewPacketLabel fail, label:",
				convert.PrintBytesOneLine(label), "    newOffsetFromStart:", newOffsetFromStart)
			return nil, 0, err
		}
		belogs.Debug("NewPacketDomainNoCompression():packetLabel:", jsonutil.MarshalJson(packetLabel),
			"   bytes:", convert.PrintBytesOneLine(packetLabel.Bytes()))
		packetLabels = append(packetLabels, packetLabel)
		if i == len(split)-1 {
			newOffsetFromStart += uint16(len(label))
		} else {
			newOffsetFromStart += uint16(len(label) + 1) // add "."
		}
		belogs.Debug("NewPacketDomainNoCompression():packetLabels:", jsonutil.MarshalJson(packetLabels),
			"   newOffsetFromStart:", newOffsetFromStart)
	}
	c := &PacketDomain{
		PacketLabels:  packetLabels,
		FullDomain:    domain,
		IsCompression: false,
	}
	belogs.Debug("NewPacketDomainNoCompression():packetDomain:", jsonutil.MarshalJson(c),
		"  newOffsetFromStart:", newOffsetFromStart, "  bytes:", convert.PrintBytesOneLine(c.Bytes()))
	return c, newOffsetFromStart, nil
}

//
func ParseBytesToPacketDomain(packetBytess []byte, offsetFromStart uint16, packetDecompressionLabel *PacketDecompressionLabel) (pakcetDomain *PacketDomain,
	newOffsetFromStart uint16, err error) {

	belogs.Debug("ParseBytesToPacketDomain():packetBytess:", convert.PrintBytesOneLine(packetBytess),
		"   offsetFromStart:", offsetFromStart, "  packetDecompressionLabel:", jsonutil.MarshalJson(packetDecompressionLabel))
	var tmpBytess, bytess []byte
	bytess = make([]byte, len(packetBytess))
	copy(bytess, packetBytess)
	packetLabels := make([]*PacketLabel, 0)
	for {
		packetLable, tmpStart, err :=
			ParseBytesToPacketLabel(bytess, offsetFromStart)
		if err != nil {
			belogs.Error("ParseBytesToPacketDomain(): ParseBytesToPacketLabel fail, bytess:",
				convert.PrintBytesOneLine(bytess), "   offsetFromStart:", offsetFromStart, err)
			return nil, 0, err
		}
		packetLabels = append(packetLabels, packetLable)
		belogs.Debug("ParseBytesToPacketDomain():ParseBytesToPacketLabel,  packetLable:", jsonutil.MarshalJson(packetLable),
			"   tmpStart:", tmpStart)

		// only one byte,cannot parse
		offsetLen := int(tmpStart - offsetFromStart)
		if offsetLen <= 1 {
			newOffsetFromStart = tmpStart
			belogs.Debug("ParseBytesToPacketDomain():break, tmpStart-offsetFromStart<= 1: ",
				"   offsetFromStart:", offsetFromStart, "  offsetLen:", offsetLen,
				"   packetLabels:", jsonutil.MarshalJson(packetLabels),
				"   tmpStart:", tmpStart, "  newOffsetFromStart:", newOffsetFromStart)

			break
		}
		belogs.Debug("ParseBytesToPacketDomain():tmpStart:", tmpStart, "   offsetFromStart:", offsetFromStart,
			"  offsetLen:", offsetLen)

		// label is compression
		if packetLable.IsCompression {
			newOffsetFromStart = tmpStart
			belogs.Debug("ParseBytesToPacketDomain():break, IsCompression:",
				"   packetLabels:", jsonutil.MarshalJson(packetLabels),
				"   tmpStart:", tmpStart, "  newOffsetFromStart:", newOffsetFromStart)

			break
		}

		// end with 0x00
		if len(bytess) > offsetLen && bytess[offsetLen] == 0x00 {
			tmpStart++ // +0x00
			newOffsetFromStart = tmpStart
			belogs.Debug("ParseBytesToPacketDomain():break, len(bytess) >offsetLen && bytess[len] == 0x00:",
				"   packetLabels:", jsonutil.MarshalJson(packetLabels),
				"   tmpStart:", tmpStart, "  newOffsetFromStart:", newOffsetFromStart)

			break
		}
		if len(bytess)-offsetLen == 0 {
			newOffsetFromStart = tmpStart
			belogs.Debug("ParseBytesToPacketDomain():break, len(bytess)-offsetLen == 0:",
				"   packetLabels:", jsonutil.MarshalJson(packetLabels),
				"   tmpStart:", tmpStart, "  newOffsetFromStart:", newOffsetFromStart)
			break
		}
		tmpBytess = make([]byte, len(bytess)-offsetLen)
		copy(tmpBytess, bytess[offsetLen:])
		bytess = tmpBytess
		offsetFromStart = tmpStart

		belogs.Debug("ParseBytesToPacketDomain():new loop,len(bytess):", len(bytess), "   bytess:", convert.PrintBytesOneLine(bytess), "   offsetFromStart:", offsetFromStart)
	}
	c, err := NewPacketDomainByAddPacketLabels(packetLabels, packetDecompressionLabel)
	if err != nil {
		belogs.Error("ParseBytesToPacketDomain(): NewPacketDomainByAddPacketLabels fail, bytess:",
			convert.PrintBytesOneLine(bytess), "   packetLabels:", jsonutil.MarshalJson(packetLabels),
			"  packetDecompressionLabel:", jsonutil.MarshalJson(packetDecompressionLabel), err)
		return nil, 0, err
	}
	belogs.Debug("ParseBytesToPacketDomain():packetDomain:", jsonutil.MarshalJson(c), "  newOffsetFromStart:", newOffsetFromStart,
		"   packetDecompressionLabel:", jsonutil.MarshalJson(packetDecompressionLabel))
	return c, newOffsetFromStart, nil

}
func (c *PacketDomain) Bytes() []byte {
	bufs := make([]byte, 0)
	for i := range c.PacketLabels {
		buf := c.PacketLabels[i].Bytes()
		bufs = append(bufs, buf...)
	}
	// when no compression, add 0x00
	if !c.PacketLabels[len(c.PacketLabels)-1].IsCompression {
		bufs = append(bufs, 0x00)
	}
	belogs.Debug("PacketDomain.Bytes():PacketDomain, bufs:", convert.PrintBytesOneLine(bufs))
	return bufs
}
func (c *PacketDomain) Length() uint16 {
	length := uint16(0)
	for i := range c.PacketLabels {
		length += c.PacketLabels[i].Length()
	}
	// when no compression, add 0x00
	if !c.PacketLabels[len(c.PacketLabels)-1].IsCompression {
		length += 1
	}
	return length
}

type PacketDomains struct {
	PacketDomains []*PacketDomain `json:"packetDomains"`
}
