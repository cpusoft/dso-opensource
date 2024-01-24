package sys

import (
	"errors"
	"time"

	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/jsonutil"
	"github.com/cpusoft/goutil/xormdb"
	"xorm.io/xorm"
)

var initSqls []string = []string{
	`drop table if exists lab_dns_origin`,
	`drop table if exists lab_dns_rr`,

	`
#################################
## dns origin
#################################	
CREATE TABLE lab_dns_origin (
	id int(10) unsigned not null primary key auto_increment,
	origin varchar(256) NOT NULL,
	ttl int(10)  unsigned not null ,
	updateTime datetime NOT NULL,
	unique origin (origin)	
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin COMMENT='origin table'
`,
	`
#################################
## dns rr
#################################		
CREATE TABLE lab_dns_rr (
	id int(10) unsigned not null primary key auto_increment,
	originId int(10) unsigned not null, 
	rrName varchar(512) not null ,
	rrFullDomain varchar(512) not null ,
	rrType varchar(256) not null ,
	rrClass varchar(256),
	rrTtl int(10) unsigned,
	rrData varchar(1024),
	updateTime datetime NOT NULL,
	key rrFullDomain (rrFullDomain),
	key rrFullDomainAndRrType (rrFullDomain,rrType),
	FOREIGN key (originId) REFERENCES lab_dns_origin(id)	
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin COMMENT='rr table'
`,
}
var resetSqls []string = []string{
	`truncate  table  lab_dns_origin`,
	`truncate  table  lab_dns_rr`,

	`optimize  table  lab_dns_origin`,
	`optimize  table  lab_dns_rr`,
}

func initResetDb(sysStyle SysStyle) error {
	belogs.Debug("initResetDb():sysStyle:", jsonutil.MarshalJson(sysStyle))
	session, err := xormdb.NewSession()
	if err != nil {
		belogs.Error("initResetDb(): NewSession fail: err:", err)
		return err
	}
	defer session.Close()

	//truncate all table
	err = initResetImplDb(session, sysStyle)
	if err != nil {
		return xormdb.RollbackAndLogError(session, "initResetDb(): initResetImplDb fail", err)
	}

	err = xormdb.CommitSession(session)
	if err != nil {
		return xormdb.RollbackAndLogError(session, "initResetDb(): CommitSession fail", err)
	}
	return nil
}

// need to init sessionId when it is empty
func initResetImplDb(session *xorm.Session, sysStyle SysStyle) error {

	start := time.Now()
	sql := `set foreign_key_checks=0;`
	if _, err := session.Exec(sql); err != nil {
		belogs.Error("initResetImplDb(): SET foreign_key_checks=0 fail", err)
		return xormdb.RollbackAndLogError(session, "initResetImplDb():SET foreign_key_checks=0 fail: ", err)
	}
	belogs.Debug("initResetImplDb():foreign_key_checks=0;   time(s):", time.Since(start))

	// delete rtr_session
	var sqls []string
	if sysStyle.SysStyle == "init" {
		sqls = initSqls
	} else if sysStyle.SysStyle == "resetall" {
		sqls = resetSqls
	} else {
		return xormdb.RollbackAndLogError(session, "initResetImplDb():sysStyle fail: "+jsonutil.MarshalJson(sysStyle),
			errors.New("sysStyle is illegal"))
	}

	belogs.Info("initResetImplDb():will Exec len(sqls):", len(sqls))
	for _, sq := range sqls {
		now := time.Now()
		if _, err := session.Exec(sq); err != nil {
			belogs.Error("initResetImplDb():  "+sq+" fail", err)
			return xormdb.RollbackAndLogError(session, "initResetImplDb():sql fail: "+sq, err)
		}
		belogs.Info("initResetImplDb(): sq:", sq, ", sql time(s):", time.Since(now))
	}
	sql = `set foreign_key_checks=1;`
	if _, err := session.Exec(sql); err != nil {
		belogs.Error("initResetImplDb(): SET foreign_key_checks=1 fail", err)
		return xormdb.RollbackAndLogError(session, "initResetImplDb():SET foreign_key_checks=1 fail", err)
	}
	belogs.Info("initResetImplDb(): len(sqls):", len(sqls), ",  time(s):", time.Since(start))

	if err := session.Commit(); err != nil {
		return xormdb.RollbackAndLogError(session, "initResetImplDb():commit fail", err)
	}
	return nil
}
