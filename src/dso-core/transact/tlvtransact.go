package transact

import (
	"errors"

	connect "dns-connect"
	"dns-model/message"
	dsomodel "dso-core/model"
	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/dnsutil"
	"github.com/cpusoft/goutil/jsonutil"
	"github.com/cpusoft/goutil/transportutil"
)

// return dsoError
func performTlvTransact(dnsConnect *connect.DnsConnect,
	receiveSimpleDsoModel *dsomodel.SimpleDsoModel, dnsToProcessMsg *message.DnsToProcessMsg) (err error) {
	switch dnsToProcessMsg.DnsTransactSide {
	case message.DNS_TRANSACT_SIDE_SERVER:
		err = performTlvTransactInServer(dnsConnect, receiveSimpleDsoModel)
	case message.DNS_TRANSACT_SIDE_CLIENT:
		err = performTlvTransactInClient(dnsConnect, receiveSimpleDsoModel)
	default:
		belogs.Error("performTlvTransact(): dnsTransactSide for tlv is not supported, fail:", dnsToProcessMsg.DnsTransactSide)
		return errors.New("dnsTransactSide for tlv is not supported")
	}
	if err != nil {
		belogs.Error("performTlvTransact(): callTransactMessageFunc fail:", err)
		return err
	}
	return nil
}

// common check in server
// return dsoError
func performTlvTransactInServer(dnsConnect *connect.DnsConnect,
	receiveSimpleDsoModel *dsomodel.SimpleDsoModel) (err error) {

	// check
	messageId := receiveSimpleDsoModel.MessageId
	// QR: server receive qr should be 0
	qr := receiveSimpleDsoModel.Qr
	belogs.Debug("performTlvTransactInServer():qr:", qr)
	if qr != dnsutil.DNS_QR_REQUEST {
		belogs.Error("performTlvTransactInServer(): QR is not DNS_QR_REQUEST(0), fail:",
			jsonutil.MarshalJson(receiveSimpleDsoModel), err)
		return dnsutil.NewDnsError("Qr should be Request(0) in receiveTlvModel",
			messageId, dnsutil.DNS_OPCODE_DSO, dnsutil.DNS_RCODE_REFUSED, transportutil.NEXT_CONNECT_POLICY_CLOSE_FORCIBLE)
	}

	// rcode: server receive rcode should be noerror
	rCode := receiveSimpleDsoModel.RCode
	belogs.Debug("performTlvTransactInServer():rCode:", rCode)
	if rCode != dnsutil.DNS_RCODE_NOERROR {
		belogs.Error("performTlvTransactInServer(): RCode should be NOERROR(0), fail:",
			jsonutil.MarshalJson(receiveSimpleDsoModel), err)
		return dnsutil.NewDnsError("RCode should be NOERROR(0) in receiveTlvModel.",
			messageId, dnsutil.DNS_OPCODE_DSO, dnsutil.DNS_RCODE_REFUSED, transportutil.NEXT_CONNECT_POLICY_CLOSE_FORCIBLE)
	}
	return nil
}

// return dsoError
func performTlvTransactInClient(dnsConnect *connect.DnsConnect,
	receiveSimpleDsoModel *dsomodel.SimpleDsoModel) (err error) {

	// check
	return nil
}
