package convert

import (
	"errors"
	"strings"

	"dns-model/packet"
	pushmodel "dns-model/push"
	"dns-model/rr"
	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/convert"
	"github.com/cpusoft/goutil/dnsutil"
	"github.com/cpusoft/goutil/jsonutil"
	"github.com/guregu/null"
)

// zName: may be ""
func ConvertPacketToRr(zName string, packetModel *packet.PacketModel) (rrModel *rr.RrModel, err error) {
	belogs.Debug("ConvertPacketToRr():zName:", zName, " pakcetModel:", jsonutil.MarshalJson(packetModel))
	fullDomain := packetModel.PacketDomain.FullDomain
	packetType := packetModel.PacketType
	packetClass := packetModel.PacketClass
	packetTtl := packetModel.PacketTtl
	packetDataJson := jsonutil.MarshalJson(packetModel.PacketData)
	belogs.Debug("ConvertPacketToRr():fullDomain:", string(fullDomain), " packetType:", packetType,
		"  packetClass:", packetClass, "  packetTtl:", packetTtl,
		"  packetModel.PacketDataLength:", packetModel.PacketDataLength,
		"  packetModel.PacketData:", packetDataJson,
		"  packetModel.PacketDataBytes:", convert.PrintBytesOneLine(packetModel.PacketDataBytes))

	var rrData string
	switch packetType {
	case dnsutil.DNS_TYPE_INT_A:
		var aModel packet.AModel
		err = jsonutil.UnmarshalJson(packetDataJson, &aModel)
		if err != nil {
			belogs.Error("ConvertPacketToRr():aModel fail, packetDataJson:", packetDataJson, err)
			return nil, err
		}
		rrData = aModel.ToRrData()
		belogs.Debug("ConvertPacketToRr():aModel:", jsonutil.MarshalJson(aModel), "  rrData:", rrData)
	case dnsutil.DNS_TYPE_INT_NS:
		var nsModel packet.NsModel
		err = jsonutil.UnmarshalJson(packetDataJson, &nsModel)
		if err != nil {
			belogs.Error("ConvertPacketToRr():nsModel fail, packetDataJson:", packetDataJson, err)
			return nil, err
		}
		rrData = nsModel.ToRrData()
		belogs.Debug("ConvertPacketToRr():nsModel:", jsonutil.MarshalJson(nsModel), "  rrData:", rrData)
	case dnsutil.DNS_TYPE_INT_CNAME:
		var cNameModel packet.CNameModel
		err = jsonutil.UnmarshalJson(packetDataJson, &cNameModel)
		if err != nil {
			belogs.Error("ConvertPacketToRr():cNameModel fail, packetDataJson:", packetDataJson, err)
			return nil, err
		}
		rrData = cNameModel.ToRrData()
		belogs.Debug("ConvertPacketToRr():cNameModel:", jsonutil.MarshalJson(cNameModel), "  rrData:", rrData)
	case dnsutil.DNS_TYPE_INT_SOA:
		var soaModel packet.SoaModel
		err = jsonutil.UnmarshalJson(packetDataJson, &soaModel)
		if err != nil {
			belogs.Error("ConvertPacketToRr():soaModel fail, packetDataJson:", packetDataJson, err)
			return nil, err
		}
		rrData = soaModel.ToRrData()
		belogs.Debug("ConvertPacketToRr():soaModel:", jsonutil.MarshalJson(soaModel), "  rrData:", rrData)
	case dnsutil.DNS_TYPE_INT_PTR:
		var ptrModel packet.PtrModel
		err = jsonutil.UnmarshalJson(packetDataJson, &ptrModel)
		if err != nil {
			belogs.Error("ConvertPacketToRr():ptrModel fail, packetDataJson:", packetDataJson, err)
			return nil, err
		}
		rrData = ptrModel.ToRrData()
		belogs.Debug("ConvertPacketToRr():ptrModel:", jsonutil.MarshalJson(ptrModel), "  rrData:", rrData)

	case dnsutil.DNS_TYPE_INT_MX:
		var mxModel packet.MxModel
		err = jsonutil.UnmarshalJson(packetDataJson, &mxModel)
		if err != nil {
			belogs.Error("ConvertPacketToRr():mxModel fail, packetDataJson:", packetDataJson, err)
			return nil, err
		}
		rrData = mxModel.ToRrData()
		belogs.Debug("ConvertPacketToRr():mxModel:", jsonutil.MarshalJson(mxModel), "  rrData:", rrData)

	case dnsutil.DNS_TYPE_INT_TXT:
		var txtModel packet.TxtModel
		err = jsonutil.UnmarshalJson(packetDataJson, &txtModel)
		if err != nil {
			belogs.Error("ConvertPacketToRr():txtModel fail, packetDataJson:", packetDataJson, err)
			return nil, err
		}
		rrData = txtModel.ToRrData()
		belogs.Debug("ConvertPacketToRr():txtModel:", jsonutil.MarshalJson(txtModel), "  rrData:", rrData)

	case dnsutil.DNS_TYPE_INT_AAAA:
		var aaaaModel packet.AaaaModel
		err = jsonutil.UnmarshalJson(packetDataJson, &aaaaModel)
		if err != nil {
			belogs.Error("ConvertPacketToRr():aaaaModel fail, packetDataJson:", packetDataJson, err)
			return nil, err
		}
		rrData = aaaaModel.ToRrData()
		belogs.Debug("ConvertPacketToRr():aaaaModel:", jsonutil.MarshalJson(aaaaModel), "  rrData:", rrData)

	case dnsutil.DNS_TYPE_INT_SRV:
		var srvModel packet.SrvModel
		err = jsonutil.UnmarshalJson(packetDataJson, &srvModel)
		if err != nil {
			belogs.Error("ConvertPacketToRr():srvModel fail, packetDataJson:", packetDataJson, err)
			return nil, err
		}
		rrData = srvModel.ToRrData()
		belogs.Debug("ConvertPacketToRr():srvModel:", jsonutil.MarshalJson(srvModel), "  rrData:", rrData)
	case dnsutil.DNS_TYPE_INT_ANY:
		rrData = packetDataJson
		belogs.Debug("ConvertPacketToRr():type is ANY, rrData is packetDataJson:", rrData)
	default:
		belogs.Error("ConvertPacketToRr(): not support TYPE fail, packetType:", packetType)
		return nil, errors.New("not support TYPE")
	}

	if len(zName) > 0 {
		rrName := strings.TrimSuffix(string(fullDomain), zName)
		belogs.Debug("ConvertPacketToRr(): len(zName)>0, fullDomain:", string(fullDomain),
			"  zName:", zName, " rrName:", rrName)
		rrModel = rr.NewRrModel(rr.FormatRrOrigin(zName), rr.FormatRrName(rrName),
			dnsutil.DnsIntTypes[packetType], dnsutil.DnsIntClasses[packetClass],
			null.IntFrom(int64(packetTtl)), rrData)
	} else {
		rrModel = rr.NewRrModelByFullDomain(string(fullDomain), dnsutil.DnsIntTypes[packetType], dnsutil.DnsIntClasses[packetClass],
			null.IntFrom(int64(packetTtl)), rrData)
	}

	belogs.Debug("ConvertPacketToRr(): rrModel:", jsonutil.MarshalJson(rrModel))
	return rrModel, nil
}

// only packetModel. header is different in query/updata/dso
func ConvertRrToPacket(originTtl null.Int, rrModel *rr.RrModel, offsetFromStart uint16,
	packetModelType int) (packetModel *packet.PacketModel, newOffsetFromStart uint16, err error) {
	belogs.Debug("ConvertRrModelToPacketModel(): originTtl:", originTtl,
		"  rrModel:", jsonutil.MarshalJson(rrModel), "   packetModelType:", packetModelType)

	fullDomain := rrModel.RrFullDomain
	if len(fullDomain) == 0 {
		if len(rrModel.RrName) > 0 {
			fullDomain = rrModel.RrName + "." + strings.TrimSuffix(rrModel.Origin, ".")
		} else {
			fullDomain = strings.TrimSuffix(rrModel.Origin, ".")
		}
	}
	belogs.Debug("ConvertRrModelToPacketModel(): fullDomain:", fullDomain)

	packetDomain, newOffsetFromStart, err := packet.NewPacketDomainNoCompression([]byte(fullDomain), offsetFromStart)
	if err != nil {
		belogs.Error("ConvertRrToPacket(): NewPacketDomainNoCompression fail, RrName:", rrModel.RrName, "  originTtl:", originTtl, err)
		return nil, 0, err
	}
	belogs.Debug("ConvertRrToPacket(): packetDomain:", jsonutil.MarshalJson(packetDomain), "   newOffsetFromStart:", newOffsetFromStart)

	packetType, ok := dnsutil.DnsStrTypes[rrModel.RrType]
	if !ok {
		belogs.Error("ConvertRrToPacket(): DnsStrTypes fail, RrType:", rrModel.RrType)
		return nil, 0, errors.New("RrType is illegal")
	}
	newOffsetFromStart += 2
	belogs.Debug("ConvertRrToPacket(): packetType:", packetType, "   newOffsetFromStart:", newOffsetFromStart)

	packetClass, ok := dnsutil.DnsStrClasses[rrModel.RrClass]
	if !ok {
		belogs.Error("ConvertRrToPacket(): DnsStrClasses fail, RrClass:", rrModel.RrClass)
		return nil, 0, errors.New("RrClass is illegal")
	}
	newOffsetFromStart += 2
	belogs.Debug("ConvertRrToPacket(): packetClass:", packetClass, "   newOffsetFromStart:", newOffsetFromStart)

	var packetTtl uint32
	if packetModelType == packet.DNS_PACKET_NAME_TYPE_CLASS_TTL_RDLENGTH ||
		packetModelType == packet.DNS_PACKET_NAME_TYPE_CLASS_TTL_RDLENGTH_RDATA {
		if !rrModel.RrTtl.IsZero() {
			packetTtl = uint32(rrModel.RrTtl.ValueOrZero())
		} else {
			if !originTtl.IsZero() {
				packetTtl = uint32(originTtl.ValueOrZero())
			}
		}
		newOffsetFromStart += 4
		belogs.Debug("ConvertRrToPacket():packetModelType have _TTL_, packetTtl:", packetTtl, "   newOffsetFromStart:", newOffsetFromStart)
	}

	var packetDataLength uint16
	var packetData interface{}
	var packetDataBytes = make([]byte, 0)
	if len(rrModel.RrData) > 0 &&
		(packetModelType == packet.DNS_PACKET_NAME_TYPE_CLASS_RDATA || packetModelType == packet.DNS_PACKET_NAME_TYPE_CLASS_TTL_RDLENGTH_RDATA) {
		belogs.Debug("ConvertRrToPacket(): packetType:", packetType)
		switch packetType {
		case dnsutil.DNS_TYPE_INT_A:
			aModel, newOffsetFromStartTmp, err := packet.NewAModel(rrModel.RrData, newOffsetFromStart)
			if err != nil {
				belogs.Error("ConvertRrToPacket(): NewAModel fail, RrData:", rrModel.RrData)
				return nil, 0, err
			}
			belogs.Debug("ConvertRrToPacket(): aModel:", jsonutil.MarshalJson(aModel))
			packetDataLength = aModel.Length()
			packetData = aModel
			packetDataBytes = aModel.Bytes()
			newOffsetFromStart = newOffsetFromStartTmp
		case dnsutil.DNS_TYPE_INT_NS:
			nsModel, newOffsetFromStartTmp, err := packet.NewNsModel(rrModel.RrData, newOffsetFromStart)
			if err != nil {
				belogs.Error("ConvertRrToPacket(): NewNsModel fail, RrData:", rrModel.RrData)
				return nil, 0, err
			}
			belogs.Debug("ConvertRrToPacket(): nsModel:", jsonutil.MarshalJson(nsModel))
			packetDataLength = nsModel.Length()
			packetData = nsModel
			packetDataBytes = nsModel.Bytes()
			newOffsetFromStart = newOffsetFromStartTmp
		case dnsutil.DNS_TYPE_INT_CNAME:
			cNameModel, newOffsetFromStartTmp, err := packet.NewCNameModel(rrModel.RrData, newOffsetFromStart)
			if err != nil {
				belogs.Error("ConvertRrToPacket(): NewCNameModel fail, RrData:", rrModel.RrData)
				return nil, 0, err
			}
			belogs.Debug("ConvertRrToPacket(): cNameModel:", jsonutil.MarshalJson(cNameModel))
			packetDataLength = cNameModel.Length()
			packetData = cNameModel
			packetDataBytes = cNameModel.Bytes()
			newOffsetFromStart = newOffsetFromStartTmp
		case dnsutil.DNS_TYPE_INT_SOA:
			split := strings.Split(rrModel.RrData, " ")
			if len(split) != 7 {
				belogs.Error("ConvertRrToPacket(): SOA split fail, RrData:", rrModel.RrData)
				return nil, 0, errors.New("SOA illegal format")
			}
			soaModel, newOffsetFromStartTmp, err := packet.NewSoaModel(split[0], split[1], split[2], split[3],
				split[4], split[5], split[6], newOffsetFromStart)
			if err != nil {
				belogs.Error("ConvertRrToPacket(): NewSoaModel fail, RrData:", rrModel.RrData)
				return nil, 0, err
			}
			belogs.Debug("ConvertRrToPacket(): soaModel:", jsonutil.MarshalJson(soaModel))
			packetDataLength = soaModel.Length()
			packetData = soaModel
			packetDataBytes = soaModel.Bytes()
			newOffsetFromStart = newOffsetFromStartTmp
		case dnsutil.DNS_TYPE_INT_PTR:
			ptrModel, newOffsetFromStartTmp, err := packet.NewPtrModel(rrModel.RrData, newOffsetFromStart)
			if err != nil {
				belogs.Error("ConvertRrToPacket(): NewPtrModel fail, RrData:", rrModel.RrData)
				return nil, 0, err
			}
			belogs.Debug("ConvertRrToPacket(): ptrModel:", jsonutil.MarshalJson(ptrModel))
			packetDataLength = ptrModel.Length()
			packetData = ptrModel
			packetDataBytes = ptrModel.Bytes()
			newOffsetFromStart = newOffsetFromStartTmp
		case dnsutil.DNS_TYPE_INT_MX:
			split := strings.Split(rrModel.RrData, " ")
			if len(split) != 2 {
				belogs.Error("ConvertRrToPacket(): MX split fail, RrData:", rrModel.RrData)
				return nil, 0, errors.New("MX illegal format")
			}
			mxModel, newOffsetFromStartTmp, err := packet.NewMxModel(split[0], split[1], newOffsetFromStart)
			if err != nil {
				belogs.Error("ConvertRrToPacket(): NewMxModel fail, RrData:", rrModel.RrData)
				return nil, 0, err
			}
			belogs.Debug("ConvertRrToPacket(): mxModel:", jsonutil.MarshalJson(mxModel))
			packetDataLength = mxModel.Length()
			packetData = mxModel
			packetDataBytes = mxModel.Bytes()
			newOffsetFromStart = newOffsetFromStartTmp
		case dnsutil.DNS_TYPE_INT_TXT:
			txtModel, newOffsetFromStartTmp, err := packet.NewTxtModel(rrModel.RrData, newOffsetFromStart)
			if err != nil {
				belogs.Error("ConvertRrToPacket(): NewTxtModel fail, RrData:", rrModel.RrData)
				return nil, 0, err
			}
			belogs.Debug("ConvertRrToPacket(): txtModel:", jsonutil.MarshalJson(txtModel))
			packetDataLength = txtModel.Length()
			packetData = txtModel
			packetDataBytes = txtModel.Bytes()
			newOffsetFromStart = newOffsetFromStartTmp
		case dnsutil.DNS_TYPE_INT_AAAA:
			aaaaModel, newOffsetFromStartTmp, err := packet.NewAaaaModel(rrModel.RrData, newOffsetFromStart)
			if err != nil {
				belogs.Error("ConvertRrToPacket(): NewAaaaModel fail, RrData:", rrModel.RrData)
				return nil, 0, err
			}
			belogs.Debug("ConvertRrToPacket(): aaaaModel:", jsonutil.MarshalJson(aaaaModel))
			packetDataLength = aaaaModel.Length()
			packetData = aaaaModel
			packetDataBytes = aaaaModel.Bytes()
			newOffsetFromStart = newOffsetFromStartTmp
		case dnsutil.DNS_TYPE_INT_SRV:
			split := strings.Split(rrModel.RrData, " ")
			if len(split) != 4 {
				belogs.Error("ConvertRrToPacket(): SRV split fail, RrData:", rrModel.RrData)
				return nil, 0, errors.New("SRV illegal format")
			}
			srvModel, newOffsetFromStartTmp, err := packet.NewSrvModel(split[0], split[1], split[2], split[3], newOffsetFromStart)
			if err != nil {
				belogs.Error("ConvertRrToPacket(): NewMxModel fail, RrData:", rrModel.RrData)
				return nil, 0, err
			}
			belogs.Debug("ConvertRrToPacket(): srvModel:", jsonutil.MarshalJson(srvModel))
			packetDataLength = srvModel.Length()
			packetData = srvModel
			packetDataBytes = srvModel.Bytes()
			newOffsetFromStart = newOffsetFromStartTmp
		case dnsutil.DNS_TYPE_INT_ANY:
			// just for update (delete any type)
			packetDataLength = 0
			//packetData=interface{}
			//packetDataBytes=make([]byte, 0)
			belogs.Debug("ConvertRrToPacket(): TYPE is ANY, packetDataLength is 0")
		default:
			belogs.Error("ConvertRrToPacket(): not support TYPE fail, RrData:", rrModel.RrData)
			return nil, 0, errors.New("not support TYPE")
		}
		belogs.Debug("ConvertRrToPacket(): packetModelType have _RDATA, packetDataLength:", packetDataLength,
			"   packetData:", jsonutil.MarshalJson(packetData),
			"   packetDataBytes:", convert.PrintBytesOneLine(packetDataBytes),
			"   newOffsetFromStart:", newOffsetFromStart)
	}

	c := packet.NewPacketModel(packetDomain, nil, packetType, packetClass, packetTtl,
		packetDataLength, packetData, packetDataBytes, packetModelType)
	belogs.Debug("ConvertRrToPacket(): packetModel:", jsonutil.MarshalJson(c), "   bytes:", c.Bytes(),
		"  newOffsetFromStart:", newOffsetFromStart)
	return c, newOffsetFromStart, nil
}

// call push
/*update del:
	if rrClass==DNS_CLASS_INT_***, will add
	else if rrClass==DNS_CLASS_INT_ANY
	  	 if rrType==DNS_TYPE_INT_ANY, will delete by fullDomain
		 else will delete by fullDomain and rrType
	else if rrClass==DNS_CLASS_INT_NONE, will delete by fullDomain and rrType and RrData

	to push del
    if TTL==0xFFFFFFFF and Class!=ANY and Type!=ANY , then del by FullDomain and Class and Type and RData
	if TTL==0xFFFFFFFE
		if Class!=ANY and Type!=ANY, then del by FullDomain and Type and Class
	 	if Class!=ANY and Type==ANY, then del by FullDomain and Class
	 	if Class==ANY, delete by fullDomain
*/
func ConvertUpdateRrModelsToPushModels(connKey string, subscribeMessageId uint16, updateRrModels []*rr.RrModel) (pushRrModels []*pushmodel.PushRrModel, err error) {
	pushRrModels = make([]*pushmodel.PushRrModel, 0)
	for i := range updateRrModels {
		pushRrModel, err := convertUpdateRrModelToPushRrModel(connKey, subscribeMessageId, updateRrModels[i])
		if err != nil {
			belogs.Error("ConvertUpdateRrModelsToPushModels(): convertUpdateRrModelToPushRrModel fail, updateRrModels[i]:", jsonutil.MarshalJson(updateRrModels[i]))
			return nil, errors.New("convert update RR fail")
		}
		pushRrModels = append(pushRrModels, pushRrModel)
	}
	return pushRrModels, nil
}

func convertUpdateRrModelToPushRrModel(connKey string, subscribeMessageId uint16, updateRrModel *rr.RrModel) (pushRrModel *pushmodel.PushRrModel, err error) {
	if updateRrModel == nil {
		belogs.Error("convertUpdateRrModelToPushRrModel(): updateRrModel is nil")
		return nil, errors.New("updateRrModel is nil")
	}
	belogs.Debug("convertUpdateRrModelToPushRrModel():connKey:", connKey,
		"  subscribeMessageId:", subscribeMessageId, "  updateRrModel:", jsonutil.MarshalJson(updateRrModel))
	var rrTtl null.Int
	var rrClass string
	if updateRrModel.RrClass != dnsutil.DNS_CLASS_STR_ANY &&
		updateRrModel.RrClass != dnsutil.DNS_CLASS_STR_NONE &&
		updateRrModel.RrType != dnsutil.DNS_TYPE_STR_ANY {
		rrTtl = updateRrModel.RrTtl
		rrClass = updateRrModel.RrClass
		belogs.Debug("convertUpdateRrModelToPushRrModel(): is not CLASS_ANY/CLASS_NONE/TYPE_ANY:",
			"  rrTtl:", rrTtl, "  rrClass:", rrClass)
	} else if updateRrModel.RrClass == dnsutil.DNS_CLASS_STR_ANY {
		if updateRrModel.RrType == dnsutil.DNS_TYPE_STR_ANY {
			//will delete by fullDomain  0xFFFFFFFE
			rrTtl = null.IntFrom(dnsutil.DSO_DEL_COLLECTIVE_RESOURCE_RECORD_TTL)
			rrClass = dnsutil.DNS_CLASS_STR_ANY
			belogs.Debug("convertUpdateRrModelToPushRrModel(): DNS_CLASS_STR_ANY and DNS_TYPE_STR_ANY:",
				"  rrTtl:", rrTtl, "  rrClass:", rrClass)
		} else {
			//will delete by fullDomain and rrType (and rrClass) 0xFFFFFFFE
			rrTtl = null.IntFrom(dnsutil.DSO_DEL_COLLECTIVE_RESOURCE_RECORD_TTL)
			rrClass = dnsutil.DNS_CLASS_STR_IN
			belogs.Debug("convertUpdateRrModelToPushRrModel(): DNS_CLASS_STR_ANY and NOT DNS_TYPE_STR_ANY:",
				"  rrTtl:", rrTtl, "  rrClass:", rrClass)
		}
	} else if updateRrModel.RrClass == dnsutil.DNS_CLASS_STR_NONE {
		//0xFFFFFFFF
		rrTtl = null.IntFrom(dnsutil.DSO_DEL_SPECIFIED_RESOURCE_RECORD_TTL)
		rrClass = dnsutil.DNS_CLASS_STR_IN
		belogs.Debug("convertUpdateRrModelToPushRrModel(): DNS_CLASS_STR_NONE:",
			"  rrTtl:", rrTtl, "  rrClass:", rrClass)
	}
	pushRrModel = pushmodel.NewPushRrModel(
		connKey, updateRrModel.RrFullDomain, updateRrModel.RrType, rrClass, rrTtl, updateRrModel.RrData, subscribeMessageId)
	belogs.Debug("convertUpdateRrModelToPushRrModel():pushRrModel:", jsonutil.MarshalJson(pushRrModel))
	return pushRrModel, nil
}
