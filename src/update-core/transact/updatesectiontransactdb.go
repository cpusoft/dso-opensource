package transact

import (
	"strconv"
	"strings"
	"time"

	"dns-model/packet"
	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/dnsutil"
	"github.com/cpusoft/goutil/jsonutil"
	"github.com/cpusoft/goutil/transportutil"
	"github.com/cpusoft/goutil/xormdb"
	updatemodel "update-core/model"
)

/*
shaodebug

 if rrClass==DNS_CLASS_INT_***, will add
 else if rrClass==DNS_CLASS_INT_ANY
	if rrType==DNS_TYPE_STR_ANY, will delete by fullDomain
	else will delete by fullDomain and rrType
 else if rrClass==DNS_CLASS_INT_NONE, will delete by fullDomain and rrType and RrData
*/
// rfc2136  3.4.2.7 - Pseudocode For Update Section Processing
func performUpdateSectionTransactDb(receiveUpdateRrModel *updatemodel.UpdateRrModel) (err error) {
	belogs.Debug("performUpdateSectionTransactDb():  receiveUpdateRrModel:", jsonutil.MarshalJson(receiveUpdateRrModel))
	start := time.Now()

	id := receiveUpdateRrModel.GetHeaderModel().GetIdOrMessageId()
	receiveZoneRrModel := receiveUpdateRrModel.UpdateDataRrModel.ZoneRrModel
	receiveUpdateRrModels := receiveUpdateRrModel.UpdateDataRrModel.UpdateRrModels
	if receiveZoneRrModel == nil || len(receiveUpdateRrModels) == 0 {
		belogs.Error("performUpdateSectionTransactDb(): receiveZoneRrModel or receiveUpdateRrModels is empty fail:")
		return dnsutil.NewDnsError("zone or receiveUpdate(s) is empty",
			id, dnsutil.DNS_OPCODE_UPDATE, dnsutil.DNS_RCODE_SERVFAIL, transportutil.NEXT_CONNECT_POLICY_KEEP)
	}

	session, err := xormdb.NewSession()
	if err != nil {
		belogs.Error("performUpdateSectionTransactDb(): NewSession fail:", err)
		return dnsutil.NewDnsError("fail to connect to mysql",
			id, dnsutil.DNS_OPCODE_UPDATE, dnsutil.DNS_RCODE_SERVFAIL, transportutil.NEXT_CONNECT_POLICY_KEEP)
	}
	defer session.Close()

	zName := receiveZoneRrModel.RrFullDomain
	origin := zName
	zRrType := receiveZoneRrModel.RrType
	zRrClass := receiveZoneRrModel.RrClass
	var sql string
	var myCount int
	belogs.Debug("performUpdateSectionTransactDb():id:", id, "  zName:", zName, "  origin:", origin,
		"  zRrType:", zRrType, "  zRrClass:", zRrClass)

	for i := range receiveUpdateRrModels {
		fullDomain := receiveUpdateRrModels[i].RrFullDomain
		rrType := receiveUpdateRrModels[i].RrType
		rrClass := receiveUpdateRrModels[i].RrClass
		rrTtl := receiveUpdateRrModels[i].RrTtl
		rrData := receiveUpdateRrModels[i].RrData
		belogs.Debug("performUpdateSectionTransactDb(): will check receiveUpdateRrModels[i]:", jsonutil.MarshalJson(receiveUpdateRrModels[i]),
			"   fullDomain:", fullDomain, "  packetRrTyp:", rrType,
			"   rrClass:", rrClass, "  rrTtl:", rrTtl, "  rrData:", rrData)
		// if (rr.class == zclass)
		if rrClass == zRrClass {
			belogs.Debug("performUpdateSectionTransactDb(): rrClass==zRrClass:", rrClass, "   zRrClass:", zRrClass)

			// if (rr.type == CNAME)
			if rrType == dnsutil.DNS_TYPE_STR_CNAME { //zRrType
				//    if (zone_rrset<rr.name, ~CNAME>) 				next [rr]
				belogs.Debug("performUpdateSectionTransactDb(): rrClass==zRrClass And type==CNAME, rrClass:", rrClass,
					"   zRrClass:", zRrClass, "   zRrType:", zRrType, "  rrType:", rrType)
				sql = `select count(*) as myCount from lab_dns_rr r 
				        where rrFullDomain= ? and rrType != 'CNAME' `
				found, err := session.SQL(sql, fullDomain).Get(&myCount)
				if err != nil {
					dnsErr := dnsutil.NewDnsError(err.Error(),
						id, dnsutil.DNS_OPCODE_UPDATE, dnsutil.DNS_RCODE_SERVFAIL, transportutil.NEXT_CONNECT_POLICY_KEEP)
					return xormdb.RollbackAndLogError(session, "performUpdateSectionTransactDb(): get myCount from lab_dns_rr,no CNAME, fail, fullDomain:"+fullDomain, dnsErr)
				}
				if found {
					// next [rr] ignore
					belogs.Debug("performUpdateSectionTransactDb(): get myCount from lab_dns_rr,no CNAME, myCount>0:ingore, fullDomain:", fullDomain,
						"   myCount:", myCount)
					continue
				}

			} else {
				belogs.Debug("performUpdateSectionTransactDb(): rrClass==zRrClass And type!=CNAME, rrClass:", rrClass,
					"   zRrClass:", zRrClass, "   zRrType:", zRrType, "  rrType:", rrType)
				//   elsif (zone_rrset<rr.name, CNAME>)				next [rr]
				sql = `select count(*) as myCount from lab_dns_rr r 
				where rrFullDomain= ? and rrType = 'CNAME' `
				_, err := session.SQL(sql, fullDomain).Get(&myCount)
				if err != nil {
					dnsErr := dnsutil.NewDnsError(err.Error(),
						id, dnsutil.DNS_OPCODE_UPDATE, dnsutil.DNS_RCODE_SERVFAIL, transportutil.NEXT_CONNECT_POLICY_KEEP)
					return xormdb.RollbackAndLogError(session, "performUpdateSectionTransactDb(): get myCount from lab_dns_rr,CNAME, fail, fullDomain:"+fullDomain, dnsErr)
				}
				if myCount > 0 {
					// next [rr] ignore
					belogs.Debug("performUpdateSectionTransactDb(): get myCount from lab_dns_rr,CNAME, myCount>0:ingore, fullDomain:", fullDomain,
						"   myCount:", myCount)
					continue
				}
			}

			// if (rr.type == SOA)
			if rrType == dnsutil.DNS_TYPE_STR_SOA { // zRrType
				// if (!zone_rrset<rr.name, SOA> )          next [rr]
				belogs.Debug("performUpdateSectionTransactDb(): rrClass==zRrClass And type==SOA, rrClass:", rrClass,
					"   zRrClass:", zRrClass, "   zRrType:", zRrType, "   rrType:", rrType)
				sql = `select count(*) as myCount from lab_dns_rr r 
					where rrFullDomain= ? and rrType = 'SOA' `
				found, err := session.SQL(sql, fullDomain).Get(&myCount)
				if err != nil {
					dnsErr := dnsutil.NewDnsError(err.Error(),
						id, dnsutil.DNS_OPCODE_UPDATE, dnsutil.DNS_RCODE_SERVFAIL, transportutil.NEXT_CONNECT_POLICY_KEEP)
					return xormdb.RollbackAndLogError(session, "performUpdateSectionTransactDb(): get myCount from lab_dns_rr, SOA, fail, fullDomain:"+fullDomain, dnsErr)
				}
				if !found {
					// next [rr] ignore
					belogs.Debug("performUpdateSectionTransactDb(): get myCount from lab_dns_rr,SOA, myCount>0: ingore, fullDomain:", fullDomain,
						"   myCount:", myCount)
					continue
				}

				// if zone_rr<rr.name, SOA>.serial > rr.soa.serial   next [rr]

				soaModel, _, err := packet.NewSoaModelFromRrData(rrData, 0)
				if err != nil {
					dnsErr := dnsutil.NewDnsError(err.Error(),
						id, dnsutil.DNS_OPCODE_UPDATE, dnsutil.DNS_RCODE_SERVFAIL, transportutil.NEXT_CONNECT_POLICY_KEEP)
					return xormdb.RollbackAndLogError(session, "performUpdateSectionTransactDb(): SOA, get soaModel fail, rrData:"+rrData, dnsErr)
				}
				//
				var rrDatas []string
				belogs.Debug("performUpdateSectionTransactDb(): rrClass==zRrClass And type==SOA, soaModel:", jsonutil.MarshalJson(soaModel))
				sql = `select  rrData from lab_dns_rr r 
					where rrFullDomain= ? and rrType = 'SOA'  `
				err = session.SQL(sql, fullDomain).Find(&rrDatas)
				if err != nil {
					dnsErr := dnsutil.NewDnsError(err.Error(),
						id, dnsutil.DNS_OPCODE_UPDATE, dnsutil.DNS_RCODE_SERVFAIL, transportutil.NEXT_CONNECT_POLICY_KEEP)
					return xormdb.RollbackAndLogError(session, "performUpdateSectionTransactDb(): get rrData from lab_dns_rr, SOA, fail, fullDomain:"+fullDomain, dnsErr)
				}
				for i := range rrDatas {
					split := strings.Split(rrDatas[i], " ")
					if len(split) > 3 {
						serial, err := strconv.Atoi(split[2])
						if err != nil {
							dnsErr := dnsutil.NewDnsError(err.Error(),
								id, dnsutil.DNS_OPCODE_UPDATE, dnsutil.DNS_RCODE_SERVFAIL, transportutil.NEXT_CONNECT_POLICY_KEEP)
							return xormdb.RollbackAndLogError(session, "performUpdateSectionTransactDb(): get serial from rrData, SOA, fail, rrData: "+jsonutil.MarshalJson(rrDatas[i]), dnsErr)
						}
						if serial > int(soaModel.Serial) {
							belogs.Debug("performUpdateSectionTransactDb():serial > soaModel.Serial ,SOA, ingore:", serial,
								"   soaModel.Serial:", soaModel.Serial)
							continue
						}
					}
				}
			}

			// if (rr.type == CNAME || rr.type == SOA ||
			//	(rr.type == WKS && rr.proto == zrr.proto &&	rr.address == zrr.address) ||  // will ignore WKS
			//   rr.rdata == zrr.rdata)

			// different:
			// if CNAME or SOA, will replace current
			// else added or update(if exists)
			if rrType == dnsutil.DNS_TYPE_STR_SOA || rrType == dnsutil.DNS_TYPE_STR_CNAME { //zRrType
				belogs.Debug("performUpdateSectionTransactDb(): rrClass==zRrClass, will update lab_dns_rr, CNAME/ROA, rrTtl:", rrTtl,
					"  zRrType:", zRrType,
					"  rrData:", rrData, "  fullDomain:", fullDomain,
					"  rrClass:", rrClass, "  rrType:", rrType)
				sql = `update lab_dns_rr set rrTtl=?, rrData=? 
					where  rrFullDomain=? and rrCLass=? and rrType=? 
					 and rrType in ('SOA','CNAME') order by id `
				_, err = session.Exec(sql, rrTtl, rrData,
					fullDomain, rrClass, rrType)
				if err != nil {
					dnsErr := dnsutil.NewDnsError(err.Error(),
						id, dnsutil.DNS_OPCODE_UPDATE, dnsutil.DNS_RCODE_SERVFAIL, transportutil.NEXT_CONNECT_POLICY_KEEP)
					return xormdb.RollbackAndLogError(session, "performUpdateSectionTransactDb(): update rrClass==zRrClass,,lab_dns_rr, 'SOA','CNAME', fail, receiveUpdateRrModels[i]:"+jsonutil.MarshalJson(receiveUpdateRrModels[i]), dnsErr)
				}
				belogs.Debug("performUpdateSectionTransactDb(): rrClass==zRrClass, update lab_dns_rr ok, CNAME/ROA, receiveRrModel:", jsonutil.MarshalJson(receiveUpdateRrModels[i]))
			} else {
				belogs.Debug("performUpdateSectionTransactDb(): rrClass==zRrClass, normal update, will update lab_dns_rr,no CNAME/ROA, rrTtl:", rrTtl,
					"  zRrType:", zRrType,
					"  rrData:", rrData, "  fullDomain:", fullDomain,
					"  rrClass:", rrClass, "  rrType:", rrType)
				// get originId
				var originId int
				sql = `select id as originId from lab_dns_origin where origin = ? `
				_, err = session.SQL(sql, origin).Get(&originId) // end with "."
				if err != nil {
					dnsErr := dnsutil.NewDnsError(err.Error(),
						id, dnsutil.DNS_OPCODE_UPDATE, dnsutil.DNS_RCODE_SERVFAIL, transportutil.NEXT_CONNECT_POLICY_KEEP)
					return xormdb.RollbackAndLogError(session, "performUpdateSectionTransactDb(): rrClass==zRrClass, get originid from lab_dns_origin, zName:"+zName, dnsErr)
				}
				// will add ,when is not SOA/CNAME

				myCount := 0
				sql = `select count(*) as myCount from lab_dns_rr where rrFullDomain=? and rrType=? and rrData=? `
				_, err := session.SQL(sql, fullDomain, rrType, rrData).Get(&myCount)
				if err != nil {
					dnsErr := dnsutil.NewDnsError(err.Error(),
						id, dnsutil.DNS_OPCODE_UPDATE, dnsutil.DNS_RCODE_SERVFAIL, transportutil.NEXT_CONNECT_POLICY_KEEP)
					return xormdb.RollbackAndLogError(session, "performUpdateSectionTransactDb(): rrClass==zRrClass, select count from lab_dns_rr,  not CNAME/ROA, fail, receiveRrModel:"+jsonutil.MarshalJson(receiveUpdateRrModels[i]), dnsErr)
				}
				belogs.Debug("performUpdateSectionTransactDb(): rrClass==zRrClass, select count from lab_dns_rr, myCount:", myCount)
				if myCount > 0 {
					belogs.Debug("performUpdateSectionTransactDb(): rrClass==zRrClass, select count from lab_dns_rr,will continue, myCount >0:", myCount)
					continue
				}

				rrName := strings.TrimSuffix(fullDomain, zName)
				rrName = strings.TrimSuffix(rrName, ".")
				belogs.Debug("performUpdateSectionTransactDb(): rrClass==zRrClass,will insert lab_dns_rr , not CNAME/ROA,",
					"  originId:", originId, "  rrName:", rrName, "  fullDomain:", fullDomain,
					"  rrType:", rrType, "  rrClass:", rrClass, "  rrTtl:", rrTtl,
					"  rrData:", rrData, start)

				sql = `insert into lab_dns_rr (originId,rrName,rrFullDomain,
					     rrType,rrClass,rrTtl,
						 rrData,updateTime) 
					   values(?,?,?,   ?,?,?,   ?,?)`
				_, err = session.Exec(sql, originId, xormdb.SqlNullString(rrName), fullDomain,
					rrType, rrClass, rrTtl,
					rrData, start)
				if err != nil {
					dnsErr := dnsutil.NewDnsError(err.Error(),
						id, dnsutil.DNS_OPCODE_UPDATE, dnsutil.DNS_RCODE_SERVFAIL, transportutil.NEXT_CONNECT_POLICY_KEEP)
					return xormdb.RollbackAndLogError(session, "performUpdateSectionTransactDb(): rrClass==zRrClass, insert lab_dns_rr,  not CNAME/ROA, fail, receiveRrModel:"+jsonutil.MarshalJson(receiveUpdateRrModels[i]), dnsErr)
				}
				belogs.Debug("performUpdateSectionTransactDb(): rrClass==zRrClass, insert lab_dns_rr, not CNAME/ROA, ok, receiveRrModel:", jsonutil.MarshalJson(receiveUpdateRrModels[i]))
			}

		} else if rrClass == dnsutil.DNS_CLASS_STR_ANY { // elsif (rr.class == ANY)
			belogs.Debug("performUpdateSectionTransactDb(): rrClass==CLASS_ANY, rrClass:", rrClass)

			//if (rr.type == ANY)
			//if zRrType == dnsutil.DNS_TYPE_STR_ANY {
			if rrType == dnsutil.DNS_TYPE_STR_ANY {
				belogs.Debug("performUpdateSectionTransactDb(): rrClass==CLASS_ANY && zRrType==CLASS_ANY,rrClass:", rrClass, "   zRrType:", zRrType)
				//  if (rr.name == zname)			zone_rrset<rr.name, ~(SOA|NS)> = Nil
				if fullDomain == zName {
					belogs.Debug("performUpdateSectionTransactDb(): rrClass==CLASS_ANY && zRrType==TYPE_ANY && fullDomain==zName, will delete lab_dns_rr no(SOA/NS)",
						"  rrClass:", rrClass, "   rrType:", rrType, "  fullDomain:", fullDomain, "   zName:", zName)
					sql := `delete from lab_dns_rr r 
						where rrFullDomain= ? and rrType not in ('SOA','NS') `
					_, err = session.Exec(sql, fullDomain)
					if err != nil {
						dnsErr := dnsutil.NewDnsError(err.Error(),
							id, dnsutil.DNS_OPCODE_UPDATE, dnsutil.DNS_RCODE_SERVFAIL, transportutil.NEXT_CONNECT_POLICY_KEEP)
						return xormdb.RollbackAndLogError(session, "performUpdateSectionTransactDb(): delete from lab_dns_rr,CLASS_ANY, TYPE_ANY, not in ('SOA','NS') fail, fullDomain:"+fullDomain, dnsErr)
					}
					belogs.Debug("performUpdateSectionTransactDb(): rrClass==CLASS_ANY && zRrType==TYPE_ANY && fullDomain==zName,  delete lab_dns_rr no(SOA/NS) ok,",
						"  rrClass:", rrClass, "   rrType:", rrType, "  fullDomain:", fullDomain, "   zName:", zName)

				} else {
					belogs.Debug("performUpdateSectionTransactDb(): rrClass==CLASS_ANY && zRrType==CLASS_ANY && fullDomain!=zName",
						"  rrClass:", rrClass, "   rrType:", rrType, "  fullDomain:", fullDomain, "   zName:", zName)
					sql := `delete from lab_dns_rr r 
				       where rrFullDomain= ? `
					_, err = session.Exec(sql, fullDomain)
					if err != nil {
						dnsErr := dnsutil.NewDnsError(err.Error(),
							id, dnsutil.DNS_OPCODE_UPDATE, dnsutil.DNS_RCODE_SERVFAIL, transportutil.NEXT_CONNECT_POLICY_KEEP)
						return xormdb.RollbackAndLogError(session, "performUpdateSectionTransactDb(): delete from lab_dns_rr,CLASS_ANY, fail, fullDomain:"+fullDomain, dnsErr)
					}
					belogs.Debug("performUpdateSectionTransactDb(): rrClass==CLASS_ANY && zRrType==TYPE_ANY , delete lab_dns_rr ok,",
						"  rrClass:", rrClass, "   rrType:", rrType, "  fullDomain:", fullDomain, "   zName:", zName)
				}
			} else if fullDomain == zName &&
				(zRrType == dnsutil.DNS_TYPE_STR_SOA ||
					zRrType == dnsutil.DNS_TYPE_STR_NS) {
				//  elsif (rr.name == zname &&	(rr.type == SOA || rr.type == NS))
				belogs.Debug("performUpdateSectionTransactDb(): rrClass==CLASS_ANY && zRrType==TYPE_ANY, SOA/NS, will ignore,",
					"  rrClass:", rrClass, "   rrType:", rrType, "  fullDomain:", fullDomain, "   zName:", zName)
				continue
			} else {
				//	else zone_rrset<rr.name, *> = Nil
				belogs.Debug("performUpdateSectionTransactDb(): rrClass==CLASS_ANY && zRrType==CLASS_ANY , will delete:",
					"  rrClass:", rrClass, "   rrType:", rrType, "  fullDomain:", fullDomain, "   zName:", zName)
				sql := `delete from lab_dns_rr r 
			  		    where rrFullDomain= ? and rrType=? `
				_, err = session.Exec(sql, fullDomain, rrType)
				if err != nil {
					dnsErr := dnsutil.NewDnsError(err.Error(),
						id, dnsutil.DNS_OPCODE_UPDATE, dnsutil.DNS_RCODE_SERVFAIL, transportutil.NEXT_CONNECT_POLICY_KEEP)
					return xormdb.RollbackAndLogError(session, "performUpdateSectionTransactDb(): rrClass==CLASS_ANY && zRrType==CLASS_ANY,delete from lab_dns_rr fail, fullDomain:"+fullDomain+"  rrType:"+rrType, dnsErr)
				}
				belogs.Debug("performUpdateSectionTransactDb(): rrClass==CLASS_ANY && zRrType==TYPE_ANY else , delete lab_dns_rr ok,",
					"  rrClass:", rrClass, "   zRrType:", zRrType, "  fullDomain:", fullDomain, "   zName:", zName)
			}
		} else if rrClass == dnsutil.DNS_CLASS_STR_NONE { // elsif (rr.class == NONE)
			//  elsif (rr.class == NONE)
			belogs.Debug("performUpdateSectionTransactDb(): rrClass==CLASS_NONE")
			if rrType == dnsutil.DNS_TYPE_STR_SOA { //zRrType
				belogs.Debug("performUpdateSectionTransactDb(): rrClass==CLASS_NONE and rrType==SOA, ignore: zRrType:", zRrType)
				continue
			} else if rrType == dnsutil.DNS_TYPE_STR_NS { //zRrType
				// NAME, TYPE, RDATA and RDLENGTH
				belogs.Debug("performUpdateSectionTransactDb(): rrClass==CLASS_NONE And rrType==NS, rrClass:", rrClass,
					"   zRrClass:", zRrClass, "   zRrType:", zRrType)
				sql = `select count(*) as myCount from lab_dns_rr r 
				where rrFullDomain= ? and rrType = ? and rrData=? `
				_, err := session.SQL(sql, fullDomain, rrType, rrData).Get(&myCount)
				if err != nil {
					dnsErr := dnsutil.NewDnsError(err.Error(),
						id, dnsutil.DNS_OPCODE_UPDATE, dnsutil.DNS_RCODE_SERVFAIL, transportutil.NEXT_CONNECT_POLICY_KEEP)
					return xormdb.RollbackAndLogError(session, "performUpdateSectionTransactDb(): get myCount from lab_dns_rr, rrClass==CLASS_NONE And type==NS, fail, fullDomain:"+fullDomain, dnsErr)
				}
				// only one, will ignore
				if myCount == 1 {
					// next [rr] ignore
					belogs.Debug("performUpdateSectionTransactDb(): get myCount from lab_dns_rr,rrClass==CLASS_NONE And type==NS, myCount==1: ingore, fullDomain:", fullDomain,
						"   myCount:", myCount)
					continue
				}
			}
			sql := `delete from lab_dns_rr r 
				       where rrFullDomain= ?  and rrType = ? and rrData=? `
			belogs.Debug("performUpdateSectionTransactDb():will delete from lab_dns_rr,CLASS_ANY, sql:", sql,
				"  fullDomain:"+fullDomain, "  rrType:", rrType, " rrData:", rrData)
			_, err = session.Exec(sql, fullDomain, rrType, rrData)
			if err != nil {
				dnsErr := dnsutil.NewDnsError(err.Error(),
					id, dnsutil.DNS_OPCODE_UPDATE, dnsutil.DNS_RCODE_SERVFAIL, transportutil.NEXT_CONNECT_POLICY_KEEP)
				return xormdb.RollbackAndLogError(session, "performUpdateSectionTransactDb(): delete from lab_dns_rr,CLASS_ANY, fail, fullDomain:"+fullDomain, dnsErr)
			}
			belogs.Debug("performUpdateSectionTransactDb():delete ok from lab_dns_rr,CLASS_ANY,  fullDomain:"+fullDomain, "  rrType:", rrType, " rrData:", rrData)

		}
	}

	err = xormdb.CommitSession(session)
	if err != nil {
		belogs.Error("performUpdateSectionTransactDb(): CommitSession fail :", err)
		return dnsutil.NewDnsError(err.Error(),
			id, dnsutil.DNS_OPCODE_UPDATE, dnsutil.DNS_RCODE_SERVFAIL, transportutil.NEXT_CONNECT_POLICY_KEEP)
	}
	belogs.Debug("performUpdateSectionTransactDb(): receiveUpdateRrModel:", jsonutil.MarshalJson(receiveUpdateRrModel), " ok,  time(s):", time.Since(start))

	return nil
}
