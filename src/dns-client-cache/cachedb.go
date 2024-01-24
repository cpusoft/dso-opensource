package clientcache

import (
	"time"

	"dns-model/rr"
	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/dnsutil"
	"github.com/cpusoft/goutil/jsonutil"
	"github.com/cpusoft/goutil/xormdb"
)

var (
	DNS_DSO_MESSAGE_ID_STATE_APPLY   = "apply"
	DNS_DSO_MESSAGE_ID_STATE_CONFIRM = "confirm"
)

func initDb() error {
	err := xormdb.InitSqlite()
	if err != nil {
		belogs.Error("initDb(): InitSqlite fail:", err)
		return err
	}
	// if table is not exist, wil create table
	err = createTableDb()
	if err != nil {
		belogs.Error("initDb(): createTableDb fail:", err)
		return err
	}
	return nil
}
func createTableDb() error {
	belogs.Debug("createTableDb(): start")
	found, err := foundTable("lab_dns_rr")
	if err != nil {
		belogs.Error("createTableDb(): foundTable fail:", err)
		return err
	}
	if !found {
		err = initTableDb()
		if err != nil {
			belogs.Error("createTableDb(): initTableDb fail:", err)
			return err
		}
		belogs.Info("createTableDb(): not foundTable and initTableDb ok")
	}
	belogs.Info("createTableDb(): foundTable and initTableDb ok")
	return nil
}

func foundTable(table string) (bool, error) {
	var myCount int
	sql := `SELECT count(*) as mycount FROM sqlite_master WHERE type="table" AND name = ? `
	found, err := xormdb.XormEngine.SQL(sql, table).Get(&myCount)
	if err != nil {
		belogs.Error("foundTable(): table:", table, err)
		return false, err
	}
	belogs.Debug("foundTable(): found:", found, "   myCount:", myCount)
	if found && myCount >= 1 {
		return true, nil
	} else {
		return false, nil
	}
}

func initTableDb() error {
	session, err := xormdb.NewSession()
	if err != nil {
		belogs.Error("initTableDb(): NewSession fail: err:", err)
		return err
	}
	defer session.Close()
	var initSqls []string = []string{
		`drop table if exists lab_dns_rr`,
		`drop table if exists lab_dns_message`,

		`CREATE TABLE IF NOT EXISTS lab_dns_rr (
			id integer primary key autoincrement,
			rrFullDomain varchar(512) not null ,
			rrType varchar(256) not null ,
			rrClass varchar(256),
			rrTtl int(10),
			rrData varchar(1024),
			updateTime text not null
		)`,
		`CREATE INDEX IF NOT EXISTS rrFullDomain ON lab_dns_rr(rrFullDomain)`,
		`CREATE INDEX IF NOT EXISTS rrFullDomainAndRrType ON lab_dns_rr(rrFullDomain,rrType)`,

		`CREATE TABLE IF NOT EXISTS lab_dns_message (
			id integer primary key autoincrement,
			state varchar(16) not null,
			applyTime text not null,
			confirmTime text,
			opCode varchar(16),
			dsoType varchar(16),
			subscirbeRrFullDomain varchar(512),
			subscirbeRrType varchar(256),
			unsubscirbeTime text
		)`,
	}
	for _, sql := range initSqls {
		_, err = session.Exec(sql)
		if err != nil {
			belogs.Error("initTableDb(): initSqls fail:", sql, err)
			return xormdb.RollbackAndLogError(session, "initSqls fail:"+sql, err)
		}
	}
	if err = session.Commit(); err != nil {
		return xormdb.RollbackAndLogError(session, "initTableDb():commit fail", err)
	}
	return nil
}

func resetTableDb() error {
	session, err := xormdb.NewSession()
	if err != nil {
		belogs.Error("resetTableDb(): NewSession fail: err:", err)
		return err
	}
	defer session.Close()
	var resetSqls []string = []string{
		`delete from lab_dns_rr`,
		`update sqlite_sequence SET seq = 0 where name ='lab_dns_rr'`,
		`delete from lab_dns_message`,
		`update sqlite_sequence SET seq = 0 where name ='lab_dns_message'`,
	}
	for _, sql := range resetSqls {
		_, err = session.Exec(sql)
		if err != nil {
			belogs.Error("resetTableDb(): resetSqls fail:", sql, err)
			return xormdb.RollbackAndLogError(session, "resetSqls fail:"+sql, err)
		}
	}
	if err = session.Commit(); err != nil {
		return xormdb.RollbackAndLogError(session, "resetTableDb():commit fail", err)
	}
	return nil
}

func updateRrModelDb(rrModel *rr.RrModel, justDel bool) error {
	belogs.Debug("updateRrModelDb(): rrModel:", jsonutil.MarshalJson(rrModel))
	start := time.Now()
	session, err := xormdb.NewSession()
	if err != nil {
		belogs.Error("updateRrModelDb(): NewSession fail: err:", err)
		return err
	}
	defer session.Close()

	var sql string
	if len(rrModel.RrType) > 0 {
		sql = `delete from lab_dns_rr where rrFullDomain=? and rrType=?`
		_, err = session.Exec(sql, rrModel.RrFullDomain, rrModel.RrType)
	} else {
		sql = `delete from lab_dns_rr where rrFullDomain=?`
		_, err = session.Exec(sql, rrModel.RrFullDomain)
	}
	if err != nil {
		belogs.Error("updateRrModelDb(): delete Exec fail: sql:", sql, "  rrModel:", jsonutil.MarshalJson(rrModel), err)
		return xormdb.RollbackAndLogError(session, "updateRrModelDb():exec fail:"+sql, err)
	}
	if !justDel {
		sql = `insert into lab_dns_rr(	
			rrFullDomain, rrType, rrClass, 
			rrTtl, rrData, updateTime) values (
			?,?,?,
			?,?,?)
			`
		_, err = session.Exec(sql,
			rrModel.RrFullDomain, rrModel.RrType, rrModel.RrClass,
			rrModel.RrTtl, rrModel.RrData, start.Format("2006-01-02 15:04:05"))
		if err != nil {
			belogs.Error("updateRrModelDb():insert Exec fail: sql:", sql, "  rrModel:", jsonutil.MarshalJson(rrModel), err)
			return xormdb.RollbackAndLogError(session, "updateRrModelDb():exec fail:"+sql, err)
		}
	}

	if err = session.Commit(); err != nil {
		return xormdb.RollbackAndLogError(session, "updateRrModelDb():commit fail", err)
	}
	belogs.Debug("updateRrModelDb(): rrModel:", jsonutil.MarshalJson(rrModel),
		"   time(s):", time.Since(start))
	return nil
}

func queryRrModelsDb(rrModel *rr.RrModel) ([]*rr.RrModel, error) {
	belogs.Debug("queryRrModelsDb(): rrModel:", jsonutil.MarshalJson(rrModel))
	start := time.Now()
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

	belogs.Debug("queryRrModelsDb(): len(rrModels):", len(rrModels), "  rrModels:", jsonutil.MarshalJson(rrModels),
		"   time(s):", time.Since(start))
	return rrModels, nil
}
func queryAllRrModelsDb() ([]*rr.RrModel, error) {
	belogs.Debug("queryAllRrModelsDb():")
	start := time.Now()
	rrModels := make([]*rr.RrModel, 0)
	err := xormdb.XormEngine.Table("lab_dns_rr").Cols("rrFullDomain, rrType, rrClass, rrTtl, rrData").OrderBy("id").Find(&rrModels)
	if err != nil {
		belogs.Error("queryAllRrModelsDb(): select lab_dns_rr fail:", jsonutil.MarshalJson(rrModels), err)
		return nil, err
	}

	belogs.Debug("queryAllRrModelsDb(): len(rrModels):", len(rrModels), "  rrModels:", jsonutil.MarshalJson(rrModels),
		"   time(s):", time.Since(start))
	return rrModels, nil
}
func clearAllRrModelsDb() error {
	session, err := xormdb.NewSession()
	if err != nil {
		belogs.Error("clearAllRrModelsDb(): NewSession fail: err:", err)
		return err
	}
	defer session.Close()
	start := time.Now()
	sql := `delete from lab_dns_rr`
	_, err = session.Exec(sql)
	if err != nil {
		belogs.Error("clearAllRrModelsDb():delete Exec fail: sql:", sql, err)
		return xormdb.RollbackAndLogError(session, "clearAllRrModelsDb():delete exec fail:"+sql, err)
	}
	sql = `update sqlite_sequence set seq=0 where name='lab_dns_rr'`
	_, err = session.Exec(sql)
	if err != nil {
		belogs.Error("clearAllRrModelsDb():set sqlite_sequence seq=0 Exec fail: sql:", sql, err)
		return xormdb.RollbackAndLogError(session, "clearAllRrModelsDb():sqlite_sequence exec fail:"+sql, err)
	}

	if err = session.Commit(); err != nil {
		return xormdb.RollbackAndLogError(session, "clearAllRrModelsDb():commit fail", err)
	}
	belogs.Debug("clearAllRrModelsDb(): time(s):", time.Since(start))
	return nil
}
func getNewMessageIdDb(opCode uint8) (uint16, error) {
	opCodeStr, _ := dnsutil.DnsIntOpCodes[opCode]
	belogs.Debug("getNewMessageIdDb():opCode:", opCode, "  opCodeStr:", opCodeStr)

	var messageId uint16
	session, err := xormdb.NewSession()
	if err != nil {
		belogs.Error("getNewMessageIdDb(): NewSession fail: err:", err)
		return 0, err
	}
	defer session.Close()

	sql := `insert into lab_dns_message(state,opCode,applyTime) values(?,?,?) `
	_, err = session.Exec(sql, DNS_DSO_MESSAGE_ID_STATE_APPLY, opCodeStr, time.Now().Format("2006-01-02 15:04:05"))
	if err != nil {
		belogs.Error("getNewMessageIdDb():insert lab_dns_message fail: sql:", sql, err)
		return 0, xormdb.RollbackAndLogError(session, "getNewMessageIdDb():insert lab_dns_message, fail:"+sql, err)
	}

	sql = `select last_insert_rowid() as messageId from lab_dns_message`
	_, err = session.SQL(sql).Get(&messageId)
	if err != nil {
		belogs.Error("getNewMessageIdDb():get messageId fail: sql:", sql, err)
		return 0, xormdb.RollbackAndLogError(session, "getNewMessageIdDb():get messageId fail:"+sql, err)
	}

	if err = session.Commit(); err != nil {
		return 0, xormdb.RollbackAndLogError(session, "getNewMessageIdDb():commit fail", err)
	}
	belogs.Info("getNewMessageIdDb(): messageId:", messageId)
	return messageId, nil
}

func updateDsoMessageDsoTypeAndRrModelDb(messageId uint16, dsoType uint8, rrModel *rr.RrModel) error {
	dsoTypeStr, _ := dnsutil.DsoIntTypes[dsoType]
	belogs.Debug("updateDsoMessageDsoTypeAndRrModelDb():messageId:", messageId, " dsoType:", dsoType, "  dsoTypeStr:", dsoTypeStr,
		"   rrModel:", jsonutil.MarshalJson(rrModel))

	session, err := xormdb.NewSession()
	if err != nil {
		belogs.Error("updateDsoMessageDsoTypeAndRrModelDb(): NewSession fail: err:", err)
		return err
	}
	defer session.Close()
	if rrModel != nil {
		sql := `update lab_dns_message set dsoType=?, subscirbeRrFullDomain=?, subscirbeRrType =? where id= ? `
		_, err = session.Exec(sql, dsoTypeStr, rrModel.RrFullDomain, rrModel.RrType, messageId)
		if err != nil {
			belogs.Error("updateDsoMessageDsoTypeAndRrModelDb():update lab_dns_message have rrModel, fail: sql:", sql, err)
			return xormdb.RollbackAndLogError(session, "updateDsoMessageDsoTypeAndRrModelDb():update lab_dns_message have rrModel, fail:"+sql, err)
		}
	} else {
		sql := `update lab_dns_message set dsoType=? where id= ? `
		_, err = session.Exec(sql, dsoTypeStr, messageId)
		if err != nil {
			belogs.Error("updateDsoMessageDsoTypeAndRrModelDb():update lab_dns_message no rrModel, fail: sql:", sql, err)
			return xormdb.RollbackAndLogError(session, "updateDsoMessageDsoTypeAndRrModelDb():update lab_dns_message no rrModel, fail:"+sql, err)
		}
	}
	if err = session.Commit(); err != nil {
		return xormdb.RollbackAndLogError(session, "updateDsoMessageDsoTypeAndRrModelDb():commit fail", err)
	}
	belogs.Info("updateDsoMessageDsoTypeAndRrModelDb(): messageId:", messageId)
	return nil
}
func updateDsoMessageUnsubscribeTimeDb(messageId uint16) error {
	session, err := xormdb.NewSession()
	if err != nil {
		belogs.Error("updateDsoMessageUnsubscribeTimeDb(): NewSession fail: err:", err)
		return err
	}
	defer session.Close()
	sql := `update lab_dns_message set unsubscirbeTime = ? where id = ?`
	if err = session.Commit(); err != nil {
		return xormdb.RollbackAndLogError(session, "updateDsoMessageUnsubscribeTimeDb():commit fail", err)
	}
	_, err = session.Exec(sql, time.Now().Format("2006-01-02 15:04:05"), messageId)
	if err != nil {
		belogs.Error("updateDsoMessageUnsubscribeTimeDb():update lab_dns_message unsubscirbeTime fail: sql:", sql, err)
		return xormdb.RollbackAndLogError(session, "updateDsoMessageUnsubscribeTimeDb():update lab_dns_message unsubscirbeTime fail, fail:"+sql, err)
	}
	belogs.Info("updateDsoMessageUnsubscribeTimeDb(): messageId:", messageId)
	return nil
}
func queryDsoMessageIdByRrModelDb(rrModel *rr.RrModel) (bool, uint16, error) {
	var messageId uint16
	sql := `select id from lab_dns_message where subscirbeRrFullDomain=? and subscirbeRrType =?
	   and opCode='dso' and dsoType='subscribe' and unsubscirbeTime is null order by id desc limit 1`
	found, err := xormdb.XormEngine.SQL(sql, rrModel.RrFullDomain, rrModel.RrType).Get(&messageId)
	if err != nil {
		belogs.Error("queryDsoMessageIdByRrModelDb(): select lab_dns_message fail, rrModel:", jsonutil.MarshalJson(rrModel), err)
		return false, 0, err
	} else if !found {
		belogs.Debug("queryDsoMessageIdByRrModelDb(): select lab_dns_message not found, rrModel:", jsonutil.MarshalJson(rrModel), err)
		return false, 0, nil
	}
	belogs.Debug("queryDsoMessageIdByRrModelDb(): messageId:", messageId)
	return true, messageId, nil
}

func confirmMessageIdDb(messageId uint16) (bool, error) {
	belogs.Debug("confirmMessageIdDb():messageId:", messageId)

	session, err := xormdb.NewSession()
	if err != nil {
		belogs.Error("confirmMessageIdDb(): NewSession fail: err:", err)
		return false, err
	}
	defer session.Close()

	sql := `update lab_dns_message set state = ? , confirmTime = ? where id = ? `
	affected, err := session.Exec(sql, DNS_DSO_MESSAGE_ID_STATE_CONFIRM, time.Now().Format("2006-01-02 15:04:05"), messageId)
	if err != nil {
		belogs.Error("confirmMessageIdDb():insert lab_dns_message fail: sql:", sql, err)
		return false, xormdb.RollbackAndLogError(session, "confirmMessageIdDb():insert lab_dns_message, fail:"+sql, err)
	}
	if err = session.Commit(); err != nil {
		return false, xormdb.RollbackAndLogError(session, "confirmMessageIdDb():commit fail", err)
	}
	rowsNum, _ := affected.RowsAffected()
	belogs.Info("confirmMessageIdDb(): rowsNum:", rowsNum)
	return rowsNum > 0, nil

}
