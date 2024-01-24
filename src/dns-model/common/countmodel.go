package common

import (
	"bytes"
	"encoding/binary"
	"errors"

	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/convert"
	"github.com/cpusoft/goutil/jsonutil"
)

const (
	DNS_COUNT_LENGTH    = 8
	DNS_COUNT_TYPE_QANA = 1
	DNS_COUNT_TYPE_ZPUA = 2
)

type CountModel interface {
	Bytes() []byte
	GetCount(index int) uint16

	GetCountType() int
}
type CountQANAModel struct {
	QdCount uint16 `json:"qdCount"`
	AnCount uint16 `json:"anCount"`
	NsCount uint16 `json:"nsCount"`
	ArCount uint16 `json:"arCount"`
}

type CountZPUAModel struct {
	ZoCount uint16 `json:"zoCount"`
	PrCount uint16 `json:"prCount"`
	UpCount uint16 `json:"upCount"`
	AdCount uint16 `json:"adCount"`
}

func NewCountModel(count1, count2, count3, count4 uint16, countType int) (countModel CountModel, err error) {
	if countType == DNS_COUNT_TYPE_QANA {
		c := &CountQANAModel{
			QdCount: count1,
			AnCount: count2,
			NsCount: count3,
			ArCount: count4,
		}
		return c, nil
	} else if countType == DNS_COUNT_TYPE_ZPUA {
		c := &CountZPUAModel{
			ZoCount: count1,
			PrCount: count2,
			UpCount: count3,
			AdCount: count4,
		}
		return c, nil
	}
	return nil, errors.New("not support count type")
}

func ParseBytesToCountModel(countBytes []byte, headerType int) (countModel CountModel, err error) {
	if len(countBytes) != DNS_COUNT_LENGTH {
		belogs.Error("ParseBytesToCountModel(): len(countBytes) < DNS_COUNT_LENGTH:", len(countBytes))
		return nil, errors.New("Counter Length is illegal")
	}
	var countType int
	if headerType == DNS_HEADER_TYPE_QUERY || headerType == DNS_HEADER_TYPE_DSO {
		countType = DNS_COUNT_TYPE_QANA
	} else if headerType == DNS_HEADER_TYPE_UPDATE {
		countType = DNS_COUNT_TYPE_ZPUA
	} else {
		belogs.Error("ParseBytesToCountModel(): headerType is fail, headerType:", headerType)
		return nil, errors.New("Counter type is illegal")
	}
	belogs.Debug("ParseBytesToCountModel():countBytes: ", convert.PrintBytesOneLine(countBytes))
	buf := bytes.NewReader(countBytes)

	// get count1,2,3,4
	var count1 uint16
	err = binary.Read(buf, binary.BigEndian, &count1)
	if err != nil {
		belogs.Error("ParseBytesToCountModel():get count1 fail, countBytes: ", convert.PrintBytesOneLine(countBytes))
		return nil, errors.New("Fail to get count1:" + err.Error())
	}
	belogs.Debug("ParseBytesToCountModel(): count1:", count1)

	var count2 uint16
	err = binary.Read(buf, binary.BigEndian, &count2)
	if err != nil {
		belogs.Error("ParseBytesToCountModel():get count1 fail, countBytes: ", convert.PrintBytesOneLine(countBytes))
		return nil, errors.New("Fail to get count2:" + err.Error())
	}
	belogs.Debug("ParseBytesToCountModel(): count2:", count2)

	var count3 uint16
	err = binary.Read(buf, binary.BigEndian, &count3)
	if err != nil {
		belogs.Error("ParseBytesToCountModel():get count1 fail, countBytes: ", convert.PrintBytesOneLine(countBytes))
		return nil, errors.New("Fail to get count3:" + err.Error())
	}
	belogs.Debug("ParseBytesToCountModel(): count3:", count3)

	var count4 uint16
	err = binary.Read(buf, binary.BigEndian, &count4)
	if err != nil {
		belogs.Error("ParseBytesToCountModel():get count1 fail, countBytes: ", convert.PrintBytesOneLine(countBytes))
		return nil, errors.New("Fail to get count4:" + err.Error())
	}
	belogs.Debug("ParseBytesToCountModel(): count4:", count4)

	c, err := NewCountModel(count1, count2, count3, count4, countType)
	if err != nil {
		belogs.Error("ParseBytesToCountModel():NewCountModel fail, count1:", count1, " count2:", count2,
			"  count4:", count4, "  count4:", count4, "  countType:", countType)
		return nil, errors.New("NewCountModel fail")
	}

	belogs.Info("ParseBytesToCountModel(): NewCountModel:", jsonutil.MarshalJson(c))
	return c, nil
}

func (c CountQANAModel) Bytes() []byte {
	wr := bytes.NewBuffer([]byte{})
	binary.Write(wr, binary.BigEndian, c.QdCount)
	binary.Write(wr, binary.BigEndian, c.AnCount)
	binary.Write(wr, binary.BigEndian, c.NsCount)
	binary.Write(wr, binary.BigEndian, c.ArCount)
	return wr.Bytes()
}
func (c CountQANAModel) GetCount(index int) uint16 {
	switch index {
	case 0:
		return c.QdCount
	case 1:
		return c.AnCount
	case 2:
		return c.NsCount
	case 3:
		return c.ArCount
	default:
		return 0
	}
}

func (c CountQANAModel) GetCountType() int {
	return DNS_COUNT_TYPE_QANA
}
func (c CountZPUAModel) Bytes() []byte {
	wr := bytes.NewBuffer([]byte{})
	binary.Write(wr, binary.BigEndian, c.ZoCount)
	binary.Write(wr, binary.BigEndian, c.PrCount)
	binary.Write(wr, binary.BigEndian, c.UpCount)
	binary.Write(wr, binary.BigEndian, c.AdCount)
	return wr.Bytes()
}
func (c CountZPUAModel) GetCountType() int {
	return DNS_COUNT_TYPE_ZPUA
}
func (c CountZPUAModel) GetCount(index int) uint16 {
	switch index {
	case 0:
		return c.ZoCount
	case 1:
		return c.PrCount
	case 2:
		return c.UpCount
	case 3:
		return c.AdCount
	default:
		return 0
	}
}
