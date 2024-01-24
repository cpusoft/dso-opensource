package push

import (
	pushmodel "dns-model/push"
	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/ginserver"
	"github.com/cpusoft/goutil/jsonutil"
	"github.com/gin-gonic/gin"
)

func Subscribe(c *gin.Context) {
	belogs.Info("Subscribe():")
	pushRrModel := pushmodel.PushRrModel{}
	err := c.ShouldBindJSON(&pushRrModel)
	if err != nil {
		belogs.Error("Subscribe(): ShouldBindJSON:", err)
		ginserver.ResponseFail(c, err, "")
		return
	}
	belogs.Debug("Subscribe(): pushRrModel:", jsonutil.MarshalJson(pushRrModel))

	pushResultRrModels, err := subscribe(&pushRrModel)
	if err != nil {
		belogs.Error("Subscribe(): subscribe:", err)
		ginserver.ResponseFail(c, err, "")
		return
	}
	belogs.Debug("Subscribe(): subscribe, pushResultRrModels:", jsonutil.MarshalJson(pushResultRrModels))

	ginserver.ResponseOk(c, pushResultRrModels)
}

func Unsubscribe(c *gin.Context) {
	belogs.Info("Unsubscribe():")
	unpushRrModel := pushmodel.UnpushRrModel{}
	err := c.ShouldBindJSON(&unpushRrModel)
	if err != nil {
		belogs.Error("Unsubscribe(): ShouldBindJSON:", err)
		ginserver.ResponseFail(c, err, "")
		return
	}
	belogs.Debug("Unsubscribe(): unpushRrModel:", jsonutil.MarshalJson(unpushRrModel))

	err = unsubscribe(&unpushRrModel)
	if err != nil {
		belogs.Error("Unsubscribe(): unsubscribe:", err)
		ginserver.ResponseFail(c, err, "")
		return
	}
	ginserver.ResponseOk(c, nil)
}

func DelConn(c *gin.Context) {
	belogs.Info("DelConn():")
	pushRrModel := pushmodel.PushRrModel{}
	err := c.ShouldBindJSON(&pushRrModel)
	if err != nil {
		belogs.Error("DelConn(): ShouldBindJSON:", err)
		ginserver.ResponseFail(c, err, "")
		return
	}
	belogs.Debug("DelConn(): pushRrModel:", jsonutil.MarshalJson(pushRrModel))
	connKey := pushRrModel.ConnKey
	delConn(connKey)
	ginserver.ResponseOk(c, nil)
}

func QueryRrModelsShouldPush(c *gin.Context) {
	belogs.Info("QueryRrModelsShouldPush():")
	pushRrModels := make([]*pushmodel.PushRrModel, 0)
	err := c.ShouldBindJSON(&pushRrModels)
	if err != nil {
		belogs.Error("QueryRrModelsShouldPush(): ShouldBindJSON:", err)
		ginserver.ResponseFail(c, err, "")
		return
	}
	belogs.Debug("QueryRrModelsShouldPush(): pushRrModels:", jsonutil.MarshalJson(pushRrModels))

	pushResultRrModels, err := queryRrModelsShouldPush(pushRrModels)
	if err != nil {
		belogs.Error("QueryRrModelsShouldPush(): queryRrModelsShouldPush fail, rrModels:", jsonutil.MarshalJson(pushRrModels), err)
		ginserver.ResponseFail(c, err, "")
		return
	}
	belogs.Info("QueryRrModelsShouldPush(): pushResultRrModels:", jsonutil.MarshalJson(pushResultRrModels))
	ginserver.ResponseOk(c, pushResultRrModels)
}

func ActivePushAll(c *gin.Context) {
	belogs.Info("ActivePushAll():")
	pushResultRrModels, err := activePushAll()
	if err != nil {
		belogs.Error("QueryRrModelsShouldPush(): activePushAll fail:", err)
		ginserver.ResponseFail(c, err, "")
		return
	}
	belogs.Info("ActivePushAll(): pushResultRrModels:", jsonutil.MarshalJson(pushResultRrModels))
	ginserver.ResponseOk(c, pushResultRrModels)
}

func GetAllSubscribedRrs(c *gin.Context) {
	belogs.Debug("GetAllSubscribedRrs():")
	subscribedRrs := getAllSubscribedRrs()
	belogs.Info("GetAllSubscribedRrs(): subscribedRrs:", jsonutil.MarshalJson(subscribedRrs))
	ginserver.ResponseOk(c, subscribedRrs)
}

/*
func Reconfirm(c *gin.Context) {
	belogs.Info("Unsubscribe():")
	connKey, resourceRecord, err := getConnKeyAndResourceRecord(c)
	if err != nil {
		belogs.Error("Unsubscribe(): ShouldBindJSON:", err)
		ginserver.ResponseFail(c, err, "")
		return
	}
	belogs.Debug("Reconfirm(): connKey:", connKey, " resourceRecord:", jsonutil.MarshalJson(resourceRecord))
	err = reconfirmResourceRecordInConn(pushServer, connKey, resourceRecord)
	if err != nil {
		belogs.Error("Unsubscribe(): unsubscribeResourceRecordInConn:", err)
		ginserver.ResponseFail(c, err, "")
		return
	}
	ginserver.ResponseOk(c, nil)
}




func getConnKeyAndResourceRecord(c *gin.Context) (connKey string,
	resourceRecord *dnsmodel.RrModel, err error) {
	connResourceRecordModel := PushRrModel{}
	err = c.ShouldBindJSON(&connResourceRecordModel)
	if err != nil {
		belogs.Error("getConnKeyAndResourceRecord(): ShouldBindJSON:", err)
		return "", nil, err
	}
	belogs.Debug("getConnKeyAndResourceRecord(): connResourceRecordModel:", jsonutil.MarshalJson(connResourceRecordModel))
	connKey, resourceRecord, err = convertToResourceRecord(connResourceRecordModel)
	if err != nil {
		belogs.Error("getConnKeyAndResourceRecord(): convertToResourceRecord:", err)
		return "", nil, err
	}
	belogs.Debug("getConnKeyAndResourceRecord(): connKey:", connKey, "   resourceRecord:", jsonutil.MarshalJson(resourceRecord))

	return connKey, resourceRecord, nil
}
*/
