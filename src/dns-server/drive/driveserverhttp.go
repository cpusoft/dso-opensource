package drive

import (
	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/ginserver"
	"github.com/cpusoft/goutil/jsonutil"
	"github.com/gin-gonic/gin"
)

// start to
func ActivePushAll(c *gin.Context) {

	go activePushAll()
	ginserver.ResponseOk(c, nil)
}

func QueryServerDnsRrs(c *gin.Context) {
	serverDnsRrModel := ServerDnsRrModel{}
	err := c.ShouldBindJSON(&serverDnsRrModel)
	if err != nil {
		belogs.Error("QueryServerDnsRrs(): ShouldBindJSON:", err)
		ginserver.ResponseFail(c, err, "")
		return
	}
	belogs.Info("QueryServerDnsRrs(): serverDnsRrModel:", jsonutil.MarshalJson(serverDnsRrModel))

	resultDnsRrs, err := queryServerDnsRrs(&serverDnsRrModel)
	if err != nil {
		belogs.Error("QueryServerDnsRrs(): queryServerDnsRrs fail:", err)
		ginserver.ResponseFail(c, err, "")
		return
	}

	ginserver.ResponseOk(c, resultDnsRrs)
}
func QueryServerAllDnsRrs(c *gin.Context) {
	belogs.Info("QueryServerAllDnsRrs():")
	resultDnsRrs, err := queryServerAllDnsRrs()
	if err != nil {
		belogs.Error("QueryServerAllDnsRrs(): queryServerAllDnsRrs fail:", err)
		ginserver.ResponseFail(c, err, "")
		return
	}

	ginserver.ResponseOk(c, resultDnsRrs)

}

func QueryRpkiRepos(c *gin.Context) {
	belogs.Info("QueryRpkiRepos():")
	go queryRpkiRepos()
	ginserver.ResponseOk(c, nil)
}
