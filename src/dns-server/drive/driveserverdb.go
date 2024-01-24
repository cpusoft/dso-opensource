package drive

import (
	"bytes"
	"time"

	"dns-model/rr"
	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/convert"
	"github.com/cpusoft/goutil/dnsutil"
	"github.com/cpusoft/goutil/jsonutil"
	"github.com/cpusoft/goutil/xormdb"
	"github.com/guregu/null"
)

func queryRrModelsDb(rrModel *rr.RrModel) ([]*rr.RrModel, error) {
	belogs.Debug("queryRrModelsDb(): rrModel:", jsonutil.MarshalJson(rrModel))

	rrModels := make([]*rr.RrModel, 0)

	eng := xormdb.XormEngine.Table("lab_dns_rr").Where(" 1 = 1 ")
	if len(rrModel.RrFullDomain) > 0 {
		eng = eng.And(` rrFullDomain = ? `, rrModel.RrFullDomain)
	}
	if len(rrModel.RrType) > 0 {
		eng = eng.And(` rrType = ? `, rrModel.RrType)
	}
	if len(rrModel.RrData) > 0 {
		eng = eng.And(` rrData = ? `, rrModel.RrData)
	}

	err := eng.Cols("rrFullDomain, rrType, rrClass, rrTtl, rrData").OrderBy("id").Find(&rrModels)
	if err != nil {
		belogs.Error("queryRrModelsDb(): select lab_dns_rr fail:", jsonutil.MarshalJson(rrModel), err)
		return nil, err
	}

	belogs.Debug("queryRrModelsDb(): len(rrModels):", len(rrModels), "  rrModels:", jsonutil.MarshalJson(rrModels))
	return rrModels, nil
}

func queryAllRrModelsDb() ([]*rr.RrModel, error) {
	belogs.Debug("queryAllRrModelsDb():")
	rrModels := make([]*rr.RrModel, 0)
	err := xormdb.XormEngine.Table("lab_dns_rr").Cols("rrFullDomain, rrType, rrClass, rrTtl, rrData").OrderBy("id").Find(&rrModels)
	if err != nil {
		belogs.Error("queryAllRrModelsDb(): select lab_dns_rr fail:", jsonutil.MarshalJson(rrModels), err)
		return nil, err
	}

	belogs.Debug("queryAllRrModelsDb(): len(rrModels):", len(rrModels), "  rrModels:", jsonutil.MarshalJson(rrModels))
	return rrModels, nil
}

func saveRrDependentModelsDb(rrDependentModels []*RrDependentModel) error {
	start := time.Now()
	belogs.Debug("saveRrDependentModelsDb(): len(rrDependentModels):", len(rrDependentModels))

	session, err := xormdb.NewSession()
	if err != nil {
		belogs.Error("saveRrDependentModelsDb(): NewSession fail: err:", err)
		return err
	}
	defer session.Close()

	sql := `delete from lab_dns_rr_dependent`
	_, err = session.Exec(sql)
	if err != nil {
		belogs.Error("saveRrDependentModelsDb(): delete rr_dependent fail:", sql, err)
		return xormdb.RollbackAndLogError(session, "delete rr_dependent fail:", err)
	}
	sql = `insert into lab_dns_rr_dependent(id,rrFullDomain,childId) values (?,?,?)`
	for i := range rrDependentModels {
		var buffer bytes.Buffer
		for j := 0; j < len(rrDependentModels[i].ChildIds); j++ {
			buffer.WriteString(convert.ToString(rrDependentModels[i].ChildIds[j]))
			if i < len(rrDependentModels[i].ChildIds)-1 {
				buffer.WriteString(",")
			}
		}
		childIds := buffer.String()
		rrId, err := queryByRrFullDomain(rrDependentModels[i].RrFullDomain)
		if err != nil {
			belogs.Error("saveRrDependentModelsDb(): queryByRrFullDomain fail, rrFullDomain:", rrDependentModels[i].RrFullDomain, err)
			continue
		}
		if rrId == 0 {
			belogs.Debug("saveRrDependentModelsDb(): queryByRrFullDomain, not found id by rrFullDomain:", rrDependentModels[i].RrFullDomain)
			continue
		}
		_, err = session.Exec(sql, rrId, rrDependentModels[i].RrFullDomain, childIds)
		if err != nil {
			belogs.Error("saveRrDependentModelsDb(): insert rr_dependent fail:", sql,
				" rrDependentModel:", jsonutil.MarshalJson(rrDependentModels[i]), "  childIds:", childIds, err)
			return xormdb.RollbackAndLogError(session, "insert rr_dependent fail:", err)
		}
	}

	// just insert ignore
	sql = `insert ignore into lab_dns_origin(origin,ttl,updateTime) values (?,?,?)`
	for i := range rrDependentModels {
		ttl := null.IntFrom(1000)
		origin := dnsutil.DomainTrimFirstLabel(rrDependentModels[i].RrFullDomain)
		_, err = session.Exec(sql, origin, ttl, start)
		if err != nil {
			belogs.Error("saveRrDependentModelsDb(): insert dns_origin fail:", sql,
				"  rrDependentModel:", jsonutil.MarshalJson(rrDependentModels[i]), "   origin:", origin, err)
			return xormdb.RollbackAndLogError(session, "insert dns_origin fail:", err)
		}
	}

	err = xormdb.CommitSession(session)
	if err != nil {
		return xormdb.RollbackAndLogError(session, "saveRrDependentModelsDb(): CommitSession fail", err)
	}
	belogs.Debug("saveRrDependentModelsDb(): len(rrDependentModels):", len(rrDependentModels), "  time(s):", time.Since(start))
	return nil

}

func queryByRrFullDomain(rrFullDomain string) (id uint64, err error) {
	belogs.Debug("queryByRrFullDomain(): rrFullDomain:", rrFullDomain)
	resultIds := make([]uint64, 0)
	sql := `select r.id from lab_dns_rr r where r.rrFullDomain=?`
	err = xormdb.XormEngine.SQL(sql, rrFullDomain).Find(&resultIds)
	if err != nil {
		belogs.Error("queryByRrFullDomain(): Find resultIds fail, rrFullDomain:", rrFullDomain, err)
		return 0, err
	}
	belogs.Debug("queryByRrFullDomain(): found resultIds:", resultIds, " rrFullDomain:", rrFullDomain)
	if len(resultIds) == 0 {
		belogs.Debug("queryByRrFullDomain(): not found rrFullDomain:", rrFullDomain)
		return 0, nil
	}
	return resultIds[0], nil
}
