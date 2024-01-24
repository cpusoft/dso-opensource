package push

import (
	"testing"

	pushmodel "dns-model/push"
	"github.com/cpusoft/goutil/belogs"
	_ "github.com/cpusoft/goutil/conf"
	"github.com/cpusoft/goutil/jsonutil"
	_ "github.com/cpusoft/goutil/logs"
	"github.com/cpusoft/goutil/xormdb"
)

func TestSubscribe(t *testing.T) {
	err := xormdb.InitMySql()
	if err != nil {
		belogs.Error("startWebServer(): start InitMySql failed:", err)
		return
	}
	defer xormdb.XormEngine.Close()
	connKey := "1.1.1.1:8080-1.1.1.1:8081"
	pushRrModel := &pushmodel.PushRrModel{
		ConnKey:      connKey,
		RrFullDomain: "dns1.zdns.cn",
		RrType:       "A",
		RrClass:      "IN",
		RrData:       "1.1.1.1",
	}
	belogs.Debug(jsonutil.MarshalJson(pushRrModel))
	rrs, err := subscribe(pushRrModel)
	belogs.Debug(rrs, err)

	pushRrModel2 := &pushmodel.PushRrModel{
		ConnKey:      connKey,
		RrFullDomain: "dns2.zdns.cn",
		RrType:       "A",
		RrClass:      "IN",
		RrData:       "2.2.2.2",
	}
	belogs.Debug(jsonutil.MarshalJson(pushRrModel2))
	rrs, err = subscribe(pushRrModel2)
	belogs.Debug(rrs, err)

	queryRr := &pushmodel.PushRrModel{
		ConnKey:      connKey,
		RrFullDomain: "dns2.zdns.cn",
		RrType:       "A",
		RrClass:      "IN",
	}
	belogs.Debug(jsonutil.MarshalJson(pushRrModel2))
	queryRrs := make([]*pushmodel.PushRrModel, 0)
	queryRrs = append(queryRrs, queryRr)

	results, err := queryRrModelsShouldPush(queryRrs)
	belogs.Debug(results, err)

	/*
		err = unsubscribe(pushRrModel)
		belogs.Debug(rrs, err)

		delConn(connKey)
		belogs.Debug(rrs, err)
	*/
}
