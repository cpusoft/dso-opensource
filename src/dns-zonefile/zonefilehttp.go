package zonefile

import (
	"os"
	"time"

	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/conf"
	"github.com/cpusoft/goutil/ginserver"
	"github.com/cpusoft/goutil/jsonutil"
	"github.com/gin-gonic/gin"
)

func ImportZoneFile(c *gin.Context) {
	// saved dir
	zoneFilePath := conf.String("zoneFile::zoneFilePath")
	timePath := time.Now().Local().Format("20060102T150405.999999999")
	zoneFilePath = zoneFilePath + string(os.PathSeparator) + timePath + string(os.PathSeparator)
	os.MkdirAll(zoneFilePath, os.ModePerm)
	belogs.Debug("ImportZoneFile():MkdirAll zoneFilePath:", zoneFilePath)

	// saved files
	zoneFile, err := ginserver.ReceiveFile(c, zoneFilePath)
	if err != nil {
		belogs.Error("ImportZoneFile():ReceiveFile: err:", zoneFilePath, err)
		ginserver.ResponseFail(c, err, "")
		return
	}
	belogs.Debug("ImportZoneFile(): zoneFilePath,zoneFile:", zoneFilePath, zoneFile)

	err = importZoneFile(zoneFile)
	if err != nil {
		belogs.Error("ImportZoneFile(): importZoneFile fail:", zoneFile, err)
		ginserver.ResponseFail(c, err, "")
		return
	}
	belogs.Info("ImportZoneFile():importZoneFile ok,  receiveFile:", zoneFile)
	ginserver.ResponseOk(c, nil)
}

func ExportZoneFile(c *gin.Context) {

	exportOrigin := ExportOrigin{}
	err := c.ShouldBindJSON(&exportOrigin)
	if err != nil {
		belogs.Error("ExportZoneFile(): ShouldBindJSON:", err)
		ginserver.ResponseFail(c, err, "")
		return
	}
	belogs.Debug("ExportZoneFile(): exportOrigin:", jsonutil.MarshalJson(exportOrigin))

	originModel, err := exportZoneFile(exportOrigin)
	if err != nil {
		belogs.Error("ExportZoneFile(): exportZoneFile fail:", err)
		ginserver.ResponseFail(c, err, "")
		return
	}
	str := originModel.String()
	belogs.Info("ExportZoneFile():exportZoneFile ok,  originModel:", jsonutil.MarshalJson(originModel), str)
	ginserver.String(c, str)
}

/*
func Query(c *gin.Context) {
	belogs.Info("Query():")
	queryRr := zonefileutil.ResourceRecord{}
	err := c.ShouldBindJSON(&queryRr)
	if err != nil {
		belogs.Error("Query(): ShouldBindJSON fail:", err)
		ginserver.ResponseFail(c, err, "")
		return
	}
	belogs.Info("Query(): query :", jsonutil.MarshalJson(queryRr))

	// query get result rr
	resultRrs, err := query(zoneFileServer, &queryRr)
	if err != nil {
		belogs.Error("Query(): query, fail:", queryRr, err)
		ginserver.ResponseFail(c, err, "")
		return
	}
	belogs.Info("Query(): queryRr :", jsonutil.MarshalJson(queryRr),
		"  result resultRrs:", jsonutil.MarshalJson(resultRrs))
	ginserver.ResponseOk(c, resultRrs)
}

func QueryAndPush(c *gin.Context) {
	belogs.Info("QueryAndPush():")
	queryRr := zonefileutil.ResourceRecord{}
	err := c.ShouldBindJSON(&queryRr)
	if err != nil {
		belogs.Error("QueryAndPush(): ShouldBindJSON fail:", err)
		ginserver.ResponseFail(c, err, "")
		return
	}
	belogs.Info("QueryAndPush(): query :", jsonutil.MarshalJson(queryRr))

	// common : save->query->push results
	err = saveAndQueryAndPush(zoneFileServer, &queryRr, nil, false)
	if err != nil {
		belogs.Error("saveAndQueryAndPush(): saveAndQueryAndPush fail:", jsonutil.MarshalJson(queryRr), err)
		ginserver.ResponseFail(c, err, "")
		return
	}
	ginserver.ResponseOk(c, nil)
}

// addRr: {"rrDomain":"a.mydomain.com.","rrName":"a","rrType":"A","rrClass":"IN","rrValues":["1.1.1.1"]}
func AddAndPush(c *gin.Context) {
	belogs.Info("AddAndPush():")
	addRr := zonefileutil.ResourceRecord{}
	err := c.ShouldBindJSON(&addRr)
	if err != nil {
		belogs.Error("AddAndPush(): ShouldBindJSON:", err)
		ginserver.ResponseFail(c, err, "")
		return
	}

	err = addResourceRecord(zoneFileServer, &addRr)
	if err != nil {
		belogs.Error("AddAndPush(): addResourceRecord:", err)
		ginserver.ResponseFail(c, err, "")
		return
	}
	belogs.Info("AddAndPush(): addRr will set as queryRt :", jsonutil.MarshalJson(addRr))

	// common : save->query->push results
	err = saveAndQueryAndPush(zoneFileServer, &addRr, nil, true)
	if err != nil {
		belogs.Error("AddAndPush(): saveAndQueryAndPush fail:", err)
		ginserver.ResponseFail(c, err, "")
		return
	}
	ginserver.ResponseOk(c, nil)

}

func UpdateAndPush(c *gin.Context) {
	belogs.Info("UpdateAndPush():")
	updateRr := UpdateResourceRecord{}
	err := c.ShouldBindJSON(&updateRr)
	if err != nil {
		belogs.Error("UpdateAndPush(): ShouldBindJSON:", err)
		ginserver.ResponseFail(c, err, "")
		return
	}
	oldRr := updateRr.OldResourceRecord
	newRr := updateRr.NewResourceRecord
	err = updateResourceRecord(zoneFileServer, &oldRr, &newRr)
	if err != nil {
		belogs.Error("UpdateAndPush(): updateResourceRecord:", jsonutil.MarshalJson(updateRr), err)
		ginserver.ResponseFail(c, err, "")
		return
	}
	belogs.Info("UpdateAndPush(): updateResourceRecord  updateRr :", jsonutil.MarshalJson(updateRr))

	// common : save->query->push results
	err = saveAndQueryAndPush(zoneFileServer, &newRr, &oldRr, true)
	if err != nil {
		belogs.Error("UpdateAndPush(): saveAndQueryAndPush fail:", err)
		ginserver.ResponseFail(c, err, "")
		return
	}
	ginserver.ResponseOk(c, nil)

}

// del, not unsubscribe
func DelAndPush(c *gin.Context) {
	belogs.Info("DelAndPush():")
	delRr := zonefileutil.ResourceRecord{}
	err := c.ShouldBindJSON(&delRr)
	if err != nil {
		belogs.Error("DelAndPush(): ShouldBindJSON:", err)
		ginserver.ResponseFail(c, err, "")
		return
	}

	newDelRr, err := delResourceRecord(zoneFileServer, &delRr)
	if err != nil {
		belogs.Error("DelAndPush(): delResourceRecord:", err)
		ginserver.ResponseFail(c, err, "")
		return
	}
	belogs.Info("DelAndPush(): delResourceRecord newDelRr :", jsonutil.MarshalJson(newDelRr))

	// common : save->query->push results
	err = saveAndQueryAndPush(zoneFileServer, nil, newDelRr, true)
	if err != nil {
		belogs.Error("AddAndPush(): saveAndQueryAndPush fail:", err)
		ginserver.ResponseFail(c, err, "")
		return
	}
	ginserver.ResponseOk(c, nil)

}
*/
