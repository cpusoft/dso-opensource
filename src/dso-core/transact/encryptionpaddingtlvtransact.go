package transact

import (
	"errors"
	"time"

	connect "dns-connect"
	"dns-model/message"
	dsomodel "dso-core/model"
	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/dnsutil"
	"github.com/cpusoft/goutil/jsonutil"
)

// return dsoError
func performEncryptionPaddingTlvTransact(dnsConnect *connect.DnsConnect,
	receiveSimpleDsoModel *dsomodel.SimpleDsoModel, dnsToProcessMsg *message.DnsToProcessMsg) (responseTlvModel dsomodel.TlvModel, err error) {
	start := time.Now()
	belogs.Debug("performEncryptionPaddingTlvTransact():dnsConnect:", jsonutil.MarshalJson(dnsConnect),
		"  receiveEncryptionPaddingTlvModel:", jsonutil.MarshalJson(receiveSimpleDsoModel), "  dnsToProcessMsg:", jsonutil.MarshalJson(dnsToProcessMsg))

	// simpleMessage --> EncryptionPaddingTlv
	receiveEncryptionPaddingTlvModel, ok := (receiveSimpleDsoModel.TlvModel).(*dsomodel.EncryptionPaddingTlvModel)
	if !ok {
		belogs.Error("performEncryptionPaddingTlvTransact(): receiveTlvModel is not receiveEncryptionPaddingTlvModel:", jsonutil.MarshalJson(receiveSimpleDsoModel))
		return nil, errors.New("receiveTlvModel is not receiveEncryptionPaddingTlvModel")
	}
	belogs.Debug("performEncryptionPaddingTlvTransact(): receiveEncryptionPaddingTlvModel:", jsonutil.MarshalJson(receiveEncryptionPaddingTlvModel),
		" ok:", ok, "   time(s):", time.Since(start))

	sessionState, err := dnsConnect.GetSessionState()
	if err != nil {
		belogs.Error("performEncryptionPaddingTlvTransact(): GetSessionState fail, dnsConnect:", jsonutil.MarshalJson(dnsConnect), err)
		return nil, errors.New("receiveTlvModel is not from dso connect")
	}
	if sessionState == dnsutil.DSO_SESSION_STATE_ESTABLISHED_SESSION {
		belogs.Error("performEncryptionPaddingTlvTransact():SessionState is not established_session, sessionState: ", sessionState)
		return nil, errors.New("SessionState is not established_session: " + sessionState)
	}

	belogs.Debug("performEncryptionPaddingTlvTransact():tlvIndex:", receiveSimpleDsoModel.TlvIndex)
	if receiveSimpleDsoModel.TlvIndex == 0 {
		belogs.Error("performEncryptionPaddingTlvTransact():encryptionPadding should not be primary tlv: ", receiveSimpleDsoModel.TlvIndex)
		return nil, errors.New("encryptionPadding should not be primary tlv")

	}

	// set keep session
	dnsConnect.SetCanKeepSession(true)

	// no different in server/client

	return nil, nil
}
