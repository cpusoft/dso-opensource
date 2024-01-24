package transact

import (
	"dns-model/message"
	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/dnsutil"
	"github.com/cpusoft/goutil/jsonutil"
	"github.com/cpusoft/goutil/transportutil"
	"github.com/cpusoft/goutil/xormdb"
	updatemodel "update-core/model"
)

// err: dnserror
func performZoneSectionTransact(id uint16, receiveZoneModel *updatemodel.ZoneModel, dnsToProcessMsg *message.DnsToProcessMsg) (err error) {
	belogs.Debug("performZoneSectionTransact():id:", id, " receiveZoneModel:", jsonutil.MarshalJson(receiveZoneModel),
		"  dnsToProcessMsg:", jsonutil.MarshalJson(dnsToProcessMsg))
	if dnsToProcessMsg.DnsTransactSide == message.DNS_TRANSACT_SIDE_CLIENT {
		belogs.Debug("performZoneSectionTransact(): is DNS_TRANSACT_TYPE_CLIENT, no check zone")
		return nil
	}
	if receiveZoneModel.ZNamePacketDomain == nil || len(receiveZoneModel.ZNamePacketDomain.FullDomain) == 0 {
		belogs.Error("performZoneSectionTransact(): ZNamePacketDomain is empty:", jsonutil.MarshalJson(receiveZoneModel))
		return dnsutil.NewDnsError("ZName is empty",
			id, dnsutil.DNS_OPCODE_UPDATE, dnsutil.DNS_RCODE_FORMERR, transportutil.NEXT_CONNECT_POLICY_KEEP)
	}
	zName := string(receiveZoneModel.ZNamePacketDomain.FullDomain)
	origin := zName
	belogs.Debug("performZoneSectionTransact(): zName:", zName, "  origin:", origin)

	if receiveZoneModel.ZType != dnsutil.DNS_TYPE_INT_SOA {
		belogs.Error("performZoneSectionTransact(): receiveZoneModel zType is not SOA:", jsonutil.MarshalJson(receiveZoneModel))
		return dnsutil.NewDnsError("ZType is not SOA",
			id, dnsutil.DNS_OPCODE_UPDATE, dnsutil.DNS_RCODE_FORMERR, transportutil.NEXT_CONNECT_POLICY_KEEP)
	}

	if receiveZoneModel.ZClass != dnsutil.DNS_CLASS_INT_IN {
		belogs.Error("performZoneSectionTransact(): receiveZoneModel zClass is not IN:", jsonutil.MarshalJson(receiveZoneModel))
		return dnsutil.NewDnsError("ZClass is not IN",
			id, dnsutil.DNS_OPCODE_UPDATE, dnsutil.DNS_RCODE_FORMERR, transportutil.NEXT_CONNECT_POLICY_KEEP)
	}
	var myCount int
	sql := `select count(*) as mycount from lab_dns_origin   
			where lower(origin) = lower(?)  `
	found, err := xormdb.XormEngine.SQL(sql, origin).Get(&myCount)
	if err != nil {
		belogs.Error("performZoneSectionTransact(): found zName mycount from lab_dns_origin, fail, origin:", origin, err)
		return dnsutil.NewDnsError("fail to found zName in mysql",
			id, dnsutil.DNS_OPCODE_UPDATE, dnsutil.DNS_RCODE_SERVFAIL, transportutil.NEXT_CONNECT_POLICY_KEEP)
	}
	if !found || myCount == 0 {
		belogs.Error("performZoneSectionTransact(): found zName mycount from lab_dns_origin, not found, origin:", origin,
			" found:", found, "  myCount:", myCount)
		return dnsutil.NewDnsError("no auth for '"+zName+"'",
			id, dnsutil.DNS_OPCODE_UPDATE, dnsutil.DNS_RCODE_NOTAUTH, transportutil.NEXT_CONNECT_POLICY_KEEP)
	}
	return nil
}
