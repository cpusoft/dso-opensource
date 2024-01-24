package packet

import (
	"bytes"
	"encoding/binary"
	"errors"
	"strconv"
	"strings"

	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/convert"
	"github.com/cpusoft/goutil/jsonutil"
)

type SoaModel struct {
	MNamePacketDomain *PacketDomain           `json:"mNamePacketDomain"`
	MName             jsonutil.PrintableBytes `json:"mName"`
	RNamePacketDomain *PacketDomain           `json:"rNamePacketDomain"`
	RName             jsonutil.PrintableBytes `json:"rName"`
	Serial            uint32                  `json:"serial"`
	Refresh           int32                   `json:"refresh"`
	Retry             int32                   `json:"retry"`
	Expire            int32                   `json:"expire"`
	Minimum           uint32                  `json:"minimum"`
}

func NewSoaModelFromRrData(rrData string, offsetFromStart uint16) (soaModel *SoaModel, newOffsetFromStart uint16, err error) {
	belogs.Debug("NewSoaModel(): rrData:", rrData)
	rrData = strings.Replace(strings.Replace(rrData, "(", "", -1), ")", "", -1)
	belogs.Debug("NewSoaModel(): after replace, rrData:", rrData)
	split := strings.Split(rrData, " ")
	if len(split) == 7 {
		return NewSoaModel(split[0], split[1], split[2], split[3], split[4], split[5], split[6], offsetFromStart)
	}
	return nil, 0, errors.New("rrData is format error for SOA")
}
func NewSoaModel(mNameStr, rNameStr, serialStr, refreshStr, retryStr, expireStr, minimumStr string,
	offsetFromStart uint16) (soaModel *SoaModel, newOffsetFromStart uint16, err error) {
	belogs.Debug("NewSoaModel(): mNameStr:", mNameStr, "   serialStr:", serialStr, "   refreshStr:", refreshStr, "  retryStr:", retryStr,
		"  expireStr:", expireStr, "  minimumStr:", minimumStr, "   offsetFromStart:", offsetFromStart)

	mNamePacketDomain, newOffsetFromStart, err := NewPacketDomainNoCompression([]byte(mNameStr), offsetFromStart)
	if err != nil {
		belogs.Error("NewPtrModel(): ParseBytesToPacketDomain fail, mNameStr:", mNameStr, err)
		return nil, 0, err
	}
	belogs.Debug("NewSoaModel(): mNamePacketDomain:", jsonutil.MarshalJson(mNamePacketDomain), "   newOffsetFromStart:", newOffsetFromStart)

	rNamePacketDomain, newOffsetFromStartTmp, err := NewPacketDomainNoCompression([]byte(rNameStr), newOffsetFromStart)
	if err != nil {
		belogs.Error("NewPtrModel(): ParseBytesToPacketDomain fail, rNameStr:", rNameStr, err)
		return nil, 0, err
	}
	newOffsetFromStart = newOffsetFromStartTmp
	belogs.Debug("NewSoaModel(): rNamePacketDomain:", jsonutil.MarshalJson(rNamePacketDomain), "   newOffsetFromStart:", newOffsetFromStart)

	serial, err := strconv.Atoi(serialStr)
	if err != nil {
		belogs.Error("NewSoaModel():serial atoi fail,", serialStr, err)
		return nil, 0, err
	}
	newOffsetFromStart += 4

	refresh, err := strconv.Atoi(refreshStr)
	if err != nil {
		belogs.Error("NewSoaModel():refresh atoi fail,", refreshStr, err)
		return nil, 0, err
	}
	newOffsetFromStart += 4

	retry, err := strconv.Atoi(retryStr)
	if err != nil {
		belogs.Error("NewSoaModel():retry atoi fail,", retryStr, err)
		return nil, 0, err
	}
	expire, err := strconv.Atoi(expireStr)
	if err != nil {
		belogs.Error("NewSoaModel():expire atoi fail,", expireStr, err)
		return nil, 0, err
	}
	newOffsetFromStart += 4

	minimum, err := strconv.Atoi(minimumStr)
	if err != nil {
		belogs.Error("NewSoaModel():minimum atoi fail,", minimumStr, err)
		return nil, 0, err
	}
	newOffsetFromStart += 4

	c := &SoaModel{
		MNamePacketDomain: mNamePacketDomain,
		MName:             []byte(mNameStr),
		RNamePacketDomain: rNamePacketDomain,
		RName:             []byte(rNameStr),
		Serial:            uint32(serial),
		Refresh:           int32(refresh),
		Retry:             int32(retry),
		Expire:            int32(expire),
		Minimum:           uint32(minimum),
	}
	belogs.Debug("NewSoaModel():  soaModel:", jsonutil.MarshalJson(c), "  newOffsetFromStart:", newOffsetFromStart)
	return c, newOffsetFromStart, nil
}

// currently not support compression format
func ParseBytesToSoaModel(bytess []byte, offsetFromStart uint16, packetDecompressionLabel *PacketDecompressionLabel) (soaModel *SoaModel, newOffsetFromStart uint16, err error) {
	numbersLen := 5 * 4 // 5*int32
	if len(bytess) < numbersLen {
		belogs.Error("ParseBytesToSoaModel(): len(bytess):", len(bytess), "   numbersLen:", numbersLen)
		return nil, 0, errors.New("Illegal SOA format")
	}

	namesLen := len(bytess) - numbersLen
	namesBytes := bytess[:namesLen]
	numberBytes := bytess[namesLen:]
	belogs.Debug("ParseBytesToSoaModel(): namesLen:", namesLen,
		"   namesBytes:", convert.PrintBytesOneLine(namesBytes), "    numberBytes:", convert.PrintBytesOneLine(numberBytes))

	// found 0x00, then mName/rName
	mNamePacketDomain, newOffsetFromStart, err := ParseBytesToPacketDomain(namesBytes, offsetFromStart, packetDecompressionLabel)
	if err != nil {
		belogs.Error("ParseBytesToSoaModel():ParseBytesToPacketDomain mNamePacketDomain fail, namesBytes:", convert.PrintBytesOneLine(namesBytes), err)
		return nil, 0, err
	}
	mNameLenght := newOffsetFromStart - offsetFromStart
	mName := namesBytes[:mNameLenght]
	belogs.Debug("ParseBytesToSoaModel(): mNamePacketDomain:", jsonutil.MarshalJson(mNamePacketDomain),
		"   newOffsetFromStart:", newOffsetFromStart,
		"   mNameLenght:", mNameLenght, "  mName:", convert.PrintBytesOneLine(mName))
	rNameBytes := namesBytes[mNameLenght:]
	rNamePacketDomain, newOffsetFromStart, err := ParseBytesToPacketDomain(rNameBytes, newOffsetFromStart, packetDecompressionLabel)
	if err != nil {
		belogs.Error("ParseBytesToSoaModel():ParseBytesToPacketDomain rNamePacketDomain fail, rNameBytes:", convert.PrintBytesOneLine(rNameBytes), err)
		return nil, 0, err
	}
	rName := namesBytes[mNameLenght:]
	belogs.Debug("ParseBytesToSoaModel(): rNamePacketDomain:", jsonutil.MarshalJson(rNamePacketDomain),
		"   newOffsetFromStart:", newOffsetFromStart,
		"   rName:", convert.PrintBytesOneLine(rName))

	// numbers
	buf := bytes.NewReader(numberBytes)
	var serial uint32
	err = binary.Read(buf, binary.BigEndian, &serial)
	if err != nil {
		belogs.Error("ParseBytesToSoaModel():serial Read buf fail,", convert.PrintBytesOneLine(numberBytes), err)
		return nil, 0, err
	}
	newOffsetFromStart += 4
	belogs.Debug("ParseBytesToSoaModel(): serial:", serial, "  newOffsetFromStart:", newOffsetFromStart)

	var refresh int32
	err = binary.Read(buf, binary.BigEndian, &refresh)
	if err != nil {
		belogs.Error("ParseBytesToSoaModel():refresh Read buf fail,", convert.PrintBytesOneLine(numberBytes), err)
		return nil, 0, err
	}
	newOffsetFromStart += 4
	belogs.Debug("ParseBytesToSoaModel(): refresh:", refresh, "  newOffsetFromStart:", newOffsetFromStart)

	var retry int32
	err = binary.Read(buf, binary.BigEndian, &retry)
	if err != nil {
		belogs.Error("ParseBytesToSoaModel():retry Read buf fail,", convert.PrintBytesOneLine(numberBytes), err)
		return nil, 0, err
	}
	newOffsetFromStart += 4
	belogs.Debug("ParseBytesToSoaModel(): retry:", retry, "  newOffsetFromStart:", newOffsetFromStart)

	var expire int32
	err = binary.Read(buf, binary.BigEndian, &expire)
	if err != nil {
		belogs.Error("ParseBytesToSoaModel():expire Read buf fail,", convert.PrintBytesOneLine(numberBytes), err)
		return nil, 0, err
	}
	newOffsetFromStart += 4
	belogs.Debug("ParseBytesToSoaModel(): expire:", expire, "  newOffsetFromStart:", newOffsetFromStart)

	var minimum uint32
	err = binary.Read(buf, binary.BigEndian, &minimum)
	if err != nil {
		belogs.Error("ParseBytesToSoaModel():minimum Read buf fail,", convert.PrintBytesOneLine(numberBytes), err)
		return nil, 0, err
	}
	newOffsetFromStart += 4
	belogs.Debug("ParseBytesToSoaModel(): minimum:", minimum, "  newOffsetFromStart:", newOffsetFromStart)

	c := &SoaModel{
		MNamePacketDomain: mNamePacketDomain,
		MName:             mName,
		RNamePacketDomain: rNamePacketDomain,
		RName:             rName,
		Serial:            serial,
		Refresh:           refresh,
		Retry:             retry,
		Expire:            expire,
		Minimum:           minimum,
	}
	belogs.Debug("ParseBytesToSoaModel(): bytes:", convert.PrintBytesOneLine(bytess),
		"   soaModel:", jsonutil.MarshalJson(c))
	return c, newOffsetFromStart, nil
}

func (c *SoaModel) Bytes() []byte {
	wr := bytes.NewBuffer([]byte{})
	binary.Write(wr, binary.BigEndian, c.MNamePacketDomain.Bytes())
	binary.Write(wr, binary.BigEndian, c.RNamePacketDomain.Bytes())
	binary.Write(wr, binary.BigEndian, c.Serial)
	binary.Write(wr, binary.BigEndian, c.Refresh)
	binary.Write(wr, binary.BigEndian, c.Retry)
	binary.Write(wr, binary.BigEndian, c.Expire)
	binary.Write(wr, binary.BigEndian, c.Minimum)
	return wr.Bytes()
}
func (c *SoaModel) Length() uint16 {
	// []byte + 0x00+ []byte +0x00+ uint32*5
	return uint16(c.MNamePacketDomain.Length() + c.RNamePacketDomain.Length() + 4 + 4 + 4 + 4 + 4)
}
func (c *SoaModel) ToRrData() string {
	return string(c.MNamePacketDomain.FullDomain) + " " + string(c.RNamePacketDomain.FullDomain) + " ( " +
		convert.ToString(c.Serial) + " " + convert.ToString(c.Refresh) +
		convert.ToString(c.Retry) + " " + convert.ToString(c.Expire) +
		convert.ToString(c.Minimum) + " ) "
}
