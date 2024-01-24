package transact

import (
	"errors"
	"strings"
	"time"

	connect "dns-connect"
	"dns-model/message"
	dsomodel "dso-core/model"
	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/dnsutil"
	"github.com/cpusoft/goutil/jsonutil"
)

// return dsoError
func performReconfirmTlvTransact(dnsConnect *connect.DnsConnect,
	receiveSimpleDsoModel *dsomodel.SimpleDsoModel, dnsToProcessMsg *message.DnsToProcessMsg) (responseTlvModel dsomodel.TlvModel, err error) {
	belogs.Debug("performReconfirmTlvTransact():dnsConnect:", jsonutil.MarshalJson(dnsConnect),
		"  receiveSimpleDsoModel:", jsonutil.MarshalJson(receiveSimpleDsoModel),
		"  dnsToProcessMsg:", jsonutil.MarshalJson(dnsToProcessMsg))

	// simpleMessage --> ReconfirmTlv
	receiveReconfirmTlvModel, ok := (receiveSimpleDsoModel.TlvModel).(*dsomodel.ReconfirmTlvModel)
	if !ok {
		belogs.Error("performReconfirmTlvTransactInServer(): receiveTlvModel is not receiveReconfirmTlvModel:", jsonutil.MarshalJson(receiveSimpleDsoModel))
		return nil, errors.New("receiveTlvModel is not receiveReconfirmTlvModel:" + jsonutil.MarshalJson(receiveSimpleDsoModel))
	}
	belogs.Debug("performReconfirmTlvTransactInServer(): receiveReconfirmTlvModel:", jsonutil.MarshalJson(receiveReconfirmTlvModel),
		" ok:", ok)

	sessionState, err := dnsConnect.GetSessionState()
	if err != nil {
		belogs.Error("performReconfirmTlvTransact(): GetSessionState fail, dnsConnect:", jsonutil.MarshalJson(dnsConnect), err)
		return nil, errors.New("receiveTlvModel is not from dso connect")
	}
	if sessionState != dnsutil.DSO_SESSION_STATE_ESTABLISHED_SESSION {
		belogs.Error("performReconfirmTlvTransact():SessionState is not established_session: ", sessionState)
		return nil, errors.New("SessionState is not established_session: " + sessionState)
	}

	// mssageId must be zero
	belogs.Debug("performReconfirmTlvTransact():messageId:", receiveSimpleDsoModel.MessageId)
	if receiveSimpleDsoModel.MessageId != 0 {
		belogs.Error("performReconfirmTlvTransact(): messageId must be zero, fail, receiveSimpleDsoModel:", jsonutil.MarshalJson(receiveSimpleDsoModel))
		return nil, errors.New("MessageId must be zero in reconfirm receive. it is :" + jsonutil.MarshalJson(receiveReconfirmTlvModel))
	}

	// qr must be receive
	belogs.Debug("performReconfirmTlvTransact(): qr:", receiveSimpleDsoModel.Qr)
	if receiveSimpleDsoModel.Qr != dnsutil.DNS_QR_REQUEST {
		belogs.Error("performReconfirmTlvTransact(): qr must be receive, fail:", jsonutil.MarshalJson(receiveReconfirmTlvModel))
		return nil, errors.New("qr must be receive in push reconfirm. it is :" + jsonutil.MarshalJson(receiveReconfirmTlvModel))
	}

	belogs.Debug("performReconfirmTlvTransact(): dnsType:", receiveReconfirmTlvModel.DnsPacketModel.PacketType)
	if receiveReconfirmTlvModel.DnsPacketModel.PacketType == dnsutil.DNS_TYPE_INT_ANY {
		belogs.Error("performReconfirmTlvTransact(): type must notbe 0xff(255): ", jsonutil.MarshalJson(receiveReconfirmTlvModel))
		return nil, errors.New("Type must notbe 0xff(255): " + jsonutil.MarshalJson(receiveReconfirmTlvModel))
	}

	belogs.Debug("performReconfirmTlvTransact(): dnsClass:", receiveReconfirmTlvModel.DnsPacketModel.PacketClass)
	if receiveReconfirmTlvModel.DnsPacketModel.PacketClass == dnsutil.DNS_CLASS_INT_ANY {
		belogs.Error("performReconfirmTlvTransact(): class must notbe 0xff(255): ", jsonutil.MarshalJson(receiveReconfirmTlvModel))
		return nil, errors.New("Class must notbe 0xff(255): " + jsonutil.MarshalJson(receiveReconfirmTlvModel))
	}

	dnsName := string([]byte(receiveReconfirmTlvModel.DnsPacketModel.PacketDomain.FullDomain))
	belogs.Debug("performReconfirmTlvTransact(): dnsName:", dnsName)
	if strings.HasPrefix(dnsName, "*") {
		belogs.Error("performReconfirmTlvTransact(): Wildcarding is not supported: ", jsonutil.MarshalJson(receiveReconfirmTlvModel))
		return nil, errors.New("Wildcarding is not supported: " + jsonutil.MarshalJson(receiveReconfirmTlvModel))
	}

	// set keep session
	dnsConnect.SetCanKeepSession(true)
	switch dnsToProcessMsg.DnsTransactSide {
	case message.DNS_TRANSACT_SIDE_SERVER:
		return performReconfirmTlvTransactInServer(dnsConnect, receiveReconfirmTlvModel)
	case message.DNS_TRANSACT_SIDE_CLIENT:
		belogs.Error("performReconfirmTlvTransact(): client cannot receive reconfirm, fail, dnsToProcessMsg:", jsonutil.MarshalJson(dnsToProcessMsg), "  receiveSimpleDsoModel:", jsonutil.MarshalJson(receiveSimpleDsoModel))
		return nil, errors.New("client cannot receive reconfirm:" + jsonutil.MarshalJson(receiveSimpleDsoModel))
	default:
		belogs.Error("performReconfirmTlvTransact(): dnsTransactSide for reconfirm is not supported, fail:", dnsToProcessMsg.DnsTransactSide)
		return nil, errors.New("dnsTransactSide for reconfirm is not supported")
	}
}

// return dsoError
func performReconfirmTlvTransactInServer(dnsConnect *connect.DnsConnect,
	receiveReconfirmTlvModel *dsomodel.ReconfirmTlvModel) (responseTlvModel dsomodel.TlvModel, err error) {
	start := time.Now()

	/* shaodebug
	// send subscribe
	connKey := transportutil.GetTcpConnKey(dnsConnect.TcpConn)
	dnsDomain := string([]byte(receiveReconfirmTlvModel.DnsName))
	dnsType := receiveReconfirmTlvModel.DnsType
	dnsClass := receiveReconfirmTlvModel.DnsClass
	dnsData := string([]byte(receiveReconfirmTlvModel.DnsRData))
	belogs.Info("performReconfirmTlvTransactInServer():connKey:", connKey,
		"   dnsDomain:", string(dnsDomain), "  dnsType:", dnsType, "   dnsClass:", dnsClass,
		"   dnsData:", dnsData, "   time(s):", time.Since(start))
	err = sendToPushServer("reconfirm", connKey, receiveSimpleDsoModel.MessageId,
		dnsDomain, dnsType, dnsClass, dnsData)
	if err != nil {
		belogs.Error("performReconfirmTlvTransactInServer(): sendToPushServer fail:",
			jsonutil.MarshalJson(receiveReconfirmTlvModel), err)
		return nil, err
	}
	*/
	// send nil, will not be added to responseDsoModel
	belogs.Info("performReconfirmTlvTransactInServer(): reponse nil,   time(s):", time.Since(start))
	return nil, nil

}
