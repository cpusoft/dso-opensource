package transact

import (
	"strings"
	"time"

	dnsconvert "dns-model/convert"
	"dns-model/message"
	pushmodel "dns-model/push"
	"dns-model/rr"
	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/conf"
	"github.com/cpusoft/goutil/convert"
	"github.com/cpusoft/goutil/dnsutil"
	"github.com/cpusoft/goutil/httpclient"
	"github.com/cpusoft/goutil/jsonutil"
	"github.com/cpusoft/goutil/transportutil"
	updatemodel "update-core/model"
)

// err: dnserror
func performUpdateSectionTransact(receiveUpdateModel *updatemodel.UpdateModel,
	dnsToProcessMsg *message.DnsToProcessMsg) (err error) {
	belogs.Debug("performUpdateSectionTransact():  receiveUpdateModel:", jsonutil.MarshalJson(receiveUpdateModel),
		"   dnsToProcessMsg:", jsonutil.MarshalJson(dnsToProcessMsg))
	if dnsToProcessMsg.DnsTransactSide == message.DNS_TRANSACT_SIDE_CLIENT {
		belogs.Debug("performUpdateSectionTransact(): is DNS_TRANSACT_TYPE_CLIENT, no check zone")
		return nil
	}
	id := receiveUpdateModel.GetHeaderModel().GetIdOrMessageId()
	belogs.Debug("performUpdateSectionTransact():id:", id)

	receiveUpdateModels := receiveUpdateModel.UpdateDataModel.UpdateModels
	if len(receiveUpdateModels) == 0 {
		belogs.Error("performUpdateSectionTransact():len(receiveUpdateModels)==0, fail:", jsonutil.MarshalJson(receiveUpdateModels))
		return dnsutil.NewDnsError("receiveUpdateModels is empty",
			id, dnsutil.DNS_OPCODE_UPDATE, dnsutil.DNS_RCODE_FORMERR, transportutil.NEXT_CONNECT_POLICY_KEEP)
	}
	belogs.Debug("performUpdateSectionTransact(): len(receiveUpdateModels):", len(receiveUpdateModels))

	// rfc2136  3.4.1.3 - Pseudocode For Update Section Prescan
	err = performUpdateSectionTransactPrescan(receiveUpdateModel)
	if err != nil {
		belogs.Error("performUpdateSectionTransact():performUpdateSectionTransactPrescan, fail:", jsonutil.MarshalJson(receiveUpdateModels), err)
		return err
	}

	// will to db
	receiveUpdateRrModel, err := updatemodel.ConvertUpdateModelToUpdateRrModel(receiveUpdateModel)
	if err != nil {
		belogs.Error("performUpdateSectionTransact():performUpdateSectionTransactPrescan, fail:", jsonutil.MarshalJson(receiveUpdateModels), err)
		return err
	}
	belogs.Debug("performUpdateSectionTransact():ConvertUpdateModelToUpdateRrModel， receiveUpdateRrModel:", jsonutil.MarshalJson(receiveUpdateRrModel)) // update db

	// save to db
	err = performUpdateSectionTransactDb(receiveUpdateRrModel)
	if err != nil {
		belogs.Error("performUpdateSectionTransact():performUpdateSectionTransactDb, fail:", jsonutil.MarshalJson(receiveUpdateModels), err)
		return err
	}
	belogs.Debug("performUpdateSectionTransact():update ok, receiveUpdateModel:", jsonutil.MarshalJson(receiveUpdateModel),
		"  receiveUpdateRrModel:", jsonutil.MarshalJson(receiveUpdateRrModel), "  dnsToProcessMsg:", jsonutil.MarshalJson(dnsToProcessMsg))

	// call push
	updateRrModels := receiveUpdateRrModel.UpdateDataRrModel.UpdateRrModels
	if dnsToProcessMsg.DnsTransactSide == message.DNS_TRANSACT_SIDE_SERVER && len(updateRrModels) > 0 {
		go func(updateRrModels []*rr.RrModel) {
			start := time.Now()
			// convert to push
			pushRrModels, err := dnsconvert.ConvertUpdateRrModelsToPushModels("", id, updateRrModels)
			if err != nil {
				belogs.Error("performUpdateSectionTransact(): ConvertUpdateRrModelsToPushModels, fail, id:", id, " updateRrModels:", jsonutil.MarshalJson(updateRrModels), err)
				return
			}
			belogs.Debug("performUpdateSectionTransact(): ConvertUpdateRrModelsToPushModels, id:", id, " pushRrModels:", jsonutil.MarshalJson(pushRrModels),
				"  updateRrModels:", jsonutil.MarshalJson(updateRrModels))

			// trigger push, get results
			//
			path := "https://" + conf.String("dns-server::serverHost") + ":" + conf.String("dns-server::serverHttpsPort") +
				"/push/queryrrmodelsshouldpush"
			pushResultRrModels := make([]*pushmodel.PushResultRrModel, 0)
			belogs.Debug("performUpdateSectionTransact(): httpclient/push/queryrrmodelsshouldpush,path:", path, "  pushRrModels:", jsonutil.MarshalJson(pushRrModels))
			err = httpclient.PostAndUnmarshalResponseModel(path, jsonutil.MarshalJson(pushRrModels), false, &pushResultRrModels)
			if err != nil {
				belogs.Error("performUpdateSectionTransact(): httpclient/push/queryrrmodelsshouldpush, fail:", path,
					"  pushRrModels:", jsonutil.MarshalJson(pushRrModels), err)
				return
			}
			if len(pushResultRrModels) == 0 {
				belogs.Debug("performUpdateSectionTransact():httpclient/push/queryrrmodelsshouldpush have no results, path:", path,
					"  pushRrModels:", jsonutil.MarshalJson(pushRrModels), " time(s):", time.Since(start))
				return
			}
			belogs.Info("#执行SRP的'数据更新'时,发现已有DSO客户端订阅了相关域名,触发DSO的'数据推送(PUSH)'到对应的客户端, 推送的域名数据:")
			for i := range pushResultRrModels {
				for j := range pushResultRrModels[i].RrModels {
					belogs.Info("#{'域名':'" + pushResultRrModels[i].RrModels[j].RrFullDomain +
						"','Type':'" + pushResultRrModels[i].RrModels[j].RrType +
						"','Class':'" + pushResultRrModels[i].RrModels[j].RrClass +
						"','Ttl':" + convert.ToString(pushResultRrModels[i].RrModels[j].RrTtl.ValueOrZero()) +
						",'Data':" + pushResultRrModels[i].RrModels[j].RrData + "'}")
				}
			}

			pushResultRrModelsJson := jsonutil.MarshalJson(pushResultRrModels)
			belogs.Debug("performUpdateSectionTransact(): httpclient/push/queryrrmodelsshouldpush, pushResultRrModels:", pushResultRrModelsJson)
			dnsMsg := message.NewDnsMsg(message.DNS_MSG_TYPE_DSO_SERVER_PUSH_TO_CLIENT, pushResultRrModelsJson)
			belogs.Debug("performUpdateSectionTransact(): httpclient/push/queryrrmodelsshouldpush,will send dnsMsg:", dnsMsg)
			dnsToProcessMsg.SendDnsToProcessMsgChan(dnsMsg)

		}(updateRrModels)
	}
	return nil
}

func performUpdateSectionTransactPrescan(receiveUpdateModel *updatemodel.UpdateModel) error {
	belogs.Debug("performUpdateSectionTransactPrescan():   receiveUpdateModel:", jsonutil.MarshalJson(receiveUpdateModel))
	id := receiveUpdateModel.GetHeaderModel().GetIdOrMessageId()
	receiveZoneModel := receiveUpdateModel.UpdateDataModel.ZoneModel
	receiveUpdateModels := receiveUpdateModel.UpdateDataModel.UpdateModels
	zName := strings.ToLower(string(receiveZoneModel.ZNamePacketDomain.FullDomain))
	zType := receiveZoneModel.ZType
	zClass := receiveZoneModel.ZClass
	belogs.Debug("performUpdateSectionTransactPrescan():id:", id, " zName:", zName, "  zType:", zType, "  zClass:", zClass)

	for i := range receiveUpdateModels {

		//  if (zone_of(rr.name) != ZNAME)  return (NOTZONE);
		if len(receiveUpdateModels[i].PacketDomain.FullDomain) == 0 {
			belogs.Error("performUpdateSectionTransactPrescan(): len(receiveUpdateModels[i].PacketDomain.FullDomain) is empty, fail, receiveUpdateModels[i]:",
				jsonutil.MarshalJson(receiveUpdateModels[i]))
			return dnsutil.NewDnsError("Name of one receiveUpdateModels is empty",
				id, dnsutil.DNS_OPCODE_UPDATE, dnsutil.DNS_RCODE_FORMERR, transportutil.NEXT_CONNECT_POLICY_KEEP)
		}
		fullName := strings.ToLower(string(receiveUpdateModels[i].PacketDomain.FullDomain))
		belogs.Debug("performUpdateSectionTransactPrescan(): fullName:", fullName)
		if !strings.Contains(fullName, zName) {
			belogs.Error("performUpdateSectionTransactPrescan(): fullName is not in zName, fail, fullName:", fullName, "  zName:", zName,
				jsonutil.MarshalJson(receiveUpdateModels[i]))
			return dnsutil.NewDnsError("fullName of one receiveUpdateModels is not in zone",
				id, dnsutil.DNS_OPCODE_UPDATE, dnsutil.DNS_RCODE_NOTZONE, transportutil.NEXT_CONNECT_POLICY_KEEP)
		}

		// if (rr.class == zclass)
		if receiveUpdateModels[i].PacketClass == zClass {
			//  if (rr.type & ANY|AXFR|MAILA|MAILB) 			return (FORMERR)
			if receiveUpdateModels[i].PacketType == dnsutil.DNS_TYPE_INT_ANY ||
				receiveUpdateModels[i].PacketType == dnsutil.DNS_TYPE_INT_AXFR ||
				receiveUpdateModels[i].PacketType == dnsutil.DNS_TYPE_INT_MAILA ||
				receiveUpdateModels[i].PacketType == dnsutil.DNS_TYPE_INT_MAILB {
				belogs.Error("performUpdateSectionTransactPrescan():PacketClass==zClass, packetType is ANY or AXFR or MAILA or MAILB, fail: packetType", receiveUpdateModels[i].PacketType,
					jsonutil.MarshalJson(receiveUpdateModels[i]))
				return dnsutil.NewDnsError("when class is equal to zClass, type of one updateModels is ANY or AXFR or MAILA or MAILB",
					id, dnsutil.DNS_OPCODE_UPDATE, dnsutil.DNS_RCODE_FORMERR, transportutil.NEXT_CONNECT_POLICY_KEEP)
			}

		} else if receiveUpdateModels[i].PacketClass == dnsutil.DNS_CLASS_INT_ANY { // if (rr.class == ANY)
			// if (rr.ttl != 0 || rr.rdlength != 0 	|| rr.type & AXFR|MAILA|MAILB)			return (FORMERR)
			if receiveUpdateModels[i].PacketTtl != 0 ||
				receiveUpdateModels[i].PacketDataLength != 0 ||
				receiveUpdateModels[i].PacketType == dnsutil.DNS_TYPE_INT_AXFR ||
				receiveUpdateModels[i].PacketType == dnsutil.DNS_TYPE_INT_MAILA ||
				receiveUpdateModels[i].PacketType == dnsutil.DNS_TYPE_INT_MAILB {
				belogs.Error("performUpdateSectionTransactPrescan(): DNS_CLASS_INT_ANY, packetType is ANY or AXFR or MAILA or MAILB, fail: packetType", receiveUpdateModels[i].PacketType,
					jsonutil.MarshalJson(receiveUpdateModels[i]))
				return dnsutil.NewDnsError("when class is ANY, type of one updateModels is ANY or AXFR or MAILA or MAILB",
					id, dnsutil.DNS_OPCODE_UPDATE, dnsutil.DNS_RCODE_FORMERR, transportutil.NEXT_CONNECT_POLICY_KEEP)
			}
		} else if receiveUpdateModels[i].PacketClass == dnsutil.DNS_CLASS_INT_NONE { // if (rr.class == NONE)
			//  if (rr.ttl != 0 || rr.type & ANY|AXFR|MAILA|MAILB)			return (FORMERR)
			if receiveUpdateModels[i].PacketTtl != 0 ||
				receiveUpdateModels[i].PacketType == dnsutil.DNS_TYPE_INT_AXFR ||
				receiveUpdateModels[i].PacketType == dnsutil.DNS_TYPE_INT_MAILA ||
				receiveUpdateModels[i].PacketType == dnsutil.DNS_TYPE_INT_MAILB {
				belogs.Error("performUpdateSectionTransactPrescan(): DNS_CLASS_INT_NONE, packetType is ANY or AXFR or MAILA or MAILB, fail: packetType", receiveUpdateModels[i].PacketType,
					jsonutil.MarshalJson(receiveUpdateModels[i]))
				return dnsutil.NewDnsError("when class is NONE, type of one updateModels is ANY or AXFR or MAILA or MAILB",
					id, dnsutil.DNS_OPCODE_UPDATE, dnsutil.DNS_RCODE_FORMERR, transportutil.NEXT_CONNECT_POLICY_KEEP)
			}
		} else { //else			return (FORMERR)
			belogs.Error("performUpdateSectionTransactPrescan(): class is not euqal to zClass, or ANY or NONE, fail: PacketClass", receiveUpdateModels[i].PacketClass,
				jsonutil.MarshalJson(receiveUpdateModels[i]))
			return dnsutil.NewDnsError("class is not euqal to zClass, or ANY or NONE",
				id, dnsutil.DNS_OPCODE_UPDATE, dnsutil.DNS_RCODE_FORMERR, transportutil.NEXT_CONNECT_POLICY_KEEP)
		}
	}
	return nil
}
