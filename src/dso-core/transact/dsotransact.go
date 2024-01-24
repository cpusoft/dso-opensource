package transact

import (
	"errors"

	connect "dns-connect"
	"dns-model/common"
	"dns-model/message"
	dsomodel "dso-core/model"
	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/dnsutil"
	"github.com/cpusoft/goutil/jsonutil"
	"github.com/cpusoft/goutil/transportutil"
)

func PerformDsoTransact(dnsConnect *connect.DnsConnect, receiveDnsModel common.DnsModel,
	dnsToProcessMsg *message.DnsToProcessMsg) (responseDnsModel common.DnsModel, err error) { //responseDsoModel *dsomodel.DsoModel, err error) {
	belogs.Debug("PerformDsoTransact():dnsConnect:", jsonutil.MarshalJson(dnsConnect),
		"   receiveDsoModel:", jsonutil.MarshalJson(receiveDnsModel), "   dnsToProcessMsg:", jsonutil.MarshalJson(dnsToProcessMsg))
	/*
		receiveJson := jsonutil.MarshalJson(receiveDnsModel)
		var receiveDsoModel *dsomodel.DsoModel
		err = jsonutil.UnmarshalJson(receiveJson, receiveDsoModel)
		if err != nil {
			belogs.Error("PerformDsoTransact():UnmarshalJson receiveJson fail:", receiveJson, err)
			return nil, err
		}
	*/
	receiveDsoModel, ok := (receiveDnsModel).(*dsomodel.DsoModel)
	if !ok {
		belogs.Error("PerformDsoTransact(): receiveDnsModel to receiveDsoModel,fail:", jsonutil.MarshalJson(receiveDnsModel), err)
		return nil, dnsutil.NewDnsError("fail to convert model type",
			0, dnsutil.DNS_OPCODE_UPDATE, dnsutil.DNS_RCODE_SERVFAIL, transportutil.NEXT_CONNECT_POLICY_KEEP)
	}
	belogs.Debug("PerformDsoTransact(): receiveDnsModel to receiveDsoModel,  receiveDsoModel:", jsonutil.MarshalJson(receiveDsoModel))

	responseDsoModel, err := performDsoHeaderTransact(dnsConnect, receiveDsoModel, dnsToProcessMsg)
	if err != nil {
		belogs.Error("PerformDsoTransact(): performDsoHeaderTransact fail, dnsConnect:", jsonutil.MarshalJson(dnsConnect),
			"   receiveDsoModel:", jsonutil.MarshalJson(receiveDsoModel), "   dnsToProcessMsg:", jsonutil.MarshalJson(dnsToProcessMsg), err)
		return nil, err
	}
	belogs.Debug("PerformDsoTransact(): performDsoHeaderTransact,  responseDsoModel(no tlv):", jsonutil.MarshalJson(*responseDsoModel))

	// get tlvmodels
	for tlvIndex := range receiveDsoModel.DsoDataModel.TlvModels {
		belogs.Debug("PerformDsoTransact(): will transact receive TlvModels:",
			jsonutil.MarshalJson(receiveDsoModel.DsoDataModel.TlvModels[tlvIndex]))
		receiveSimpleDsoModel := &dsomodel.SimpleDsoModel{
			MessageId: receiveDsoModel.HeaderForDsoModel.GetIdOrMessageId(),
			Qr:        receiveDsoModel.HeaderForDsoModel.GetQr(),
			RCode:     receiveDsoModel.HeaderForDsoModel.GetRCode(),
			TlvIndex:  uint64(tlvIndex),
			TlvModel:  receiveDsoModel.DsoDataModel.TlvModels[tlvIndex],
		}
		belogs.Debug("PerformDsoTransact(): will transact receive receiveSimpleDsoModel:", receiveSimpleDsoModel)

		// common tlv transact
		err = performTlvTransact(dnsConnect, receiveSimpleDsoModel, dnsToProcessMsg)
		if err != nil {
			belogs.Error("PerformDsoTransact(): callTransactMessageFunc fail:", err)
			return nil, err
		}

		// switch dsoType
		var responseTlvModels []dsomodel.TlvModel
		var responseTlvModel dsomodel.TlvModel
		belogs.Debug("PerformDsoTransact(): switch dsoType:", receiveSimpleDsoModel.TlvModel.GetDsoType())
		switch receiveSimpleDsoModel.TlvModel.GetDsoType() {
		case dnsutil.DSO_TYPE_KEEPALIVE:
			responseTlvModel, err = performKeepaliveTlvTransact(dnsConnect,
				receiveSimpleDsoModel, dnsToProcessMsg)
		case dnsutil.DSO_TYPE_RETRY_DELAY:
			responseTlvModel, err = performRetryDelayTlvTransact(dnsConnect,
				receiveSimpleDsoModel, dnsToProcessMsg)
		case dnsutil.DSO_TYPE_ENCRYPTION_PADDING:
			responseTlvModel, err = performEncryptionPaddingTlvTransact(dnsConnect,
				receiveSimpleDsoModel, dnsToProcessMsg)
		case dnsutil.DSO_TYPE_SUBSCRIBE:
			responseTlvModels, err = performSubscribeTlvTransact(dnsConnect,
				receiveSimpleDsoModel, dnsToProcessMsg)
		case dnsutil.DSO_TYPE_PUSH:
			responseTlvModel, err = performPushTlvTransact(dnsConnect,
				receiveSimpleDsoModel, dnsToProcessMsg)
		case dnsutil.DSO_TYPE_UNSUBSCRIBE:
			responseTlvModel, err = performUnsubscribeTlvTransact(dnsConnect,
				receiveSimpleDsoModel, dnsToProcessMsg)
		case dnsutil.DSO_TYPE_RECONFIRM:
			responseTlvModel, err = performReconfirmTlvTransact(dnsConnect,
				receiveSimpleDsoModel, dnsToProcessMsg)
		default:
		}
		if err != nil {
			belogs.Error("PerformDsoTransact():perform***TlvTransact fail,  tlvIndex: ", tlvIndex,
				"  receiveSimpleDsoModel:", jsonutil.MarshalJson(receiveSimpleDsoModel))
			return nil, err
		}

		belogs.Debug("PerformDsoTransact():responseTlvModels:", jsonutil.MarshalJson(responseTlvModels))
		if responseTlvModel != nil {
			responseDsoModel.AddTlvModel(responseTlvModel)
		}
		if len(responseTlvModels) != 0 {
			responseDsoModel.AddTlvModels(responseTlvModels)
		}
	}

	belogs.Info("PerformDsoTransact():after add tlv, responseDsoModel:", jsonutil.MarshalJson(*responseDsoModel))
	return responseDsoModel, nil
}

// return dsoError
func performDsoHeaderTransact(dnsConnect *connect.DnsConnect,
	receiveDsoModel *dsomodel.DsoModel, dnsToProcessMsg *message.DnsToProcessMsg) (responseDsoModel *dsomodel.DsoModel, err error) {
	belogs.Debug("performDsoHeaderTransact():dnsConnect:", jsonutil.MarshalJson(dnsConnect),
		"   receiveDsoModel:", jsonutil.MarshalJson(receiveDsoModel), "   dnsToProcessMsg:", jsonutil.MarshalJson(dnsToProcessMsg))

	// when sessionless , the primary tlv should be keepalive
	sessionState, err := dnsConnect.GetSessionState()
	if err != nil {
		belogs.Error("performDsoHeaderTransact(): GetSessionState fail, dnsConnect:", jsonutil.MarshalJson(dnsConnect), err)
		return nil, errors.New("receiveTlvModel is not from dso connect")
	}
	if sessionState == dnsutil.DSO_SESSION_STATE_CONNECTED_SESSIONLESS {
		exist, receiveDsoType := dsomodel.GetPrimaryTlvDsoType(receiveDsoModel)
		if !exist {
			belogs.Error("performDsoHeaderTransact(): receiveDsoType is not exist:", jsonutil.MarshalJson(receiveDsoModel))
			return nil, errors.New("there should have primary tlv when state is sessionless")
		}

		belogs.Debug("performDsoHeaderTransact():SessionState is connected_sessionless, sessionState:", sessionState, " receiveDsoType:", receiveDsoType)
		if receiveDsoType != dnsutil.DSO_TYPE_KEEPALIVE {
			belogs.Error("performDsoHeaderTransact():SessionState is connected_sessionless and receiveDsoType is not keepalive, fail:", receiveDsoType)
			return nil, errors.New("First primary tlv is not  dnsutil.DSO_TYPE_KEEPALIVE")
		}
	}

	messageId := receiveDsoModel.HeaderForDsoModel.MessageId
	qr := uint8(dnsutil.DNS_QR_RESPONSE)
	rCode := uint8(dnsutil.DNS_RCODE_NOERROR)
	responseDsoModel, _ = dsomodel.NewDsoModelByParameters(messageId, qr, rCode)

	// get dsomodel header (have no tlv)
	belogs.Debug("performDsoHeaderTransact():dnsToProcessMsg:", jsonutil.MarshalJson(dnsToProcessMsg))
	switch dnsToProcessMsg.DnsTransactSide {
	case message.DNS_TRANSACT_SIDE_SERVER:
		err = performDsoHeaderTransactInServer(dnsConnect, receiveDsoModel, responseDsoModel)
	case message.DNS_TRANSACT_SIDE_CLIENT:
		err = performDsoHeaderTransactInClient(dnsConnect, receiveDsoModel, responseDsoModel)
	default:
		belogs.Error("performDsoHeaderTransact(): dnsTransactSide for header is not supported, fail:", dnsToProcessMsg.DnsTransactSide)
		return nil, errors.New("dnsTransactSide for header is not supported")
	}
	if err != nil {
		belogs.Error("performDsoHeaderTransact(): performDsoHeaderTransactIn*** fail:,dnsConnect:", jsonutil.MarshalJson(dnsConnect),
			"   receiveDsoModel:", jsonutil.MarshalJson(receiveDsoModel), "   responseDsoModel:", jsonutil.MarshalJson(responseDsoModel),
			"   dnsToProcessMsg:", jsonutil.MarshalJson(dnsToProcessMsg), err)
		return nil, err
	}

	return responseDsoModel, nil
}

func performDsoHeaderTransactInServer(dnsConnect *connect.DnsConnect,
	receiveDsoModel *dsomodel.DsoModel, responseDsoModel *dsomodel.DsoModel) (err error) {

	return nil
}

func performDsoHeaderTransactInClient(dnsConnect *connect.DnsConnect,
	receiveDsoModel *dsomodel.DsoModel, responseDsoModel *dsomodel.DsoModel) (err error) {

	// check messageId from server, maybe client subscribe's messageId
	messageId := receiveDsoModel.GetHeaderModel().GetIdOrMessageId()
	rCode := receiveDsoModel.GetHeaderModel().GetRCode()
	belogs.Debug("performDsoHeaderTransactInClient(): messageId:", messageId, "   rCode:", rCode)
	if messageId > 0 {
		// shaodebug
		//clientcache.UpdateSubscribeResourceRecordResult(messageId, rCode)
	}

	return nil
}
