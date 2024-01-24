package transact

import (
	"strings"

	"dns-model/message"
	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/dnsutil"
	"github.com/cpusoft/goutil/jsonutil"
	"github.com/cpusoft/goutil/transportutil"
	updatemodel "update-core/model"
)

func performPrerequisiteSectionTransact(receiveUpdateModel *updatemodel.UpdateModel, dnsToProcessMsg *message.DnsToProcessMsg) (err error) {
	belogs.Debug("performPrerequisiteSectionTransact(): receiveUpdateModel:", jsonutil.MarshalJson(receiveUpdateModel), "  dnsToProcessMsg:", jsonutil.MarshalJson(dnsToProcessMsg))
	if dnsToProcessMsg.DnsTransactSide == message.DNS_TRANSACT_SIDE_CLIENT {
		belogs.Debug("performPrerequisiteSectionTransact(): is DNS_TRANSACT_TYPE_CLIENT, no check zone")
		return nil
	}
	id := receiveUpdateModel.GetHeaderModel().GetIdOrMessageId()
	receiveZoneModel := receiveUpdateModel.UpdateDataModel.ZoneModel
	receivePrerequisiteModels := receiveUpdateModel.UpdateDataModel.PrerequisiteModels
	if len(receivePrerequisiteModels) == 0 {
		belogs.Error("performPrerequisiteSectionTransact():len(receivePrerequisiteModels)==0, fail:", jsonutil.MarshalJson(receivePrerequisiteModels))
		return dnsutil.NewDnsError("receivePrerequisite is empty",
			id, dnsutil.DNS_OPCODE_UPDATE, dnsutil.DNS_RCODE_FORMERR, transportutil.NEXT_CONNECT_POLICY_KEEP)
	}
	zName := strings.ToLower(string(receiveZoneModel.ZNamePacketDomain.FullDomain))
	zType := receiveZoneModel.ZType
	zClass := receiveZoneModel.ZClass
	belogs.Debug("performPrerequisiteSectionTransact():id:", id, " zName:", zName, "  zType:", zType, "  zClass:", zClass)

	// rfc2136: 3.2.5
	checkFoundInClassAny := false
	foundSameNameInClassAny := false
	foundSameNameTypeInClassAny := false

	checkNoFoundInClassNone := false
	noFoundSameNameInClassNone := true
	noFoundSameNameTypeInClassNone := true
	for i := range receivePrerequisiteModels {
		// if (rr.ttl != 0) 	return (FORMERR)
		if receivePrerequisiteModels[i].PacketTtl != 0 {
			belogs.Error("performPrerequisiteSectionTransact():PacketTtl is not 0,fail, ttl:", receivePrerequisiteModels[i].PacketTtl)
			return dnsutil.NewDnsError("TTL of one receivePrerequisite is not 0",
				id, dnsutil.DNS_OPCODE_UPDATE, dnsutil.DNS_RCODE_FORMERR, transportutil.NEXT_CONNECT_POLICY_KEEP)
		}

		//  if (zone_of(rr.name) != ZNAME)		return (NOTZONE);
		if len(receivePrerequisiteModels[i].PacketDomain.FullDomain) == 0 {
			belogs.Error("performPrerequisiteSectionTransact(): len(receivePrerequisiteModels[i].PacketDomain.FullDomain) is empty, fail, receivePrerequisiteModels[i]:",
				jsonutil.MarshalJson(receivePrerequisiteModels[i]))
			return dnsutil.NewDnsError("Name of one receivePrerequisite is empty",
				id, dnsutil.DNS_OPCODE_UPDATE, dnsutil.DNS_RCODE_FORMERR, transportutil.NEXT_CONNECT_POLICY_KEEP)
		}
		fullName := strings.ToLower(string(receivePrerequisiteModels[i].PacketDomain.FullDomain))
		belogs.Debug("performPrerequisiteSectionTransact(): fullName:", fullName)
		if !strings.Contains(fullName, zName) {
			belogs.Error("performPrerequisiteSectionTransact(): fullName is not in zName, fail, fullName:", fullName, "  zName:", zName,
				jsonutil.MarshalJson(receivePrerequisiteModels[i]))
			return dnsutil.NewDnsError("fullName of one receivePrerequisite is not in zone",
				id, dnsutil.DNS_OPCODE_UPDATE, dnsutil.DNS_RCODE_FORMERR, transportutil.NEXT_CONNECT_POLICY_KEEP)
		}

		// class
		// if (rr.class == ANY)
		if receivePrerequisiteModels[i].PacketClass == dnsutil.DNS_CLASS_INT_ANY {
			//  if (rr.rdlength != 0) 			return (FORMERR)
			if receivePrerequisiteModels[i].PacketDataLength != 0 {
				belogs.Error("performPrerequisiteSectionTransact(): class_Any PacketDataLength is not 0, fail, PacketDataLength:",
					receivePrerequisiteModels[i].PacketDataLength,
					"  receivePrerequisiteModels[i]:", jsonutil.MarshalJson(receivePrerequisiteModels[i]))
				return dnsutil.NewDnsError("DataLength is not 0 of class_Any",
					id, dnsutil.DNS_OPCODE_UPDATE, dnsutil.DNS_RCODE_FORMERR, transportutil.NEXT_CONNECT_POLICY_KEEP)
			}

			// if (rr.type == ANY) and (!zone_name<rr.name>) then			return (NXDOMAIN)
			checkFoundInClassAny = true
			if receivePrerequisiteModels[i].PacketType == dnsutil.DNS_TYPE_INT_ANY {
				if !foundSameNameInClassAny && zName == fullName {
					foundSameNameInClassAny = true
					belogs.Debug("performPrerequisiteSectionTransact(): class_Any found name in pre is same to name in zone: zName:", zName,
						"  fullName:", fullName)
				}
			} else {
				//  if (!zone_rrset<rr.name, rr.type>)			return (NXRRSET)
				if !foundSameNameTypeInClassAny &&
					(zName == fullName && zType == receivePrerequisiteModels[i].PacketType) {
					foundSameNameTypeInClassAny = true
					belogs.Debug("performPrerequisiteSectionTransact(): class_Any found type&name in pre is same to type&name in zone: zName:", zName,
						"  fullName:", fullName, "  zType:", zType,
						"  packetType:", receivePrerequisiteModels[i].PacketType)
				}

			}
		} else if receivePrerequisiteModels[i].PacketClass == dnsutil.DNS_CLASS_INT_NONE {
			// if (rr.class == NONE)
			//  if (rr.rdlength != 0) 			return (FORMERR)
			if receivePrerequisiteModels[i].PacketDataLength != 0 {
				belogs.Error("performPrerequisiteSectionTransact(): class_None PacketDataLength is not 0, fail, PacketDataLength:",
					receivePrerequisiteModels[i].PacketDataLength,
					"  receivePrerequisiteModels[i]:", jsonutil.MarshalJson(receivePrerequisiteModels[i]))
				return dnsutil.NewDnsError("DataLength is not 0 of class_None",
					id, dnsutil.DNS_OPCODE_UPDATE, dnsutil.DNS_RCODE_FORMERR, transportutil.NEXT_CONNECT_POLICY_KEEP)
			}
			checkNoFoundInClassNone = true
			// if (rr.type == ANY) and (zone_name<rr.name>) then			return (YXDOMAIN)
			belogs.Debug("performPrerequisiteSectionTransact(): class_None, receivePrerequisiteModels[i].PacketType:", receivePrerequisiteModels[i].PacketType,
				"  noFoundSameNameInClassNone:", noFoundSameNameInClassNone,
				"  zName:", zName, "  fullName:", fullName)
			if receivePrerequisiteModels[i].PacketType == dnsutil.DNS_TYPE_INT_ANY {
				if noFoundSameNameInClassNone && zName == fullName {
					noFoundSameNameInClassNone = false
					belogs.Debug("performPrerequisiteSectionTransact(): class_None found name in pre is same to name in zone: zName:", zName,
						"  fullName:", fullName)
				}
			} else {
				belogs.Debug("performPrerequisiteSectionTransact(): class_None, noFoundSameNameInClassNone:", noFoundSameNameInClassNone,
					"  zName:", zName, "  fullName:", fullName,
					"  zType:", zType, "  receivePrerequisiteModels[i].PacketType:", receivePrerequisiteModels[i].PacketType)
				//  if (zone_rrset<rr.name, rr.type>)			return (YXRRSET)
				if noFoundSameNameTypeInClassNone &&
					(zName == fullName && zType == receivePrerequisiteModels[i].PacketType) {
					noFoundSameNameTypeInClassNone = false
					belogs.Debug("performPrerequisiteSectionTransact(): class_None found type&name in pre is same to type&name in zone: zName:", zName,
						"  fullName:", fullName, "  zType:", zType,
						"  packetType:", receivePrerequisiteModels[i].PacketType)
				}

			}
		} else if receivePrerequisiteModels[i].PacketClass == zClass {
			/*  shaodebug
			Then, build an RRset for each unique <NAME,TYPE> and
			compare each resulting RRset for set equality (same members, no more,
			no less) with RRsets in the zone.  If any Prerequisite RRset is not
			entirely and exactly matched by a zone RRset, signal NXRRSET to the
			requestor.
			?????
			if (rr.class == zclass)
				temp<rr.name, rr.type> += rr
					for rrset in temp
			if (zone_rrset<rrset.name, rrset.type> != rrset)
				return (NXRRSET)
			*/

		} else {
			// not ANY/NONE/zCLASS
			belogs.Error("performPrerequisiteSectionTransact(): class is not ANY/NONE/zCLASS is fail:packetClass",
				receivePrerequisiteModels[i].PacketClass, "  receivePrerequisiteModels[i]:", jsonutil.MarshalJson(receivePrerequisiteModels[i]))
			return dnsutil.NewDnsError("packetClass is not ANY/NONE/zClASS",
				id, dnsutil.DNS_OPCODE_UPDATE, dnsutil.DNS_RCODE_FORMERR, transportutil.NEXT_CONNECT_POLICY_KEEP)
		}
	}
	belogs.Debug("performPrerequisiteSectionTransact(): class_None checkFoundInClassAny:", checkFoundInClassAny,
		"  foundSameNameInClassAny:", foundSameNameInClassAny, "   foundSameNameTypeInClassAny:", foundSameNameTypeInClassAny)
	if checkFoundInClassAny && !foundSameNameInClassAny {
		belogs.Error("performPrerequisiteSectionTransact(): class_Any no name in pre is same to name in zone, checkFoundInClassAny && !foundSameNameInClassAny, fail: zName:", zName,
			"  receivePrerequisiteModels:", jsonutil.MarshalJson(receivePrerequisiteModels))
		return dnsutil.NewDnsError("no name in pre is same to name in zone of class_Any",
			id, dnsutil.DNS_OPCODE_UPDATE, dnsutil.DNS_RCODE_NXDOMAIN, transportutil.NEXT_CONNECT_POLICY_KEEP)
	}
	if checkFoundInClassAny && !foundSameNameTypeInClassAny {
		belogs.Error("performPrerequisiteSectionTransact(): class_Any no type&name in pre is same to type&name in zone, checkFoundInClassAny && !foundSameNameTypeInClassAny, fail: zName:", zName,
			" zType:", zType,
			"  receivePrerequisiteModels:", jsonutil.MarshalJson(receivePrerequisiteModels))
		return dnsutil.NewDnsError("no type&name in pre is same to type&name in zone of class_Any",
			id, dnsutil.DNS_OPCODE_UPDATE, dnsutil.DNS_RCODE_NXRRSET, transportutil.NEXT_CONNECT_POLICY_KEEP)
	}
	// shaodebug, check again
	belogs.Debug("performPrerequisiteSectionTransact(): class_None checkNoFoundInClassNone:", checkNoFoundInClassNone,
		"  noFoundSameNameInClassNone:", noFoundSameNameInClassNone, "   noFoundSameNameTypeInClassNone:", noFoundSameNameTypeInClassNone)
	if checkNoFoundInClassNone && noFoundSameNameInClassNone {
		belogs.Error("performPrerequisiteSectionTransact(): class_None there is name in pre is same to name in zone, checkNoFoundInClassNone && noFoundSameNameInClassNone, fail: zName:", zName,
			"  receivePrerequisiteModels:", jsonutil.MarshalJson(receivePrerequisiteModels))
		return dnsutil.NewDnsError("no name in pre is same to name in zone of class_None",
			id, dnsutil.DNS_OPCODE_UPDATE, dnsutil.DNS_RCODE_YXDOMAIN, transportutil.NEXT_CONNECT_POLICY_KEEP)
	} else {
		belogs.Debug("performPrerequisiteSectionTransact(): class_None name pass,checkNoFoundInClassNone:", checkNoFoundInClassNone,
			"  noFoundSameNameInClassNone:", noFoundSameNameInClassNone)
		return
	}
	if checkNoFoundInClassNone && noFoundSameNameTypeInClassNone {
		belogs.Error("performPrerequisiteSectionTransact(): class_None there is type&name in pre is same to type&name in zone, checkNoFoundInClassNone && noFoundSameNameTypeInClassNone, fail: zName:", zName,
			"  zType:", zType,
			"  receivePrerequisiteModels:", jsonutil.MarshalJson(receivePrerequisiteModels))
		return dnsutil.NewDnsError("no type&name in pre is same to type&name in zone of class_None",
			id, dnsutil.DNS_OPCODE_UPDATE, dnsutil.DNS_RCODE_YXRRSET, transportutil.NEXT_CONNECT_POLICY_KEEP)
	} else {
		belogs.Debug("performPrerequisiteSectionTransact(): class_None type&name pass,checkNoFoundInClassNone:", checkNoFoundInClassNone,
			"  noFoundSameNameTypeInClassNone:", noFoundSameNameTypeInClassNone)
		return
	}
	belogs.Info("performPrerequisiteSectionTransact(): ok, receiveUpdateModel:", jsonutil.MarshalJson(receiveUpdateModel), "  dnsToProcessMsg:", jsonutil.MarshalJson(dnsToProcessMsg))
	return
}
