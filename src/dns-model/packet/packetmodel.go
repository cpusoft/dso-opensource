package packet

import (
	"bytes"
	"encoding/binary"
	"errors"

	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/convert"
	"github.com/cpusoft/goutil/dnsutil"
	"github.com/cpusoft/goutil/jsonutil"
)

const (
	// only name/type/class
	DNS_PACKET_NAME_TYPE_CLASS = 1
	// name/type/class/rdata
	DNS_PACKET_NAME_TYPE_CLASS_RDATA = 2
	// name/type/class/ttl/rdlength  ttl=0, rdlength=0
	DNS_PACKET_NAME_TYPE_CLASS_TTL_RDLENGTH = 3
	// name/type/class/ttl/rdlength/rdata
	DNS_PACKET_NAME_TYPE_CLASS_TTL_RDLENGTH_RDATA = 4

	// only name/type/class
	DNS_PACKET_UPDATA_ZONE         = DNS_PACKET_NAME_TYPE_CLASS
	DNS_PACKET_UPDATA_PREREQUISITE = DNS_PACKET_NAME_TYPE_CLASS_TTL_RDLENGTH
)

// for packet
type PacketModel struct {

	//  domain = name + origin (will remove the end ".")  //
	// NameStr: dwn.roo.bo --> Name: 03 64 77 6e 03 72 6f 6f 02 62 6f 00
	PacketDomain      *PacketDomain     `json:"packetDomain"`
	PacketDomainBytes jsonutil.HexBytes `json:"packetDomainBytes"`
	PacketType        uint16            `json:"packetType"`  // 'type' is keyword in golang, so use dnsType
	PacketClass       uint16            `json:"packetClass"` //
	PacketTtl         uint32            `json:"packetTtl"`

	// == len(PacketRData)

	PacketDataLength uint16            `json:"packetDataLength"`
	PacketData       interface{}       `json:"packetData"`
	PacketDataBytes  jsonutil.HexBytes `json:"packetBytes"`

	// true: will ignore ttl/dataLength :RFC8765 6.5.1. RECONFIRM Message
	PacketModelType int `json:"packetModelType"`
}

func NewPacketModel(packetDomain *PacketDomain, packetDomainBytes []byte, packetType uint16, packetClass uint16, packetTtl uint32,
	packetDataLength uint16, packetData interface{}, packetDataBytes []byte, packetModelType int) (packetModel *PacketModel) {
	belogs.Debug("NewPacketModel(): packetDomain:", jsonutil.MarshalJson(packetDomain),
		"   packetDomainBytes:", convert.PrintBytesOneLine(packetDomainBytes), "  packetType:", packetType,
		"   packetClass:", packetClass, "  packetTtl:", packetTtl, "  packetDataLength:", packetDataLength, "   packetData:", packetData,
		"   packetDataBytes:", convert.PrintBytesOneLine(packetDataBytes), "  packetModelType:", packetModelType)
	c := &PacketModel{
		PacketDomain:      packetDomain,
		PacketDomainBytes: packetDomainBytes,
		PacketType:        packetType,
		PacketClass:       packetClass,
		PacketTtl:         packetTtl,
		PacketDataLength:  packetDataLength,
		PacketData:        packetData,
		PacketDataBytes:   packetDataBytes,
		PacketModelType:   packetModelType,
	}
	belogs.Debug("NewPacketModel(): packetModel:", jsonutil.MarshalJson(c))
	return c
}

// packetBytes:=receiveData[offsetFromStart:]
// rrAllLen:>0,will compare; if <=0, no use
// offsetFromStart: from 'id' start, will be used to support compression domain
// packetModelType: question/answer_full, answer_abridge
// packetModelCount: if >0, will limit to count; if <=0, no limit
// packetDecompressionLabel: saved domain's label
func ParseBytesToPacketModels(packetBytes []byte, rrAllLen uint16, offsetFromStart uint16,
	packetModelType int, packetModelCount int, packetDecompressionLabel *PacketDecompressionLabel) (packetModels []*PacketModel, newOffsetFromStart uint16, err error) {
	if len(packetBytes) == 0 {
		return nil, 0, errors.New("packetBytes is empty")
	}
	belogs.Debug("ParseBytesToPacketModels():packetBytes:", convert.PrintBytesOneLine(packetBytes),
		" rrAllLen:", rrAllLen, "   offsetFromStart:", offsetFromStart, "  packetModelType:", packetModelType, "  packetModelCount:", packetModelCount)

	packetModels = make([]*PacketModel, 0)
	rrSumLen := uint16(0)
	var inLoopBytess, newLoopBytes []byte
	newLoopBytes = make([]byte, len(packetBytes))
	copy(newLoopBytes, packetBytes)
	newLoopOffset := offsetFromStart
	for {
		// get dnsName: expectCount is 1
		belogs.Debug("ParseBytesToPacketModels():new for-loop,  newLoopBytes:", convert.PrintBytesOneLine(newLoopBytes),
			"  newLoopOffset:", newLoopOffset)
		packetDomain, inLoopOffset, err := ParseBytesToPacketDomain(newLoopBytes, newLoopOffset, packetDecompressionLabel)
		if err != nil {
			belogs.Error("ParseBytesToPacketModels():ParseBytesToPacketDomain fail, packetBytes:",
				convert.PrintBytesOneLine(packetBytes), err)
			return nil, 0, err
		}
		belogs.Debug("ParseBytesToPacketModels():after ParseBytesToPacketDomain, packetDomain:", jsonutil.MarshalJson(packetDomain),
			"   inLoopOffset:", inLoopOffset, "  newLoopOffset:", newLoopOffset)
		inLoopBytess = newLoopBytes[inLoopOffset-newLoopOffset:]
		belogs.Debug("ParseBytesToPacketModels():inLoopOffset:", inLoopOffset,
			"   inLoopBytess:", convert.PrintBytesOneLine(inLoopBytess),
			"   packetModelType:", packetModelType)
		switch packetModelType {
		case DNS_PACKET_NAME_TYPE_CLASS:
			fallthrough
		case DNS_PACKET_NAME_TYPE_CLASS_RDATA:
			// type(2)+class(2)
			if len(inLoopBytess) < (2 + 2) {
				belogs.Error("ParseBytesToPacketModels():after ParseBytesToPacketDomain len(inLoopBytess) <= 4,fail, inLoopBytess: ",
					convert.PrintBytesOneLine(inLoopBytess), err)
				return nil, 0, errors.New("len(inLoopBytess) <= 4")
			}
		case DNS_PACKET_NAME_TYPE_CLASS_TTL_RDLENGTH:
			fallthrough
		case DNS_PACKET_NAME_TYPE_CLASS_TTL_RDLENGTH_RDATA:
			// type(2)+class(2)+ttl(4)+len(2)
			if len(inLoopBytess) < (2 + 2 + 4 + 2) {
				belogs.Error("ParseBytesToPacketModels():after ParseBytesToPacketDomain len(inLoopBytess) <= 8,fail, inLoopBytess: ",
					convert.PrintBytesOneLine(inLoopBytess), err)
				return nil, 0, errors.New("len(inLoopBytess) <= 8")
			}
		}

		// get type
		packetType := binary.BigEndian.Uint16(inLoopBytess[:2])
		_, ok := dnsutil.DnsIntTypes[packetType]
		if !ok {
			belogs.Error("ParseBytesToPacketModels():DnsIntTypes fail,packetType:", packetType, " inLoopBytess[:2]:", convert.PrintBytesOneLine(inLoopBytess[:2]))
			return nil, 0, errors.New("DnsIntTypes fail")
		}
		inLoopOffset += 2
		inLoopBytess = inLoopBytess[2:]
		belogs.Debug("ParseBytesToPacketModels(): packetType:", packetType, "  inLoopOffset:", inLoopOffset,
			"  inLoopBytess:", convert.PrintBytesOneLine(inLoopBytess))

		// get class
		packetClass := binary.BigEndian.Uint16(inLoopBytess[:2])
		_, ok = dnsutil.DnsIntClasses[packetClass]
		if !ok {
			belogs.Error("ParseBytesToPacketModels():DnsIntClasses fail:packetClass:", packetClass, " inLoopBytess[:2]:", convert.PrintBytesOneLine(inLoopBytess[:2]))
			return nil, 0, errors.New("DnsIntClasses fail")
		}
		inLoopOffset += 2
		inLoopBytess = inLoopBytess[2:]
		belogs.Debug("ParseBytesToPacketModels(): packetClass:", packetClass, "  inLoopOffset:", inLoopOffset,
			"  inLoopBytess:", convert.PrintBytesOneLine(inLoopBytess))

		var packetModel *PacketModel
		var packetDataLength uint16
		var packetTtl uint32
		if packetModelType == DNS_PACKET_NAME_TYPE_CLASS {
			belogs.Debug("ParseBytesToPacketModels(): is DNS_PACKET_NAME_TYPE_CLASS:, packetDomain:", jsonutil.MarshalJson(packetDomain),
				"  packetType:", packetType, "  packetClass:", packetClass,
				"  inLoopOffset:", inLoopOffset,
				"  inLoopBytess:", convert.PrintBytesOneLine(inLoopBytess))
			packetModel = NewPacketModel(packetDomain, nil, packetType, packetClass, 0, 0, nil, nil, packetModelType)
			belogs.Debug("ParseBytesToPacketModels():DNS_PACKET_NAME_TYPE_CLASS,packetModelType:", packetModelType,
				"  packetModel:", jsonutil.MarshalJson(packetModel), "  inLoopOffset:", inLoopOffset,
				"  inLoopBytess:", convert.PrintBytesOneLine(inLoopBytess))
		} else {

			// get ttl/rdlength
			if packetModelType == DNS_PACKET_NAME_TYPE_CLASS_TTL_RDLENGTH ||
				packetModelType == DNS_PACKET_NAME_TYPE_CLASS_TTL_RDLENGTH_RDATA {
				packetTtl = binary.BigEndian.Uint32(inLoopBytess[:4])
				inLoopOffset += 4
				inLoopBytess = inLoopBytess[4:]
				belogs.Debug("ParseBytesToPacketModels(): DNS_PACKET_NAME_TYPE_CLASS_TTL_RDLENGTH | DNS_PACKET_NAME_TYPE_CLASS_TTL_RDLENGTH_RDATA,",
					"  packetTtl:", packetTtl,
					"  inLoopOffset:", inLoopOffset,
					"  inLoopBytess:", convert.PrintBytesOneLine(inLoopBytess))

				packetDataLength = binary.BigEndian.Uint16(inLoopBytess[:2])
				inLoopOffset += 2
				inLoopBytess = inLoopBytess[2:]
				belogs.Debug("ParseBytesToPacketModels(): DNS_PACKET_NAME_TYPE_CLASS_TTL_RDLENGTH | DNS_PACKET_NAME_TYPE_CLASS_TTL_RDLENGTH_RDATA,",
					"  packetDataLength:", packetDataLength,
					"  inLoopOffset:", inLoopOffset,
					"  inLoopBytess:", convert.PrintBytesOneLine(inLoopBytess))
			} else if packetModelType == DNS_PACKET_NAME_TYPE_CLASS_RDATA {
				// calc the length, inLoopBytess is not change
				packetDataLength = uint16(rrAllLen - (offsetFromStart - inLoopOffset))
				belogs.Debug("ParseBytesToPacketModels():DNS_PACKET_NAME_TYPE_CLASS_RDATA, packetDataLength:", packetDataLength,
					"  inLoopOffset:", inLoopOffset,
					"  inLoopBytess:", convert.PrintBytesOneLine(inLoopBytess))
			}
			belogs.Debug("ParseBytesToPacketModels(): packetDataLength:", packetDataLength,
				"  inLoopOffset:", inLoopOffset,
				"  inLoopBytess:", convert.PrintBytesOneLine(inLoopBytess))

			if packetDataLength == 0 {
				packetModel = NewPacketModel(packetDomain, nil, packetType, packetClass, packetTtl,
					packetDataLength, nil, nil, packetModelType)
				belogs.Debug("ParseBytesToPacketModels():packetDataLength==0,packetModelType:", packetModelType,
					"  packetModel:", jsonutil.MarshalJson(packetModel), "  inLoopOffset:", inLoopOffset,
					"  inLoopBytess:", convert.PrintBytesOneLine(inLoopBytess))

			} else {

				if uint16(len(inLoopBytess)) < packetDataLength {
					belogs.Error("ParseBytesToPacketModels(): inLoopBytess < packetDataLength, fail, inLoopBytess:", convert.PrintBytesOneLine(inLoopBytess),
						"  len(inLoopBytess):", len(inLoopBytess), "   packetDataLength:", packetDataLength)
					return nil, 0, errors.New("len(inLoopBytess) < packetDataLength")
				}

				// when remove collective rr , dsnRdLen == 0
				var packetDataLengthTmp uint16
				var packetDataBytes []byte
				var packetData interface{}
				switch packetType {
				case dnsutil.DNS_TYPE_INT_A:
					aModel, inLoopOffsetTmp, err := ParseBytesToAModel(inLoopBytess[:packetDataLength], inLoopOffset)
					if err != nil {
						belogs.Error("ParseBytesToPacketModels(): ParseBytesToAModel fail, inLoopBytess:", convert.PrintBytesOneLine(inLoopBytess))
						return nil, 0, err
					}
					packetDataLengthTmp = aModel.Length()
					packetData = aModel
					packetDataBytes = aModel.Bytes()
					inLoopOffset = inLoopOffsetTmp
				case dnsutil.DNS_TYPE_INT_NS:
					nsModel, inLoopOffsetTmp, err := ParseBytesToNsModel(inLoopBytess[:packetDataLength], inLoopOffset, packetDecompressionLabel)
					if err != nil {
						belogs.Error("ParseBytesToPacketModels(): ParseBytesToNsModel fail, inLoopBytess:", convert.PrintBytesOneLine(inLoopBytess))
						return nil, 0, err
					}
					packetDataLengthTmp = nsModel.Length()
					packetData = nsModel
					packetDataBytes = nsModel.Bytes()
					inLoopOffset = inLoopOffsetTmp
				case dnsutil.DNS_TYPE_INT_CNAME:
					cNameModel, inLoopOffsetTmp, err := ParseBytesToCNameModel(inLoopBytess[:packetDataLength], inLoopOffset, packetDecompressionLabel)
					if err != nil {
						belogs.Error("ParseBytesToPacketModels(): ParseBytesToCNameModel fail, inLoopBytess:", convert.PrintBytesOneLine(inLoopBytess))
						return nil, 0, err
					}
					packetDataLengthTmp = cNameModel.Length()
					packetData = cNameModel
					packetDataBytes = cNameModel.Bytes()
					inLoopOffset = inLoopOffsetTmp
				case dnsutil.DNS_TYPE_INT_SOA:
					soaModel, inLoopOffsetTmp, err := ParseBytesToSoaModel(inLoopBytess[:packetDataLength], inLoopOffset, packetDecompressionLabel)
					if err != nil {
						belogs.Error("ParseBytesToPacketModels(): ParseBytesToSoaModel fail, inLoopBytess:", convert.PrintBytesOneLine(inLoopBytess))
						return nil, 0, err
					}
					packetDataLengthTmp = soaModel.Length()
					packetData = soaModel
					packetDataBytes = soaModel.Bytes()
					inLoopOffset = inLoopOffsetTmp
				case dnsutil.DNS_TYPE_INT_PTR:
					ptrModel, inLoopOffsetTmp, err := ParseBytesToPtrModel(inLoopBytess[:packetDataLength], inLoopOffset, packetDecompressionLabel)
					if err != nil {
						belogs.Error("ParseBytesToPacketModels(): ParseBytesToPtrModel fail, inLoopBytess:", convert.PrintBytesOneLine(inLoopBytess))
						return nil, 0, err
					}
					packetDataLengthTmp = ptrModel.Length()
					packetData = ptrModel
					packetDataBytes = ptrModel.Bytes()
					inLoopOffset = inLoopOffsetTmp
				case dnsutil.DNS_TYPE_INT_MX:
					mxModel, inLoopOffsetTmp, err := ParseBytesToMxModel(inLoopBytess[:packetDataLength], inLoopOffset, packetDecompressionLabel)
					if err != nil {
						belogs.Error("ParseBytesToPacketModels(): ParseBytesToMxModel fail, inLoopBytess:", convert.PrintBytesOneLine(inLoopBytess))
						return nil, 0, err
					}
					packetDataLengthTmp = mxModel.Length()
					packetData = mxModel
					packetDataBytes = mxModel.Bytes()
					inLoopOffset = inLoopOffsetTmp
				case dnsutil.DNS_TYPE_INT_TXT:
					txtModel, inLoopOffsetTmp, err := ParseBytesToTxtModel(inLoopBytess[:packetDataLength], inLoopOffset)
					if err != nil {
						belogs.Error("ParseBytesToPacketModels(): ParseBytesToTxtModel fail, inLoopBytess:", convert.PrintBytesOneLine(inLoopBytess))
						return nil, 0, err
					}
					packetDataLengthTmp = txtModel.Length()
					packetData = txtModel
					packetDataBytes = txtModel.Bytes()
					inLoopOffset = inLoopOffsetTmp
				case dnsutil.DNS_TYPE_INT_AAAA:
					aaaaModel, inLoopOffsetTmp, err := ParseBytesToAaaaModel(inLoopBytess[:packetDataLength], inLoopOffset)
					if err != nil {
						belogs.Error("ParseBytesToPacketModels(): ParseBytesToAaaaModel fail, inLoopBytess:", convert.PrintBytesOneLine(inLoopBytess))
						return nil, 0, err
					}
					packetDataLengthTmp = aaaaModel.Length()
					packetData = aaaaModel
					packetDataBytes = aaaaModel.Bytes()
					inLoopOffset = inLoopOffsetTmp
				case dnsutil.DNS_TYPE_INT_SRV:
					srvModel, inLoopOffsetTmp, err := ParseBytesToSrvModel(inLoopBytess[:packetDataLength], inLoopOffset, packetDecompressionLabel)
					if err != nil {
						belogs.Error("ParseBytesToPacketModels(): ParseBytesToSrvModel fail, inLoopBytess:", convert.PrintBytesOneLine(inLoopBytess))
						return nil, 0, err
					}
					packetDataLengthTmp = srvModel.Length()
					packetData = srvModel
					packetDataBytes = srvModel.Bytes()
					inLoopOffset = inLoopOffsetTmp
				default:
					belogs.Error("ParseBytesToPacketModels(): not support TYPE fail, inLoopBytess:", convert.PrintBytesOneLine(inLoopBytess))
					return nil, 0, errors.New("not support TYPE")
				}
				belogs.Debug("ParseBytesToPacketModels():   packetDataLength:", packetDataLength, "  packetDataLengthTmp:", packetDataLengthTmp,
					"   packetData:", jsonutil.MarshalJson(packetData),
					"   packetDataBytes:", convert.PrintBytesOneLine(packetDataBytes),
					"   inLoopOffset:", inLoopOffset)

				if bytes.Compare(packetDataBytes, inLoopBytess[:packetDataLength]) != 0 {
					belogs.Error("ParseBytesToPacketModels(): packetDataBytes is not equal inLoopBytess, packetDataBytes:",
						convert.PrintBytesOneLine(packetDataBytes), "   inLoopBytess:", convert.PrintBytesOneLine(inLoopBytess[:packetDataLength]))
					return nil, 0, errors.New("compare rr data fail")
				}

				if packetDataLength != packetDataLengthTmp {
					belogs.Error("ParseBytesToPacketModels(): packetDataLength is not euqal to packetDataLengthTmp, packetDataLength:", packetDataLength,
						"   packetDataLengthTmp:", packetDataLengthTmp)
					return nil, 0, errors.New("compare rr data length fail")
				}

				packetModel = NewPacketModel(packetDomain, nil, packetType, packetClass, packetTtl,
					packetDataLength, packetData, packetDataBytes, packetModelType)
				inLoopBytess = inLoopBytess[packetDataLength:]
				belogs.Debug("ParseBytesToPacketModels():packetDataLength>0, packetModelType:", packetModelType,
					"  packetModel:", jsonutil.MarshalJson(packetModel), "  inLoopOffset:", inLoopOffset,
					"  inLoopBytess:", convert.PrintBytesOneLine(inLoopBytess))
			}
		}
		belogs.Debug("ParseBytesToPacketModels():packetModelType:", packetModelType, " packetModel:", jsonutil.MarshalJson(packetModel),
			"  inLoopOffset:", inLoopOffset)
		packetModels = append(packetModels, packetModel)

		belogs.Debug("ParseBytesToPacketModels(): after append packetModels, len(inLoopBytess):", len(inLoopBytess),
			"  inLoopOffset:", inLoopOffset, "  inLoopBytess:", convert.PrintBytesOneLine(inLoopBytess),
			"  packetDataLength:", packetDataLength, "  packetModelType:", packetModelType,
			"  len(packetModels):", len(packetModels))

		if len(inLoopBytess) == 0 {
			belogs.Debug("ParseBytesToPacketModels():break, len(inLoopBytess) == 0:", jsonutil.MarshalJson(packetModel),
				"  inLoopOffset:", inLoopOffset)
			newOffsetFromStart = inLoopOffset
			break
		}

		if packetModelCount > 0 && len(packetModels) >= packetModelCount {
			belogs.Debug("ParseBytesToPacketModels():break, len(packetModels) >= packetModelCount :", jsonutil.MarshalJson(packetModel),
				"  inLoopOffset:", inLoopOffset)
			newOffsetFromStart = inLoopOffset
			break
		}
		rrSumLen += uint16(packetModel.Length())
		belogs.Debug("ParseBytesToPacketModels(): packetModel.Length():", packetModel.Length(), " rrSumLen:", rrSumLen)
		if rrAllLen > 0 {
			if rrSumLen >= rrAllLen {
				newOffsetFromStart = inLoopOffset
				belogs.Debug("ParseBytesToPacketModels():break,  rrSumLen >= rrAllLen:", jsonutil.MarshalJson(packetModel))
				break
			}
		}
		belogs.Debug("ParseBytesToPacketModels():after rrAllLen, len(inLoopBytess):", len(inLoopBytess),
			" inLoopBytess:", convert.PrintBytesOneLine(inLoopBytess), "  inLoopOffset:", inLoopOffset)
		newLoopBytes = make([]byte, len(inLoopBytess))
		copy(newLoopBytes, inLoopBytess)
		newLoopOffset = inLoopOffset
		belogs.Debug("ParseBytesToPacketModels():will new-loop, len(newLoopBytes):", len(newLoopBytes),
			"  newLoopOffset:", newLoopOffset, " newLoopBytes:", convert.PrintBytesOneLine(newLoopBytes))
	}
	belogs.Debug("ParseBytesToPacketModels(): packetModels:", jsonutil.MarshalJson(packetModels), "  newOffsetFromStart:", newOffsetFromStart)
	return packetModels, newOffsetFromStart, nil
}

func (c *PacketModel) Length() uint16 {

	switch c.PacketModelType {
	case DNS_PACKET_NAME_TYPE_CLASS:
		//len(domain) +  type(2)+class(2)
		return uint16(len(c.PacketDomain.Bytes()) + 2 + 2)
	case DNS_PACKET_NAME_TYPE_CLASS_RDATA:
		//len(domain) +  type(2)+class(2)+ PacketDataLength
		return uint16(len(c.PacketDomain.Bytes()) + 2 + 2 + int(c.PacketDataLength))
	case DNS_PACKET_NAME_TYPE_CLASS_TTL_RDLENGTH:
		// len(domain) +  type(2)+class(2)+ttl(4)+rdlen(2)
		return uint16(len(c.PacketDomain.Bytes()) + 2 + 2 + 4 + 2)
	case DNS_PACKET_NAME_TYPE_CLASS_TTL_RDLENGTH_RDATA:
		// len(domain) +  type(2)+class(2)+ttl(4)+rdlen(2)+ packatdatalen
		return uint16(len(c.PacketDomain.Bytes()) + 2 + 2 + 4 + 2 + int(c.PacketDataLength))
	}
	belogs.Error("PacketModel(): Length fail, PacketModelType not supported, c.PacketModelType:", c.PacketModelType)
	return 0

}

func (c *PacketModel) Bytes() []byte {
	wr := bytes.NewBuffer([]byte{})
	binary.Write(wr, binary.BigEndian, c.PacketDomain.Bytes())
	belogs.Debug("PacketModel.Bytes():packetDomain:", string(c.PacketDomain.FullDomain), "   packetModelType:", c.PacketModelType)

	binary.Write(wr, binary.BigEndian, c.PacketType)
	binary.Write(wr, binary.BigEndian, c.PacketClass)
	belogs.Debug("PacketModel.Bytes():type:", c.PacketType, "  class:", c.PacketClass)

	if c.PacketModelType == DNS_PACKET_NAME_TYPE_CLASS {
		b := wr.Bytes()
		belogs.Debug("PacketModel.Bytes(): DNS_PACKET_NAME_TYPE_CLASS, return b:", convert.PrintBytesOneLine(b))
		return b
	}

	if c.PacketModelType == DNS_PACKET_NAME_TYPE_CLASS_TTL_RDLENGTH ||
		c.PacketModelType == DNS_PACKET_NAME_TYPE_CLASS_TTL_RDLENGTH_RDATA {
		binary.Write(wr, binary.BigEndian, c.PacketTtl)
		binary.Write(wr, binary.BigEndian, c.PacketDataLength)
		belogs.Debug("PacketModel.Bytes():ttl,datalength, DNS_PACKET_NAME_TYPE_CLASS_TTL_RDLENGTH or DNS_PACKET_NAME_TYPE_CLASS_TTL_RDLENGTH_RDATA:",
			"   ttl:", c.PacketTtl, "  packetDataLength:", c.PacketDataLength,
			"   b:", convert.PrintBytesOneLine(wr.Bytes()))
	}

	if c.PacketModelType == DNS_PACKET_NAME_TYPE_CLASS_RDATA ||
		c.PacketModelType == DNS_PACKET_NAME_TYPE_CLASS_TTL_RDLENGTH_RDATA {
		belogs.Debug("PacketModel.Bytes():rdata, DNS_PACKET_NAME_TYPE_CLASS_RDATA or DNS_PACKET_NAME_TYPE_CLASS_TTL_RDLENGTH_RDATA:",
			"  PacketData:", c.PacketData, "  PacketType:", c.PacketType)

		if c.PacketData != nil {
			packetDataJson := jsonutil.MarshalJson(c.PacketData)
			belogs.Debug("PacketModel.Bytes(): packetDataJson:", packetDataJson)
			var err error
			switch c.PacketType {
			case dnsutil.DNS_TYPE_INT_A:
				var aModel AModel
				err = jsonutil.UnmarshalJson(packetDataJson, &aModel)
				if err == nil {
					belogs.Debug("PacketModel.Bytes(): write aModel:", jsonutil.MarshalJson(aModel))
					binary.Write(wr, binary.BigEndian, aModel.Bytes())
				}
			case dnsutil.DNS_TYPE_INT_NS:
				var nsModel NsModel
				err = jsonutil.UnmarshalJson(packetDataJson, &nsModel)
				if err == nil {
					belogs.Debug("PacketModel.Bytes(): nsModel:", jsonutil.MarshalJson(nsModel))
					binary.Write(wr, binary.BigEndian, nsModel.Bytes())
				}
			case dnsutil.DNS_TYPE_INT_CNAME:
				var cNameModel CNameModel
				err = jsonutil.UnmarshalJson(packetDataJson, &cNameModel)
				if err == nil {
					belogs.Debug("PacketModel.Bytes(): cNameModel:", jsonutil.MarshalJson(cNameModel))
					binary.Write(wr, binary.BigEndian, cNameModel.Bytes())
				}
			case dnsutil.DNS_TYPE_INT_SOA:
				var soaModel SoaModel
				err = jsonutil.UnmarshalJson(packetDataJson, &soaModel)
				if err == nil {
					belogs.Debug("PacketModel.Bytes(): soaModel:", jsonutil.MarshalJson(soaModel))
					binary.Write(wr, binary.BigEndian, soaModel.Bytes())
				}
			case dnsutil.DNS_TYPE_INT_PTR:
				var ptrModel PtrModel
				err = jsonutil.UnmarshalJson(packetDataJson, &ptrModel)
				if err == nil {
					belogs.Debug("PacketModel.Bytes(): ptrModel:", jsonutil.MarshalJson(ptrModel))
					binary.Write(wr, binary.BigEndian, ptrModel.Bytes())
				}
			case dnsutil.DNS_TYPE_INT_MX:
				var mxModel MxModel
				err = jsonutil.UnmarshalJson(packetDataJson, &mxModel)
				if err == nil {
					belogs.Debug("PacketModel.Bytes(): mxModel:", jsonutil.MarshalJson(mxModel))
					binary.Write(wr, binary.BigEndian, mxModel.Bytes())
				}
			case dnsutil.DNS_TYPE_INT_TXT:
				var txtModel TxtModel
				err = jsonutil.UnmarshalJson(packetDataJson, &txtModel)
				if err == nil {
					belogs.Debug("PacketModel.Bytes(): txtModel:", jsonutil.MarshalJson(txtModel))
					binary.Write(wr, binary.BigEndian, txtModel.Bytes())
				}
			case dnsutil.DNS_TYPE_INT_AAAA:
				var aaaaModel AaaaModel
				err = jsonutil.UnmarshalJson(packetDataJson, &aaaaModel)
				if err == nil {
					belogs.Debug("PacketModel.Bytes(): aaaaModel:", jsonutil.MarshalJson(aaaaModel))
					binary.Write(wr, binary.BigEndian, aaaaModel.Bytes())
				}
			case dnsutil.DNS_TYPE_INT_SRV:
				var srvModel SrvModel
				err = jsonutil.UnmarshalJson(packetDataJson, &srvModel)
				if err == nil {
					belogs.Debug("PacketModel.Bytes(): srvModel:", jsonutil.MarshalJson(srvModel))
					binary.Write(wr, binary.BigEndian, srvModel.Bytes())
				}
			}
			if err != nil {
				belogs.Error("PacketModel.Bytes(): get **Model fail,PacketType:", c.PacketType,
					" packetDataJson:", packetDataJson, err)
			}
			belogs.Debug("PacketModel.Bytes():rdata, DNS_PACKET_NAME_TYPE_CLASS_RDATA or DNS_PACKET_NAME_TYPE_CLASS_TTL_RDLENGTH_RDATA:",
				"  PacketType:", c.PacketType, "   b:", convert.PrintBytesOneLine(wr.Bytes()))
		} else if len(c.PacketDataBytes) > 0 {
			binary.Write(wr, binary.BigEndian, c.PacketDataBytes)
		}

	}
	b := wr.Bytes()
	belogs.Debug("PacketModel.Bytes(): b:", convert.PrintBytesOneLine(wr.Bytes()))
	return b

}
