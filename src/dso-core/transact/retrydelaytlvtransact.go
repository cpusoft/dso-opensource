package transact

import (
	"errors"

	connect "dns-connect"
	"dns-model/message"
	dsomodel "dso-core/model"
	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/dnsutil"
	"github.com/cpusoft/goutil/jsonutil"
)

// return dsoError
func performRetryDelayTlvTransact(dnsConnect *connect.DnsConnect,
	receiveSimpleDsoModel *dsomodel.SimpleDsoModel, dnsToProcessMsg *message.DnsToProcessMsg) (responseIgnoreTlvModel dsomodel.TlvModel, err error) {
	belogs.Debug("performRetryDelayTlvTransact():dnsConnect:", jsonutil.MarshalJson(dnsConnect),
		"  receiveSimpleDsoModel:", jsonutil.MarshalJson(receiveSimpleDsoModel),
		"  dnsToProcessMsg:", jsonutil.MarshalJson(dnsToProcessMsg))

	sessionState, err := dnsConnect.GetSessionState()
	if err != nil {
		belogs.Error("performRetryDelayTlvTransact(): GetSessionState fail, dnsConnect:", jsonutil.MarshalJson(dnsConnect), err)
		return nil, errors.New("receiveTlvModel is not from dso connect")
	}
	if sessionState != dnsutil.DSO_SESSION_STATE_ESTABLISHED_SESSION {
		belogs.Error("performRetryDelayTlvTransact():SessionState is not established_session: ", sessionState)
		return nil, errors.New("SessionState is not established_session: " + sessionState)
	}

	// set keep session
	dnsConnect.SetCanKeepSession(true)

	switch dnsToProcessMsg.DnsTransactSide {
	case message.DNS_TRANSACT_SIDE_SERVER:
		belogs.Error("performRetryDelayTlvTransact(): server cannot receive retryDelay, fail, dnsToProcessMsg:", jsonutil.MarshalJson(dnsToProcessMsg))
		return nil, errors.New("server cannot receive retryDelay:" + jsonutil.MarshalJson(receiveSimpleDsoModel))
	case message.DNS_TRANSACT_SIDE_CLIENT:
		// shaodebug
		// do nothing
		belogs.Debug("performRetryDelayTlvTransact(): client just ignore retryDelay,  dnsToProcessMsg:", jsonutil.MarshalJson(dnsToProcessMsg))
		return nil, nil
	default:
		belogs.Error("performRetryDelayTlvTransact(): dnsTransactSide for retrydelay is not supported, fail:", dnsToProcessMsg.DnsTransactSide)
		return nil, errors.New("dnsTransactSide for retrydelay is not supported")
	}

}
