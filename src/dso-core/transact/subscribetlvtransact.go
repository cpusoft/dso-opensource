package transact

import (
	"errors"
	"strings"
	"time"

	connect "dns-connect"
	"dns-model/message"
	pushmodel "dns-model/push"
	dsomodel "dso-core/model"
	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/conf"
	"github.com/cpusoft/goutil/convert"
	"github.com/cpusoft/goutil/dnsutil"
	"github.com/cpusoft/goutil/httpclient"
	"github.com/cpusoft/goutil/jsonutil"
	"github.com/cpusoft/goutil/transportutil"
	"github.com/guregu/null"
)

// return dsoError
func performSubscribeTlvTransact(dnsConnect *connect.DnsConnect,
	receiveSimpleDsoModel *dsomodel.SimpleDsoModel, dnsToProcessMsg *message.DnsToProcessMsg) (responseTlvModels []dsomodel.TlvModel, err error) {
	belogs.Debug("performSubscribeTlvTransact():dnsConnect:", jsonutil.MarshalJson(dnsConnect),
		"  receiveSimpleDsoModel:", jsonutil.MarshalJson(receiveSimpleDsoModel), "  dnsToProcessMsg:", jsonutil.MarshalJson(dnsToProcessMsg))

	// simpleMessage --> SubscribeTlv
	receiveSubscribeTlvModel, ok := (receiveSimpleDsoModel.TlvModel).(*dsomodel.SubscribeTlvModel)
	if !ok {
		belogs.Error("performSubscribeTlvTransact(): receiveTlvModel is not receiveSubscribeTlvModel:", jsonutil.MarshalJson(receiveSimpleDsoModel))
		return nil, errors.New("receiveTlvModel is not receiveSubscribeTlvModel:" + jsonutil.MarshalJson(receiveSimpleDsoModel))
	}
	belogs.Debug("performSubscribeTlvTransact(): receiveSubscribeTlvModel:", jsonutil.MarshalJson(receiveSubscribeTlvModel),
		" ok:", ok)

	sessionState, err := dnsConnect.GetSessionState()
	if err != nil {
		belogs.Error("performSubscribeTlvTransact(): GetSessionState fail, dnsConnect:", jsonutil.MarshalJson(dnsConnect), err)
		return nil, errors.New("receiveTlvModel is not from dso connect")
	}
	if sessionState != dnsutil.DSO_SESSION_STATE_ESTABLISHED_SESSION {
		belogs.Error("performSubscribeTlvTransact():SessionState is not established_session: ", sessionState)
		return nil, errors.New("SessionState is not established_session: " + sessionState)
	}

	// mssageId must not be zero
	belogs.Debug("performSubscribeTlvTransact():messageId:", receiveSimpleDsoModel.MessageId)
	if receiveSimpleDsoModel.MessageId == 0 {
		belogs.Error("performSubscribeTlvTransact(): messageId must not be zero, fail:", jsonutil.MarshalJson(receiveSubscribeTlvModel))
		return nil, errors.New("MessageId must not be zero in subscribe receive. it is :" + jsonutil.MarshalJson(receiveSubscribeTlvModel))
	}

	dnsName := string([]byte(receiveSubscribeTlvModel.DnsName))
	belogs.Debug("performSubscribeTlvTransact(): dnsName:", dnsName)
	if strings.HasPrefix(dnsName, "*") {
		belogs.Error("performSubscribeTlvTransact(): Wildcarding is not supported: ", jsonutil.MarshalJson(receiveSubscribeTlvModel))
		return nil, errors.New("Wildcarding is not supported: " + jsonutil.MarshalJson(receiveSubscribeTlvModel))
	}
	// set keep session
	dnsConnect.SetCanKeepSession(true)
	switch dnsToProcessMsg.DnsTransactSide {
	case message.DNS_TRANSACT_SIDE_SERVER:
		return performSubscribeTlvTransactInServer(dnsConnect, receiveSimpleDsoModel.MessageId, receiveSubscribeTlvModel, dnsToProcessMsg)
	case message.DNS_TRANSACT_SIDE_CLIENT:
		return performSubscribeTlvTransactInClient(dnsConnect, receiveSimpleDsoModel.MessageId, receiveSubscribeTlvModel, dnsToProcessMsg)
	default:
		belogs.Error("performSubscribeTlvTransact(): dnsTransactSide for subscribe is not supported, fail:", dnsToProcessMsg.DnsTransactSide)
		return nil, errors.New("dnsTransactSide for subscribe is not supported")
	}

}

// return dsoError
func performSubscribeTlvTransactInServer(dnsConnect *connect.DnsConnect,
	subscribeMessageId uint16, receiveSubscribeTlvModel *dsomodel.SubscribeTlvModel, dnsToProcessMsg *message.DnsToProcessMsg) (responseTlvModels []dsomodel.TlvModel, err error) {
	start := time.Now()

	// send to push
	go func() {
		connKey := transportutil.GetTcpConnKey(dnsConnect.TcpConn)
		rrFullDomain := string([]byte(receiveSubscribeTlvModel.DnsNamePacketDomain.FullDomain))
		rrType := dnsutil.DnsIntTypes[receiveSubscribeTlvModel.DnsType]
		rrClass := dnsutil.DnsIntClasses[receiveSubscribeTlvModel.DnsClass]

		belogs.Info("performSubscribeTlvTransactInServer():connKey:", connKey,
			"   rrFullDomain:", rrFullDomain, "  rrType:", rrType, "   rrClass:", rrClass,
			"   dnsData:", "", " subscribeMessageId:", subscribeMessageId, "   time(s):", time.Since(start))
		pushRrModel := pushmodel.NewPushRrModel(
			connKey, rrFullDomain, rrType, rrClass, null.NewInt(0, false), "", subscribeMessageId)
		path := "https://" + conf.String("dns-server::serverHost") + ":" + conf.String("dns-server::serverHttpsPort") +
			"/push/subscribe"
		pushResultRrModels := make([]*pushmodel.PushResultRrModel, 0)
		belogs.Debug("performSubscribeTlvTransactInServer(): path:", path, "  pushRrModel:", jsonutil.MarshalJson(pushRrModel))
		err = httpclient.PostAndUnmarshalResponseModel(path, jsonutil.MarshalJson(pushRrModel), false, &pushResultRrModels)
		if err != nil {
			belogs.Error("performSubscribeTlvTransactInServer(): httpclient/push/subscribe, fail:", path,
				"  pushRrModels:", jsonutil.MarshalJson(pushRrModel), err)
			return
		}
		if len(pushResultRrModels) == 0 {
			belogs.Debug("performSubscribeTlvTransactInServer():httpclient/push/subscribe have no results, path:", path,
				"  pushRrModel:", jsonutil.MarshalJson(pushRrModel), " time(s):", time.Since(start))
			return
		}
		belogs.Info("#执行DSO的'订阅'时,发现已有域名数据满足了订阅要求,触发DSO的'数据推送(PUSH)'到对应的客户端, 推送的域名数据:")
		for i := range pushResultRrModels {
			for j := range pushResultRrModels[i].RrModels {
				belogs.Info("#{'域名':'" + pushResultRrModels[i].RrModels[j].RrFullDomain +
					"','Type':'" + pushResultRrModels[i].RrModels[j].RrType +
					"','Class':'" + pushResultRrModels[i].RrModels[j].RrClass +
					"','Ttl':" + convert.ToString(pushResultRrModels[i].RrModels[j].RrTtl.ValueOrZero()) +
					",'Data':" + pushResultRrModels[i].RrModels[j].RrData + "'}")
			}
		}

		pushResultRrModelsJson := jsonutil.MarshalJson(pushResultRrModels)
		belogs.Debug("performSubscribeTlvTransactInServer(): httpclient/push/subscribe, pushResultRrModels:", pushResultRrModelsJson)
		dnsMsg := message.NewDnsMsg(message.DNS_MSG_TYPE_DSO_SERVER_PUSH_TO_CLIENT, pushResultRrModelsJson)
		dnsToProcessMsg.SendDnsToProcessMsgChan(dnsMsg)
		belogs.Info("performSubscribeTlvTransactInServer(): httpclient/push/subscribe,will send dnsMsg:", dnsMsg, " time(s):", time.Since(start))
	}()
	// send nil, will not be added to responseDsoModel
	belogs.Info("performSubscribeTlvTransactInServer(): reponse nil, time(s):", time.Since(start))
	return responseTlvModels, nil

}

// return dsoError
func performSubscribeTlvTransactInClient(dnsConnect *connect.DnsConnect,
	subscribeMessageId uint16, receiveSubscribeTlvModel *dsomodel.SubscribeTlvModel, dnsToProcessMsg *message.DnsToProcessMsg) (responseTlvModels []dsomodel.TlvModel, err error) {
	return
}
