package packet

import (
	"bytes"
	"encoding/binary"
	"errors"
	"strconv"

	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/convert"
	"github.com/cpusoft/goutil/jsonutil"
)

type SrvModel struct {
	Priority           uint16                  `json:"priority"`
	Weight             uint16                  `json:"weight"`
	Port               uint16                  `json:"port"`
	TargetPacketDomain *PacketDomain           `json:"targetPacketDomain"`
	Target             jsonutil.PrintableBytes `json:"target"`
}

func NewSrvModel(priorityStr, weightStr, portStr, targetStr string, offsetFromStart uint16) (srvModel *SrvModel, newOffsetFromStart uint16, err error) {
	belogs.Debug("NewSrvModel(): priorityStr:", priorityStr, "   weightStr:", weightStr, "   portStr:", portStr,
		"  targetStr:", targetStr, "   offsetFromStart:", offsetFromStart)

	priority, err := strconv.Atoi(priorityStr)
	if err != nil {
		belogs.Error("NewSrvModel():priorityStr Atoi fail, priorityStr:", priorityStr, err)
		return nil, 0, err
	}
	newOffsetFromStart = offsetFromStart + 2

	weight, err := strconv.Atoi(weightStr)
	if err != nil {
		belogs.Error("NewSrvModel():weightStr Atoi fail, weightStr:", weightStr, err)
		return nil, 0, err
	}
	newOffsetFromStart = offsetFromStart + 2

	port, err := strconv.Atoi(portStr)
	if err != nil {
		belogs.Error("NewSrvModel():portStr Atoi fail, portStr:", portStr, err)
		return nil, 0, err
	}
	newOffsetFromStart = offsetFromStart + 2

	targetPacketDomain, newOffsetFromStartTmp, err := NewPacketDomainNoCompression([]byte(targetStr), newOffsetFromStart)
	if err != nil {
		belogs.Error("NewSrvModel(): ParseBytesToPacketDomain fail, targetStr:", targetStr, err)
		return nil, 0, err
	}
	newOffsetFromStart = newOffsetFromStartTmp
	c := &SrvModel{
		Priority:           uint16(priority),
		Weight:             uint16(weight),
		Port:               uint16(port),
		TargetPacketDomain: targetPacketDomain,
		Target:             []byte(targetStr),
	}
	belogs.Debug("NewSrvModel():  ptrModel:", jsonutil.MarshalJson(c), "   newOffsetFromStart:", newOffsetFromStart)
	return c, newOffsetFromStart, nil
}

// currently not support compression format
func ParseBytesToSrvModel(bytess []byte, offsetFromStart uint16, packetDecompressionLabel *PacketDecompressionLabel) (srvModel *SrvModel, newOffsetFromStart uint16, err error) {
	belogs.Debug("ParseBytesToSrvModel(): bytess:", convert.PrintBytesOneLine(bytess), "   offsetFromStart:", offsetFromStart)
	numbersLen := 3 * 2 // 3*int16
	if len(bytess) < numbersLen {
		belogs.Error("ParseBytesToSrvModel(): len(bytess):", len(bytess), "   numbersLen:", numbersLen)
		return nil, 0, errors.New("the length of srv is too short")
	}

	namesLen := len(bytess) - numbersLen
	numberBytes := bytess[:numbersLen]
	namesBytes := bytess[numbersLen:]

	belogs.Debug("ParseBytesToSrvModel():numbersLen:", numbersLen, " namesLen:", namesLen,
		"    numberBytes:", convert.PrintBytesOneLine(numberBytes), "   namesBytes:", convert.PrintBytesOneLine(namesBytes))

	// numbers
	buf := bytes.NewReader(numberBytes)
	var priority uint16
	err = binary.Read(buf, binary.BigEndian, &priority)
	if err != nil {
		belogs.Error("ParseBytesToSrvModel():priority Read buf fail,", convert.PrintBytesOneLine(numberBytes), err)
		return nil, 0, err
	}
	newOffsetFromStart = offsetFromStart + 2
	belogs.Debug("ParseBytesToSrvModel(): priority:", priority, "   newOffsetFromStart:", newOffsetFromStart)

	var weight uint16
	err = binary.Read(buf, binary.BigEndian, &weight)
	if err != nil {
		belogs.Error("ParseBytesToSrvModel():weight Read buf fail,", convert.PrintBytesOneLine(numberBytes), err)
		return nil, 0, err
	}
	newOffsetFromStart = newOffsetFromStart + 2
	belogs.Debug("ParseBytesToSrvModel(): weight:", weight, "   newOffsetFromStart:", newOffsetFromStart)

	var port uint16
	err = binary.Read(buf, binary.BigEndian, &port)
	if err != nil {
		belogs.Error("ParseBytesToSrvModel():port Read buf fail,", convert.PrintBytesOneLine(numberBytes), err)
		return nil, 0, err
	}
	newOffsetFromStart = newOffsetFromStart + 2
	belogs.Debug("ParseBytesToSrvModel(): port:", port, "   newOffsetFromStart:", newOffsetFromStart)

	// target: packetDomain
	targetPacketDomain, newOffsetFromStart, err := ParseBytesToPacketDomain(namesBytes, newOffsetFromStart, packetDecompressionLabel)
	if err != nil {
		belogs.Error("ParseBytesToSrvModel():ParseBytesToPacketDomain fail,", convert.PrintBytesOneLine(namesBytes), err)
		return nil, 0, err
	}
	c := &SrvModel{
		Priority:           priority,
		Weight:             weight,
		Port:               port,
		TargetPacketDomain: targetPacketDomain,
		Target:             namesBytes,
	}
	newOffsetFromStart = offsetFromStart + uint16(len(bytess))
	belogs.Debug("ParseBytesToSrvModel(): bytes:", convert.PrintBytesOneLine(bytess),
		"   srvModel:", jsonutil.MarshalJson(c),
		"   newOffsetFromStart:", newOffsetFromStart)
	return c, newOffsetFromStart, nil
}

func (c *SrvModel) Bytes() []byte {
	wr := bytes.NewBuffer([]byte{})
	binary.Write(wr, binary.BigEndian, c.Priority)
	binary.Write(wr, binary.BigEndian, c.Weight)
	binary.Write(wr, binary.BigEndian, c.Port)
	binary.Write(wr, binary.BigEndian, c.TargetPacketDomain.Bytes())
	return wr.Bytes()
}
func (c *SrvModel) Length() uint16 {
	return uint16(2 + 2 + 2 + c.TargetPacketDomain.Length())
}
func (c *SrvModel) ToRrData() string {
	return convert.ToString(c.Priority) + " " + convert.ToString(c.Weight) +
		convert.ToString(c.Port) + " " + string(c.TargetPacketDomain.FullDomain)
}
