package transact

import (
	"errors"

	clientcache "dns-client-cache"
	connect "dns-connect"
	"dns-model/message"
	"dns-model/rr"
	dsomodel "dso-core/model"
	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/dnsutil"
	"github.com/cpusoft/goutil/jsonutil"
)

// return dsoError
func performPushTlvTransact(dnsConnect *connect.DnsConnect,
	receiveSimpleDsoModel *dsomodel.SimpleDsoModel, dnsToProcessMsg *message.DnsToProcessMsg) (responseTlvModel dsomodel.TlvModel, err error) {
	belogs.Debug("performPushTlvTransact():dnsConnect:", jsonutil.MarshalJson(dnsConnect),
		"  receiveSimpleDsoModel:", jsonutil.MarshalJson(receiveSimpleDsoModel), "  dnsToProcessMsg:", jsonutil.MarshalJson(dnsToProcessMsg))

	receivePushTlvModel, ok := (receiveSimpleDsoModel.TlvModel).(*dsomodel.PushTlvModel)
	if !ok {
		belogs.Error("performPushTlvTransact(): receiveTlvModel is not receivePushTlvModel:", jsonutil.MarshalJson(receiveSimpleDsoModel))
		return nil, errors.New("receiveTlvModel is not receivePushTlvModel:" + jsonutil.MarshalJson(receiveSimpleDsoModel))
	}
	belogs.Debug("performPushTlvTransact(): receivePushTlvModel:", jsonutil.MarshalJson(receivePushTlvModel), ok)

	sessionState, err := dnsConnect.GetSessionState()
	if err != nil {
		belogs.Error("performPushTlvTransact(): GetSessionState fail, dnsConnect:", jsonutil.MarshalJson(dnsConnect), err)
		return nil, errors.New("receiveTlvModel is not from dso connect")
	}
	if sessionState != dnsutil.DSO_SESSION_STATE_ESTABLISHED_SESSION {
		belogs.Error("performPushTlvTransact():SessionState is not established_session: ", sessionState)
		return nil, errors.New("SessionState is not established_session: " + sessionState)
	}

	// mssageId must  be zero
	belogs.Debug("performPushTlvTransact():messageId:", receiveSimpleDsoModel.MessageId)
	if receiveSimpleDsoModel.MessageId != 0 {
		belogs.Error("performPushTlvTransact(): messageId must be zero, fail, receiveSimpleDsoModel::", jsonutil.MarshalJson(receiveSimpleDsoModel))
		return nil, errors.New("MessageId must be zero in push. it is :" + jsonutil.MarshalJson(receiveSimpleDsoModel))
	}

	// qr must be receive
	belogs.Debug("performPushTlvTransact(): qr:", receiveSimpleDsoModel.Qr)
	if receiveSimpleDsoModel.Qr != dnsutil.DNS_QR_RESPONSE {
		belogs.Error("performPushTlvTransact(): qr of push must be response, fail, receiveSimpleDsoModel:", jsonutil.MarshalJson(receiveSimpleDsoModel))
		return nil, errors.New("qr must be response in push receive. it is :" + jsonutil.MarshalJson(receivePushTlvModel))
	}

	// set keep session
	dnsConnect.SetCanKeepSession(true)
	switch dnsToProcessMsg.DnsTransactSide {
	case message.DNS_TRANSACT_SIDE_SERVER:
		belogs.Error("performPushTlvTransact(): server cannot receive push:", jsonutil.MarshalJson(receivePushTlvModel))
		return nil, errors.New("server cannot receive push:" + jsonutil.MarshalJson(receivePushTlvModel))
	case message.DNS_TRANSACT_SIDE_CLIENT:
		return performPushTlvTransactInClient(dnsConnect, receivePushTlvModel, dnsToProcessMsg)
	default:
		belogs.Error("performPushTlvTransact(): dnsTransactSide for push is not supported, fail:", dnsToProcessMsg.DnsTransactSide)
		return nil, errors.New("dnsTransactSide for push is not supported")
	}

}

// return dsoError
func performPushTlvTransactInClient(dnsConnect *connect.DnsConnect,
	receivePushTlvModel *dsomodel.PushTlvModel, dnsToProcessMsg *message.DnsToProcessMsg) (responseTlvModel dsomodel.TlvModel, err error) {

	belogs.Debug("performPushTlvTransactInClient(): len(receivePushTlvModel.DnsRrModels):", len(receivePushTlvModel.DnsRrModels))
	for i := range receivePushTlvModel.DnsRrModels {
		isDel := rr.IsDelRrModelForDso(receivePushTlvModel.DnsRrModels[i].RrTtl)
		belogs.Debug("performPushTlvTransactInClient(): isDel", isDel)

		err = clientcache.UpdateRrModel(receivePushTlvModel.DnsRrModels[i], isDel)
		if err != nil {
			belogs.Error("performPushTlvTransactInClient(): UpdateRrModel fail")
			return nil, err
		}
	}
	return
}

/*  shaodebug
	for i := range receivePushTlvModel.DnsPacketModels {
		belogs.Debug("performPushTlvTransactInClient(): dsoDataModel i:", i, jsonutil.MarshalJson(receivePushTlvModel.DnsPacketModels[i]))

		dnsType := dnsutil.DnsIntTypes[dsoDataModels[i].DnsType]
		dnsClass := dnsutil.DnsIntClasses[dsoDataModels[i].DnsClass]
		dnsTtl := null.IntFrom(int64(dsoDataModels[i].DnsTtl))
		dnsRdLen := dsoDataModels[i].DnsRdLen
		// have specified rr: may add or del

			if dnsRdLen > 0 {
				belogs.Debug("performPushTlvTransactInClient(): dnsRdLen > 0, dnsType:", dnsType, "  dnsClass:", dnsClass,
					" dnsTtl:", dnsTtl, "   dnsRdLen:", dnsRdLen)
				rrValues := strings.Split(string(dsoDataModels[i].DnsRData), " ")
				// add rr

				rr := zonefileutil.NewResourceRecord(string(dsoDataModels[i].DnsName), "", dnsType,
					dnsClass, dnsTtl, rrValues)
				belogs.Debug("performPushTlvTransactInClient(): add rr, rr:", jsonutil.MarshalJson(rr))

				if !zonefileutil.IsDelResourceRecord(dnsTtl) {
					exist, err := clientcache.AddKnownResourceRecord(rr)
					if err != nil {
						belogs.Error("performPushTlvTransactInClient(): AddKnownResourceRecord fail,dsoDataModels[i]:", jsonutil.MarshalJson(dsoDataModels[i]))
						return nil, dnsutil.NewDnsError(dnsutil.DNS_RCODE_SERVFAIL,
							transportutil.NEXT_CONNECT_POLICY_KEEP,
							"performPushTlvTransactInClient fail:"+jsonutil.MarshalJson(dsoDataModels[i]))
					}
					belogs.Debug("performPushTlvTransactInClient(): dsoDataModels,  exist:", exist, "  specified rr:", jsonutil.MarshalJson(rr))
				} else {
					// del  specified rr
					err = clientcache.DelKnownResourceRecord(rr)
					if err != nil {
						belogs.Error("performPushTlvTransactInClient(): dnsRdLen > 0, DelKnownResourceRecord fail,dsoDataModels[i]:", jsonutil.MarshalJson(dsoDataModels[i]))
						return nil, dnsutil.NewDnsError(dnsutil.DNS_RCODE_SERVFAIL,
							transportutil.NEXT_CONNECT_POLICY_KEEP,
							"performPushTlvTransactInClient fail:"+jsonutil.MarshalJson(dsoDataModels[i]))
					}
					belogs.Debug("performPushTlvTransactInClient(): del specified rr, rr:", jsonutil.MarshalJson(rr))
				}
			} else {
				// should be del collective rr
				belogs.Debug("performPushTlvTransactInClient(): dnsRdLen == 0, dnsType:", dnsType, "  dnsClass:", dnsClass,
					" dnsTtl:", dnsTtl, "   dnsRdLen:", dnsRdLen)
				rr := zonefileutil.NewResourceRecord(string(dsoDataModels[i].DnsName), "", dnsType,
					dnsClass, dnsTtl, nil)
				err = clientcache.DelKnownResourceRecord(rr)
				if err != nil {
					belogs.Error("performPushTlvTransactInClient(): dnsRdLen == 0, DelKnownResourceRecord fail,dsoDataModels[i]:", jsonutil.MarshalJson(dsoDataModels[i]))
					return nil, dnsutil.NewDnsError(dnsutil.DNS_RCODE_SERVFAIL,
								"performPushTlvTransactInClient fail:"+jsonutil.MarshalJson(dsoDataModels[i]))
				}
				belogs.Debug("performPushTlvTransactInClient(): del collective rr, rr:", jsonutil.MarshalJson(rr))
			}

	}
	// response is nil
	return nil, nil
}
*/
