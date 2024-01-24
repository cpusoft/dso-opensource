package drive

import (
	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/ginserver"
	"github.com/cpusoft/goutil/jsonutil"
	"github.com/gin-gonic/gin"
)

func CreateDnsConnect(c *gin.Context) {
	belogs.Info("CreateDnsConnect():")

	closeDnsConnect()
	go createDnsConnect()
	ginserver.ResponseOk(c, nil)
}

func CloseDnsConnect(c *gin.Context) {
	belogs.Info("CloseDnsConnect():")

	err := closeDnsConnect()
	if err != nil {
		belogs.Error("CloseDnsConnect(): closeDnsConnect fail:", err)
		ginserver.ResponseFail(c, err, "")
		return
	}

	ginserver.ResponseOk(c, nil)
}

/*
	{
		"origin": "example.com",
		"ttl": 1000,
		"rrModels": [
			{
				"id": 0,
				"originId": 0,
				"origin": "example.com.",
				"rrName": "test1",
				"rrFullDomain": "test1.example.com",
				"rrType": "A",
				"rrClass": "IN",
				"rrTtl": 1000,
				"rrData": "1.1.1.1",
				"updateTime": "0001-01-01T00:00:00Z"
			},
			{
				"id": 0,
				"originId": 0,
				"origin": "example.com.",
				"rrName": "test1",
				"rrFullDomain": "test1.example.com",
				"rrType": "TXT",
				"rrClass": "IN",
				"rrTtl": 1000,
				"rrData": "v=spf1 include:spf.mail.qq.com ip4:203.99.30.50 ~all",
				"updateTime": "0001-01-01T00:00:00Z"
			}
		]
	}
*/

/*
	{
		"origin": "example.com",
		"ttl": 10000,
		"rrModels": [
			{
				"rrFullDomain": "dns3.example.com",
				"rrType": "A",
				"rrClass": "IN",
				"rrData": "10.0.1.3"
			},
			{
				"rrFullDomain": "dns3.example.com",
				"rrType": "TXT",
				"rrClass": "IN",
				"rrTtl": 1000,
				"rrData": "v=spf1 include:spf.mail.qq.com ip4:203.99.30.50 ~all"
			}
		]
	}
*/
func AddDnsRrs(c *gin.Context) {

	belogs.Info("AddDnsRrs():")

	clientDnsRrModel := ClientDnsRrModel{}
	err := c.ShouldBindJSON(&clientDnsRrModel)
	if err != nil {
		belogs.Error("AddDnsRrs(): ShouldBindJSON:", err)
		ginserver.ResponseFail(c, err, "")
		return
	}
	belogs.Info("AddDnsRrs():clientDnsRrModel:", jsonutil.MarshalJson(clientDnsRrModel))

	err = addDnsRrs(&clientDnsRrModel)
	if err != nil {
		belogs.Error("AddDnsRrs(): addDnsRrs fail, clientDnsRrModel:", jsonutil.MarshalJson(clientDnsRrModel), err)
		ginserver.ResponseFail(c, err, "")
		return
	}
	ginserver.ResponseOk(c, nil)
}

/*
CLASS=NONE: delete  rrFullDomain= ?  and rrType = ? and rrData=?
{
	"origin": "zdns.cn",
	"ttl": 0,
	"rrModels": [
		{
			"rrFullDomain": "printer.zdns.cn",
			"rrType": "A",
			"rrClass": "NONE",
			"rrTtl": 0,
			"rrData": "1.1.1.1"
		}
	]
}

CLASS=ANY,TYPE="**" delete  rrFullDomain= ? and rrType= ?
{
	"origin": "zdns.cn",
	"ttl": 0,
	"rrModels": [
		{
			"rrFullDomain": "printer.zdns.cn",
			"rrType": "A",
			"rrClass": "ANY",
			"rrTtl": 0,
			"rrData": "1.1.1.1"
		}
	]
}

CLASS=ANY,TYPE=ANY delete  rrFullDomain= ?
{
	"origin": "zdns.cn",
	"ttl": 0,
	"rrModels": [
		{
			"rrFullDomain": "printer.zdns.cn",
			"rrType": "ANY",
			"rrClass": "ANY",
			"rrTtl": 0,
			"rrData": "1.1.1.1"
		}
	]
}

*/
func DelDnsRrs(c *gin.Context) {

	belogs.Info("DelDnsRrs():")

	clientDnsRrModel := ClientDnsRrModel{}
	err := c.ShouldBindJSON(&clientDnsRrModel)
	if err != nil {
		belogs.Error("DelDnsRrs(): ShouldBindJSON:", err)
		ginserver.ResponseFail(c, err, "")
		return
	}
	belogs.Info("DelDnsRrs():clientDnsRrModel:", jsonutil.MarshalJson(clientDnsRrModel))

	err = delDnsRrs(&clientDnsRrModel)
	if err != nil {
		belogs.Error("DelDnsRrs(): delDnsRrs fail, clientDnsRrModel:", jsonutil.MarshalJson(clientDnsRrModel), err)
		ginserver.ResponseFail(c, err, "")
		return
	}
	ginserver.ResponseOk(c, nil)
}

/*
	    {
			"origin": "example.com",
			"ttl": 0,
			"rrModels": [
				{
					"origin": "example.com.",
					"rrName": "test1",
					"rrType": "A",
					"rrClass": "NONE",
					"rrTtl": 0,
					"rrData": "1.1.1.1"
				}
			]
		}
*/
func DigServerDnsRrs(c *gin.Context) {

	belogs.Info("QueryServerDnsRrs():")

	clientDnsRrModel := ClientDnsRrModel{}
	err := c.ShouldBindJSON(&clientDnsRrModel)
	if err != nil {
		belogs.Error("QueryServerDnsRrs(): ShouldBindJSON:", err)
		ginserver.ResponseFail(c, err, "")
		return
	}
	belogs.Info("DigServerDnsRrs(): clientDnsRrModel:", jsonutil.MarshalJson(clientDnsRrModel))

	responseDnsModel, err := digServerDnsRrs(&clientDnsRrModel)
	if err != nil {
		belogs.Error("DigServerDnsRrs(): digServerDnsRrs fail:", err)
		ginserver.ResponseFail(c, err, "")
		return
	}

	ginserver.ResponseOk(c, responseDnsModel)

}
func QueryClientDnsRrs(c *gin.Context) {
	clientDnsRrModel := ClientDnsRrModel{}
	err := c.ShouldBindJSON(&clientDnsRrModel)
	if err != nil {
		belogs.Error("QueryClientDnsRrs(): ShouldBindJSON:", err)
		ginserver.ResponseFail(c, err, "")
		return
	}
	belogs.Info("QueryClientDnsRrs(): clientDnsRrModel:", jsonutil.MarshalJson(clientDnsRrModel))

	resultDnsRrs, err := queryClientDnsRrs(&clientDnsRrModel)
	if err != nil {
		belogs.Error("QueryClientDnsRrs(): queryClientDnsRrs fail:", err)
		ginserver.ResponseFail(c, err, "")
		return
	}

	ginserver.ResponseOk(c, resultDnsRrs)
}
func QueryClientAllDnsRrs(c *gin.Context) {
	belogs.Info("QueryClientAllDnsRrs():")
	resultDnsRrs, err := queryClientAllDnsRrs()
	if err != nil {
		belogs.Error("QueryClientAllDnsRrs(): queryClientAllDnsRrs fail:", err)
		ginserver.ResponseFail(c, err, "")
		return
	}

	ginserver.ResponseOk(c, resultDnsRrs)

}
func ClearClientAllDnsRrs(c *gin.Context) {
	belogs.Info("ClearClientAllDnsRrs():")
	err := clearClientAllDnsRrs()
	if err != nil {
		belogs.Error("ClearClientAllDnsRrs(): clearClientAllDnsRrs fail:", err)
		ginserver.ResponseFail(c, err, "")
		return
	}
	ginserver.ResponseOk(c, nil)

}

/*
	{
		"inactivityTimeout": 1500,
		"keepaliveInterval": 1500
	}
*/
func StartKeepalive(c *gin.Context) {
	belogs.Info("StartKeepalive():")
	clientKeepaliveModel := ClientKeepaliveModel{}
	err := c.ShouldBindJSON(&clientKeepaliveModel)
	if err != nil {
		belogs.Error("StartKeepalive(): ShouldBindJSON:", err)
		ginserver.ResponseFail(c, err, "")
		return
	}
	belogs.Info("StartKeepalive(): clientKeepaliveModel:", jsonutil.MarshalJson(clientKeepaliveModel))

	err = startKeepalive(clientKeepaliveModel)
	if err != nil {
		belogs.Error("Keepalive(): fail:", err)
		ginserver.ResponseFail(c, err, "")
		return
	}
	ginserver.ResponseOk(c, nil)
}

// subscribeRr:
/*
// only one
{
	"rrModels": [
		{
			"rrFullDomain": "dns1.example.com",
			"rrType": "A",
			"rrClass": "IN"
		}
	]
}
*/
func SubscribeDnsRr(c *gin.Context) {

	belogs.Info("SubscribeDnsRr():")

	clientDnsRrModel := ClientDnsRrModel{}
	err := c.ShouldBindJSON(&clientDnsRrModel)
	if err != nil {
		belogs.Error("SubscribeDnsRr(): ShouldBindJSON:", err)
		ginserver.ResponseFail(c, err, "")
		return
	}
	belogs.Info("SubscribeDnsRr():clientDnsRrModel:", jsonutil.MarshalJson(clientDnsRrModel))

	err = subscribeDnsRr(&clientDnsRrModel)
	if err != nil {
		belogs.Error("SubscribeDnsRr(): subscribeDnsRr fail:", err)
		ginserver.ResponseFail(c, err, "")
		return
	}
	ginserver.ResponseOk(c, nil)
}

func UnsubscribeDnsRr(c *gin.Context) {

	belogs.Info("UnsubscribeDnsRr():")
	clientDnsRrModel := ClientDnsRrModel{}
	err := c.ShouldBindJSON(&clientDnsRrModel)
	if err != nil {
		belogs.Error("UnsubscribeDnsRr(): ShouldBindJSON:", err)
		ginserver.ResponseFail(c, err, "")
		return
	}
	belogs.Info("UnsubscribeDnsRr():clientDnsRrModel:", jsonutil.MarshalJson(clientDnsRrModel))

	err = unsubscribeDnsRr(&clientDnsRrModel)
	if err != nil {
		belogs.Error("UnsubscribeDnsRr(): unsubscribeDnsRr fail:", err)
		ginserver.ResponseFail(c, err, "")
		return
	}
	ginserver.ResponseOk(c, nil)

}

/*
func TriggerReconfirm(c *gin.Context) {

	belogs.Info("TriggerReconfirm():")

		reconfirmRr := dnsmodel.RrModel{}
		err := c.ShouldBindJSON(&reconfirmRr)
		if err != nil {
			belogs.Error("UnsubscribeDnsRr(): ShouldBindJSON:", err)
			ginserver.ResponseFail(c, err, "")
			return
		}

		err = triggerReconfirm(&reconfirmRr)
		if err != nil {
			belogs.Error("TriggerReconfirm(): triggerReconfirm fail:", err)
			ginserver.ResponseFail(c, err, "")
			return
		}

	ginserver.ResponseOk(c, nil)
}
*/



func ReceivePreceptRpki(c *gin.Context) {
	belogs.Info("ReceivePreceptRpki():")
	preceptRpki := PreceptRpki{}
	err := c.ShouldBindJSON(&preceptRpki)
	if err != nil {
		belogs.Error("ReceivePreceptRpki(): ShouldBindJSON:", err)
		ginserver.ResponseFail(c, err, "")
		return
	}
	belogs.Info("ReceivePreceptRpki(): preceptRpki:", jsonutil.MarshalJson(preceptRpki))

	go receivePreceptRpki(&preceptRpki)
	ginserver.ResponseOk(c, nil)
}
