package transact

import (
	"errors"
	"time"

	connect "dns-connect"
	"dns-model/message"
	pushmodel "dns-model/push"
	dsomodel "dso-core/model"
	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/conf"
	"github.com/cpusoft/goutil/dnsutil"
	"github.com/cpusoft/goutil/httpclient"
	"github.com/cpusoft/goutil/jsonutil"
	"github.com/cpusoft/goutil/transportutil"
)

// return responseTlvModel is nil
func performUnsubscribeTlvTransact(dnsConnect *connect.DnsConnect,
	receiveSimpleDsoModel *dsomodel.SimpleDsoModel, dnsToProcessMsg *message.DnsToProcessMsg) (responseTlvModel dsomodel.TlvModel, err error) {
	belogs.Debug("performUnsubscribeTlvTransact():dnsConnect:", jsonutil.MarshalJson(dnsConnect),
		"  receiveSimpleDsoModel:", jsonutil.MarshalJson(receiveSimpleDsoModel), "  dnsToProcessMsg:", jsonutil.MarshalJson(dnsToProcessMsg))

	// simpleMessage --> UnsubscribeTlv
	receiveUnsubscribeTlvModel, ok := (receiveSimpleDsoModel.TlvModel).(*dsomodel.UnsubscribeTlvModel)
	if !ok {
		belogs.Error("performUnsubscribeTlvTransact(): receiveTlvModel is not receiveUnsubscribeTlvModel:", jsonutil.MarshalJson(receiveSimpleDsoModel))
		return nil, errors.New("receiveTlvModel is not receiveUnsubscribeTlvModel:" + jsonutil.MarshalJson(receiveSimpleDsoModel))
	}
	belogs.Debug("performUnsubscribeTlvTransact(): receiveUnsubscribeTlvModel:", jsonutil.MarshalJson(receiveUnsubscribeTlvModel), " ok:", ok)

	sessionState, err := dnsConnect.GetSessionState()
	if err != nil {
		belogs.Error("performUnsubscribeTlvTransact(): GetSessionState fail, dnsConnect:", jsonutil.MarshalJson(dnsConnect), err)
		return nil, errors.New("receiveTlvModel is not from dso connect")
	}
	if sessionState != dnsutil.DSO_SESSION_STATE_ESTABLISHED_SESSION {
		belogs.Error("performUnsubscribeTlvTransact():SessionState is not established_session: ", sessionState)
		return nil, errors.New("SessionState is not established_session: " + sessionState)
	}
	// mssageId must be zero
	// this messageid is not subscribeMessageId
	belogs.Debug("performUnsubscribeTlvTransact():messageId:", receiveSimpleDsoModel.MessageId)
	if receiveSimpleDsoModel.MessageId != 0 {
		belogs.Error("performUnsubscribeTlvTransact(): messageId must be zero, fail:", jsonutil.MarshalJson(receiveSimpleDsoModel))
		return nil, errors.New("MessageId must be zero in unsubscribe receive. it is :" + jsonutil.MarshalJson(receiveSimpleDsoModel))
	}
	// set keep session
	dnsConnect.SetCanKeepSession(true)
	switch dnsToProcessMsg.DnsTransactSide {
	case message.DNS_TRANSACT_SIDE_SERVER:
		return performUnsubscribeTlvTransactInServer(dnsConnect, receiveUnsubscribeTlvModel, dnsToProcessMsg)
	case message.DNS_TRANSACT_SIDE_CLIENT:
		return performUnsubscribeTlvTransactInClient(dnsConnect, receiveUnsubscribeTlvModel, dnsToProcessMsg)
	default:
		belogs.Error("performUnsubscribeTlvTransact(): dnsTransactSide for unsubscribe is not supported, fail:", dnsToProcessMsg.DnsTransactSide)
		return nil, errors.New("dnsTransactSide for unsubscribe is not supported")
	}

}

// return dsoError
func performUnsubscribeTlvTransactInServer(dnsConnect *connect.DnsConnect,
	receiveUnsubscribeTlvModel *dsomodel.UnsubscribeTlvModel, dnsToProcessMsg *message.DnsToProcessMsg) (responseTlvModel dsomodel.TlvModel, err error) {
	start := time.Now()
	// send to push
	go func() {
		connKey := transportutil.GetTcpConnKey(dnsConnect.TcpConn)
		subscribeMessageId := receiveUnsubscribeTlvModel.SubscribeMessageId
		unpushRrModel := pushmodel.NewUnpushRrModel(
			connKey, subscribeMessageId)
		path := "https://" + conf.String("dns-server::serverHost") + ":" + conf.String("dns-server::serverHttpsPort") +
			"/push/unsubscribe"
		belogs.Debug("performUnsubscribeTlvTransactInServer(): path:", path, "  unpushRrModel:", jsonutil.MarshalJson(unpushRrModel))
		err = httpclient.PostAndUnmarshalResponseModel(path, jsonutil.MarshalJson(unpushRrModel), false, nil)
		if err != nil {
			belogs.Error("performUnsubscribeTlvTransactInServer(): httpclient/push/unsubscribe, fail:", path,
				"  pushRrModels:", jsonutil.MarshalJson(unpushRrModel), err)
			return
		}
		belogs.Info("performUnsubscribeTlvTransactInServer(): httpclient/push/unsubscribe, time(s):", time.Since(start))
	}()
	// send nil, will not be added to responseDsoModel
	belogs.Info("performUnsubscribeTlvTransactInServer(): reponse nil, time(s):", time.Since(start))
	return responseTlvModel, nil

}

// return dsoError
func performUnsubscribeTlvTransactInClient(dnsConnect *connect.DnsConnect,
	receiveUnsubscribeTlvModel *dsomodel.UnsubscribeTlvModel, dnsToProcessMsg *message.DnsToProcessMsg) (responseTlvModel dsomodel.TlvModel, err error) {
	return
}
