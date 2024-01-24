package zonefile

import (
	"errors"
	"strings"
	"time"

	"dns-model/rr"
	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/jsonutil"
	"github.com/cpusoft/goutil/xormdb"
)

func saveOriginModelDb(originModel *rr.OriginModel) (err error) {

	belogs.Debug("saveOriginModelDb(): originModel:", jsonutil.MarshalJson(originModel))
	start := time.Now()

	session, err := xormdb.NewSession()
	if err != nil {
		belogs.Error("saveOriginModelDb(): NewSession fail: err:", err)
		return err
	}
	defer session.Close()

	// get originId
	origin := originModel.Origin
	var id int
	// get originId
	sql := `select id from lab_dns_origin o where o.origin=?`
	found, err := session.SQL(sql, origin).Get(&id)
	if err != nil {
		belogs.Error("saveOriginModelDb(): select id fail:", jsonutil.MarshalJson(origin), err)
		return xormdb.RollbackAndLogError(session, "saveOriginModelDb(): select id fail:"+jsonutil.MarshalJson(origin), err)
	}
	belogs.Debug("saveOriginModelDb(): origin:", origin, "  id:", id)
	if found && id > 0 {
		// del dns_rr
		res, err := session.Exec("delete from lab_dns_rr where originId = ?", id)
		if err != nil {
			belogs.Error("saveOriginModelDb(): delete lab_dns_rr fail, originId:", id, err)
			return xormdb.RollbackAndLogError(session, "saveOriginModelDb(): delete lab_dns_rr fail, originId:"+jsonutil.MarshalJson(id), err)
		}
		count, _ := res.RowsAffected()
		belogs.Debug("saveOriginModelDb():delete lab_dns_rr by id:", id, "  count:", count)

		// delete dns_origin
		res, err = session.Exec("delete from lab_dns_origin where id = ?", id)
		if err != nil {
			belogs.Error("saveOriginModelDb(): delete lab_dns_origin fail, id:", id, err)
			return xormdb.RollbackAndLogError(session, "saveOriginModelDb(): delete lab_dns_origin fail, id:"+jsonutil.MarshalJson(id), err)
		}
		count, _ = res.RowsAffected()
		belogs.Debug("saveOriginModelDb():delete lab_dns_origin by id:", id, "  count:", count)
	}

	insertSql := `insert into lab_dns_origin 
		(origin, ttl, updateTime) values (?,?,?) `
	res, err := session.Exec(insertSql, originModel.Origin, originModel.Ttl, start)
	if err != nil {
		belogs.Error("saveOriginModelDb(): insert lab_dns_origin fail :", jsonutil.MarshalJson(originModel), err)
		return xormdb.RollbackAndLogError(session, "saveOriginModelDb():insert lab_dns_origin fail, originModel:"+jsonutil.MarshalJson(originModel), err)
	}
	originId, _ := res.LastInsertId()
	belogs.Debug("saveOriginModelDb(): insert , originId:", originId)

	insertSql = `insert into lab_dns_rr 
		(originId, rrName, rrFullDomain,
		 rrType, rrClass, rrTtl, 
		 rrData, updateTime) 
		values (?,?,?,   ?,?,?,   ?,?) `
	for i := range originModel.RrModels {
		session.Exec(insertSql,
			originId, xormdb.SqlNullString(originModel.RrModels[i].RrName), originModel.RrModels[i].RrFullDomain,
			xormdb.SqlNullString(originModel.RrModels[i].RrType), xormdb.SqlNullString(originModel.RrModels[i].RrClass), originModel.RrModels[i].RrTtl,
			xormdb.SqlNullString(originModel.RrModels[i].RrData), start)
		if err != nil {
			belogs.Error("saveOriginModelDb(): insert lab_dns_rr fail :", jsonutil.MarshalJson(originModel.RrModels[i]), err)
			return xormdb.RollbackAndLogError(session, "saveOriginModelDb():insert lab_dns_rr fail, originModel.RrModels[i]:"+jsonutil.MarshalJson(originModel.RrModels[i]), err)
		}
	}

	err = xormdb.CommitSession(session)
	if err != nil {
		belogs.Error("saveOriginModelDb(): CommitSession fail:", jsonutil.MarshalJson(originModel), err)
		return xormdb.RollbackAndLogError(session, "saveOriginModelDb(): CommitSession fail:"+jsonutil.MarshalJson(originModel), err)
	}
	belogs.Info("saveOriginModelDb(): insert ok, originId:", originId, "  time(s):", time.Since(start))
	return nil
}
func getOriginModelDb(exportOrigin ExportOrigin) (originModel *rr.OriginModel, err error) {
	belogs.Debug("getOriginModelDb(): exportOrigin:", jsonutil.MarshalJson(exportOrigin))

	start := time.Now()
	origin := strings.ToLower(strings.TrimSpace(exportOrigin.Origin))
	originModel = &rr.OriginModel{}
	sql := `select id, origin, ttl, updateTime from lab_dns_origin where origin = ? `
	found, err := xormdb.XormEngine.SQL(sql, origin).Get(originModel)
	if err != nil {
		belogs.Error("getOriginModelDb():select lab_dns_origin, fail, origin:", origin, err)
		return nil, err
	} else if !found {
		belogs.Error("getOriginModelDb():select lab_dns_origin, not found, origin:", origin)
		return nil, errors.New("not found this origin:" + origin)
	}
	belogs.Debug("getOriginModelDb():Get originModel:", jsonutil.MarshalJson(*originModel))

	rrModels := make([]*rr.RrModel, 0)
	sql = `select r.id, r.originId, o.origin, r.rrName, r.rrType, r.rrClass, r.rrTtl, r.rrData, r.updateTime 
		   from lab_dns_rr r ,lab_dns_origin o where o.id = r.originId and o.id = ?  order by r.id `
	err = xormdb.XormEngine.SQL(sql, originModel.Id).Find(&rrModels)
	if err != nil {
		belogs.Error("getOriginModelDb():select lab_dns_rr, fail, originId:", originModel.Id, err)
		return nil, err
	}
	belogs.Debug("getOriginModelDb():Get rrModels:", jsonutil.MarshalJson(rrModels))
	originModel.RrModels = rrModels

	belogs.Info("getOriginModelDb():originModel, rrModels:", jsonutil.MarshalJson(originModel),
		"   time(s):", time.Since(start))
	return originModel, nil
}
