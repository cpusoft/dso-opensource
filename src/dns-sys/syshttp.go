package sys

import (
	"errors"

	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/ginserver"
	"github.com/cpusoft/goutil/jsonutil"
	"github.com/gin-gonic/gin"
)

//
func InitReset(c *gin.Context) {
	belogs.Debug("InitReset()")
	sysStyle := SysStyle{}
	err := c.ShouldBindJSON(&sysStyle)
	if err != nil {
		belogs.Error("InitReset(): ShouldBindJSON:", err)
		ginserver.ResponseFail(c, err, "")
		return
	}
	belogs.Info("InitReset():get sysStyle:", jsonutil.MarshalJson(sysStyle))
	if sysStyle.SysStyle != "init" && sysStyle.SysStyle != "resetall" {
		belogs.Error("InitReset(): SysStyle should be init or fullsync or resetall, it is ", sysStyle.SysStyle)
		ginserver.ResponseFail(c, errors.New("SysStyle should be init or fullsync or resetall"), "")
		return
	}
	belogs.Debug("InitReset(): sysStyle:", sysStyle)

	err = initReset(sysStyle)
	if err != nil {
		belogs.Error("InitReset(): initReset fail,err: ", err)
		ginserver.ResponseFail(c, err, "")
		return
	}
	ginserver.ResponseOk(c, nil)

}
