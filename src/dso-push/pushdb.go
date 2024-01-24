package push

import (
	"dns-model/rr"
	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/dnsutil"
	"github.com/cpusoft/goutil/jsonutil"
	"github.com/cpusoft/goutil/xormdb"
)

func queryDb(rrFullDomain, rrType string) (resultRrModels []*rr.RrModel, err error) {

	/* TODO
	根据lab_dns_dependent查询的依赖关系，修改下面的SQL，返回查询结果
	*/
	resultRrModels = make([]*rr.RrModel, 0)
	//	resultRtModelTmps := make([]*RrModelTmp, 0)
	if rrType == dnsutil.DNS_TYPE_STR_ANY {
		sql := `select o.origin, r.rrName, r.rrFullDomain, r.rrType, r.rrClass, IFNULL(r.rrTtl,o.ttl) as rrTtl, r.rrData  
			from lab_dns_rr r,	lab_dns_origin o ,lab_dns_rr_dependent d
			where r.originId = o.id and d.rrFullDomain = ? and (find_in_set(r.id,d.childId) or r.rrFullDomain=d.rrFullDomain)
			group by r.id`
		belogs.Debug("queryDb(): DNS_TYPE_STR_ANY rrType:", rrType, "  rrFullDomain:", rrFullDomain, " sql:", sql)
		err = xormdb.XormEngine.SQL(sql, rrFullDomain).Find(&resultRrModels)
	} else {
		sql := `select o.origin, r.rrName, r.rrFullDomain, r.rrType, r.rrClass, IFNULL(r.rrTtl,o.ttl) as rrTtl, r.rrData  
			from lab_dns_rr r,	lab_dns_origin o ,lab_dns_rr_dependent d
			where r.originId = o.id and d.rrFullDomain = ? and r.rrType = ? and (find_in_set(r.id,d.childId) or r.rrFullDomain=d.rrFullDomain)
			group by r.id`
		belogs.Debug("queryDb(): DNS_TYPE_STR_*** rrType:", rrType, "  rrFullDomain:", rrFullDomain, " sql:", sql)
		err = xormdb.XormEngine.SQL(sql, rrFullDomain, rrType).Find(&resultRrModels)
	}
	if err != nil {
		belogs.Error("queryDb(): lab_dns_rr fail:", err)
		return nil, err
	}
	belogs.Debug("queryDb(): resultRrModels:", jsonutil.MarshalJson(resultRrModels))
	return resultRrModels, nil
}

func queryAllDb() (resultRrModels []*rr.RrModel, err error) {
	resultRrModels = make([]*rr.RrModel, 0)
	sql := `select o.origin, r.rrName, r.rrFullDomain, r.rrType, r.rrClass, IFNULL(r.rrTtl,o.ttl) as rrTtl, r.rrData  
			from lab_dns_rr r,	lab_dns_origin o 
			where r.originId = o.id   
			group by r.id `
	belogs.Debug("queryAllDb():  sql:", sql)
	err = xormdb.XormEngine.SQL(sql).Find(&resultRrModels)
	if err != nil {
		belogs.Error("queryAllDb(): lab_dns_rr fail:", err)
		return nil, err
	}
	belogs.Debug("queryAllDb(): resultRrModels:", jsonutil.MarshalJson(resultRrModels))
	return resultRrModels, nil
}
