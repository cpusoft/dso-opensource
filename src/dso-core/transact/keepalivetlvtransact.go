package transact

import (
	"errors"
	"time"

	connect "dns-connect"
	"dns-model/message"
	dsomodel "dso-core/model"
	"dso-core/task"
	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/conf"
	"github.com/cpusoft/goutil/dnsutil"
	"github.com/cpusoft/goutil/jsonutil"
)

// return dsoError
func performKeepaliveTlvTransact(dnsConnect *connect.DnsConnect,
	receiveSimpleDsoModel *dsomodel.SimpleDsoModel, dnsToProcessMsg *message.DnsToProcessMsg) (responseTlvModel dsomodel.TlvModel, err error) {
	belogs.Debug("performKeepaliveTlvTransact():dnsConnect:", jsonutil.MarshalJson(dnsConnect),
		"  receiveKeepaliveTlvModel:", jsonutil.MarshalJson(receiveSimpleDsoModel),
		"  dnsToProcessMsg:", jsonutil.MarshalJson(dnsToProcessMsg))

	// simpleMessage --> KeepaliveTlv
	receiveKeepaliveTlvModel, ok := (receiveSimpleDsoModel.TlvModel).(*dsomodel.KeepaliveTlvModel)
	if !ok {
		belogs.Error("performKeepaliveTlvTransactInServer(): receiveTlvModel is not receiveKeepaliveTlvModel:", jsonutil.MarshalJson(receiveSimpleDsoModel))
		return nil, errors.New("receiveTlvModel is not receiveKeepaliveTlvModel:" + jsonutil.MarshalJson(receiveSimpleDsoModel))
	}
	belogs.Debug("performKeepaliveTlvTransactInServer(): receive KeepaliveTlvModel:", jsonutil.MarshalJson(receiveKeepaliveTlvModel))

	// check if < min
	if receiveKeepaliveTlvModel.InactivityTimeout < dnsutil.DSO_MIN_INACTIVITY_TIMEOUT_SECONDS ||
		receiveKeepaliveTlvModel.KeepaliveInterval < dnsutil.DSO_MIN_KEEPALIVE_INTERVAL_SECONDS {
		belogs.Error("performKeepaliveTlvTransact(): inactivityTimeout or KeepaliveInterval is smaller than min value: ", jsonutil.MarshalJson(receiveKeepaliveTlvModel))
		return nil, errors.New("inactivityTimeout or KeepaliveInterval is smaller than min value:" + jsonutil.MarshalJson(receiveKeepaliveTlvModel))
	}

	// mssageId must not be zero
	belogs.Debug("performKeepaliveTlvTransact():messageId:", receiveSimpleDsoModel.MessageId)
	if receiveSimpleDsoModel.MessageId == 0 {
		belogs.Error("performKeepaliveTlvTransact():messageId must not be zero, fail:", jsonutil.MarshalJson(receiveKeepaliveTlvModel))
		return nil, errors.New("MessageId must not be zero in keepalive receive. it is :" + jsonutil.MarshalJson(receiveKeepaliveTlvModel))
	}

	switch dnsToProcessMsg.DnsTransactSide {
	case message.DNS_TRANSACT_SIDE_SERVER:
		responseTlvModel, err = performKeepaliveTlvTransactInServer(dnsConnect, receiveKeepaliveTlvModel, receiveSimpleDsoModel.TlvIndex)
	case message.DNS_TRANSACT_SIDE_CLIENT:
		responseTlvModel, err = performKeepaliveTlvTransactInClient(dnsConnect, receiveKeepaliveTlvModel, receiveSimpleDsoModel.TlvIndex)
	default:
		belogs.Error("performKeepaliveTlvTransact(): dnsTransactSide for keepalive is not supported, fail:", dnsToProcessMsg.DnsTransactSide)
		return nil, errors.New("dnsTransactSide for keepalive is not supported")
	}
	if err != nil {
		belogs.Error("performKeepaliveTlvTransact(): performKeepaliveTlvTransactIn*** fail:dnsConnect:", jsonutil.MarshalJson(dnsConnect),
			"  receiveKeepaliveTlvModel:", jsonutil.MarshalJson(receiveSimpleDsoModel),
			"  dnsToProcessMsg:", jsonutil.MarshalJson(dnsToProcessMsg), err)
		return nil, err
	}
	belogs.Debug("performKeepaliveTlvTransactInServer(): return responseTlvModel:", jsonutil.MarshalJson(responseTlvModel))
	return responseTlvModel, nil
}

// return dsoError
func performKeepaliveTlvTransactInServer(dnsConnect *connect.DnsConnect,
	receiveKeepaliveTlvModel *dsomodel.KeepaliveTlvModel, tlvIndex uint64) (responseTlvModel dsomodel.TlvModel, err error) {
	start := time.Now()

	// sessionless --> session
	sessionState, err := dnsConnect.GetSessionState()
	if err != nil {
		belogs.Error("performKeepaliveTlvTransactInServer(): GetSessionState fail, dnsConnect:", jsonutil.MarshalJson(dnsConnect), err)
		return nil, errors.New("receiveTlvModel is not from dso connect")
	}

	if tlvIndex == 0 && sessionState == dnsutil.DSO_SESSION_STATE_CONNECTED_SESSIONLESS {
		belogs.Debug("performKeepaliveTlvTransactInServer():SessionState is connected_sessionless and tlvIndex==0,",
			" will to established session and CheckSessionTimeoutInServer", sessionState,
			" tlvIndex:", tlvIndex)
		// start session timeout ticker
		dnsConnect.EstablishedSession()
		go task.CheckSessionTimeoutInServer(dnsConnect)

	} else if sessionState == dnsutil.DSO_SESSION_STATE_ESTABLISHED_SESSION {
		// set keep session
		belogs.Debug("performKeepaliveTlvTransactInServer():SessionState is established_session, will keep session",
			sessionState, " tlvIndex:", tlvIndex)
		dnsConnect.SetCanKeepSession(true)
	}
	newSessionState, _ := dnsConnect.GetSessionState()
	belogs.Debug("performKeepaliveTlvTransactInServer():session established, newSessionState:", newSessionState,
		"   time(s):", time.Since(start))

	// send response
	responseTlvModel = dsomodel.NewKeepaliveTlvModel(uint32(conf.Int("keepalive::inactivityTimeout")),
		uint32(conf.Int("keepalive::keepaliveInterval")))
	belogs.Info("performKeepaliveTlvTransactInServer():responseTlvModel:", jsonutil.MarshalJson(responseTlvModel),
		"   time(s):", time.Since(start))
	return responseTlvModel, nil
}

// retun dsoError
// ignore responseTlvModel
func performKeepaliveTlvTransactInClient(dnsConnect *connect.DnsConnect,
	receiveKeepaliveTlvModel *dsomodel.KeepaliveTlvModel, tlvIndex uint64) (responseTlvModel dsomodel.TlvModel, err error) {
	start := time.Now()

	// sessionless --> session
	sessionState, err := dnsConnect.GetSessionState()
	if err != nil {
		belogs.Error("performKeepaliveTlvTransactInClient(): GetSessionState fail, dnsConnect:", jsonutil.MarshalJson(dnsConnect), err)
		return nil, errors.New("receiveTlvModel is not from dso connect")
	}
	if sessionState == dnsutil.DSO_SESSION_STATE_CONNECTED_SESSIONLESS {
		belogs.Debug("performKeepaliveTlvTransactInClient():SessionState is connected_sessionless and tlvIndex==0, will to established session",
			sessionState, " tlvIndex:", tlvIndex)
		// start session timeout ticker
		dnsConnect.EstablishedSession()
	} else if sessionState == dnsutil.DSO_SESSION_STATE_ESTABLISHED_SESSION {
		// set keep session
		dnsConnect.SetCanKeepSession(true)
	}
	newSessionState, _ := dnsConnect.GetSessionState()
	belogs.Debug("performKeepaliveTlvTransactInClient():session established, newSessionState:", newSessionState)

	// will save InactivityTimeout and KeepaliveInterval
	dnsConnect.InactivityTimeout = receiveKeepaliveTlvModel.InactivityTimeout
	dnsConnect.KeepaliveInterval = receiveKeepaliveTlvModel.KeepaliveInterval
	belogs.Info("performKeepaliveTlvTransactInClient(): dnsConnect.InactivityTimeout:", dnsConnect.InactivityTimeout,
		"    dnsConnect.KeepaliveInterval:", dnsConnect.KeepaliveInterval)
	// client, will keep
	//go task.KeepSessionTimeoutInClient(dnsConnect)
	responseTlvModel = dsomodel.NewKeepaliveTlvModel(receiveKeepaliveTlvModel.InactivityTimeout,
		receiveKeepaliveTlvModel.KeepaliveInterval)
	belogs.Info("performKeepaliveTlvTransactInClient():responseTlvModel:", jsonutil.MarshalJson(responseTlvModel),
		"   time(s):", time.Since(start))
	return responseTlvModel, nil
}
