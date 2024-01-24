package drive

import (
	"errors"
	"strings"
	"time"

	clientcache "dns-client-cache"
	dnsclient "dns-client/dns"
	"dns-model/common"
	"dns-model/rr"
	dsomodel "dso-core/model"
	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/conf"
	"github.com/cpusoft/goutil/dnsutil"
	"github.com/cpusoft/goutil/httpclient"
	"github.com/cpusoft/goutil/jsonutil"
	"github.com/cpusoft/goutil/rrdputil"
	"github.com/cpusoft/goutil/urlutil"
	"github.com/guregu/null"
	querymodel "query-core/model"
	updatemodel "update-core/model"
)

func createDnsConnect() (err error) {
	start := time.Now()

	serverProtocol := conf.String("dns-server::serverProtocol")
	serverHost := conf.String("dns-server::serverHost")
	tcpPort := conf.String("dns-server::serverTcpPort")
	tlsPort := conf.String("dns-server::serverTlsPort")
	udpPort := conf.String("dns-server::serverUdpPort")

	var tcpTlsPort string
	if strings.Contains(serverProtocol, "tcp") {
		tcpTlsPort = tcpPort
	} else if strings.Contains(serverProtocol, "tls") {
		tcpTlsPort = tlsPort
	}
	belogs.Debug("createDnsConnect(): serverProtocol:", serverProtocol, "  serverHost:", serverHost,
		"  tcpPort:", tcpPort, "  tlsPort:", tlsPort, "  udpPort:", udpPort)

	err = dnsclient.StartDnsClient(serverProtocol, serverHost, tcpTlsPort, udpPort)
	if err != nil {
		belogs.Error("createDnsConnect(): StartDnsClient fail, serverProtocol:", serverProtocol,
			"  tlsPort:", tlsPort, "  tcpPort:", tcpPort, "  udpPort:", udpPort, err)
	}
	belogs.Info("createDnsConnect(): StartDnsClient ok, serverProtocol:", serverProtocol,
		"  tlsPort:", tlsPort, "  tcpPort:", tcpPort, "  udpPort:", udpPort, " time(s):", time.Since(start))
	belogs.Info("#客户端与服务器成功建立连接")
	return nil
}

func closeDnsConnect() (err error) {

	err = dnsclient.StopDnsClient()
	if err != nil {
		belogs.Error("closeDnsConnect(): StopDnsClient fail:", err)
		return err
	}
	return nil
}

// update:add
func addDnsRrs(clientDnsRrModel *ClientDnsRrModel) (err error) {
	belogs.Debug("addDnsRr(): clientDnsRrModel:", jsonutil.MarshalJson(clientDnsRrModel))
	start := time.Now()
	// check
	//if !checkDnsRrs(rrModels) {
	//	}

	// to updateModel
	id, err := clientcache.GetNewMessageId(dnsutil.DNS_OPCODE_UPDATE)
	if err != nil {
		belogs.Error("addDnsRrs(): GetNewMessageId for DNS_OPCODE_UPDATE fail:", err)
		return err
	}
	belogs.Debug("addDnsRr(): GetNewMessageId for DNS_OPCODE_UPDATE, id:", id)

	updateModel, err := updatemodel.NewUpdateModelByParametersAndRrModels(id,
		clientDnsRrModel.Origin, clientDnsRrModel.Ttl.ValueOrZero(), dnsutil.DNS_CLASS_INT_IN,
		clientDnsRrModel.RrModels)
	if err != nil {
		belogs.Error("addDnsRrs(): NewUpdateModelByParametersAndRrModels fail, clientDnsRrModel:", jsonutil.MarshalJson(clientDnsRrModel), err)
		return err
	}
	// send
	responseDnsModel, err := dnsclient.SendTcpDnsModel(updateModel, true)
	if err != nil {
		belogs.Error("addDnsRrs(): SendTcpDnsModel fail, updateModel:", jsonutil.MarshalJson(updateModel), err)
		return err
	}

	belogs.Info("addDnsRrs(): send updateModel:", jsonutil.MarshalJson(updateModel),
		"  responseDnsModel:", jsonutil.MarshalJson(responseDnsModel), "  times:", time.Since(start))
	return nil
}

// update Del
func delDnsRrs(clientDnsRrModel *ClientDnsRrModel) (err error) {
	belogs.Debug("delDnsRrs(): clientDnsRrModel:", jsonutil.MarshalJson(clientDnsRrModel))
	start := time.Now()
	//if !checkDnsRrs(rrModels) {
	//	}
	// to updateModel
	id, err := clientcache.GetNewMessageId(dnsutil.DNS_OPCODE_UPDATE)
	if err != nil {
		belogs.Error("delDnsRrs(): GetNewMessageId  for DNS_OPCODE_UPDATE fail:", err)
		return err
	}
	belogs.Debug("delDnsRrs(): GetNewMessageId  for DNS_OPCODE_UPDATE, id:", id)

	//DNS_TYPE_INT_ANY
	updateModel, err := updatemodel.NewUpdateModelByParametersAndRrModels(id,
		clientDnsRrModel.Origin, clientDnsRrModel.Ttl.ValueOrZero(), dnsutil.DNS_CLASS_INT_NONE,
		clientDnsRrModel.RrModels)
	if err != nil {
		belogs.Error("delDnsRrs(): NewUpdateModelByParametersAndRrModels fail, clientDnsRrModel:", jsonutil.MarshalJson(clientDnsRrModel), err)
		return err
	}
	// send
	responseDnsModel, err := dnsclient.SendTcpDnsModel(updateModel, true)
	if err != nil {
		belogs.Error("delDnsRrs(): SendTcpDnsModel fail, updateModel:", jsonutil.MarshalJson(updateModel), err)
		return err
	}
	belogs.Info("delDnsRrs(): send updateModel:", jsonutil.MarshalJson(updateModel),
		"  responseDnsModel:", jsonutil.MarshalJson(responseDnsModel), "  times:", time.Since(start))

	return nil
}

// query:dig
func digServerDnsRrs(clientDnsRrModel *ClientDnsRrModel) (responseDnsModel common.DnsModel, err error) {
	belogs.Debug("digServerDnsRrs(): clientDnsRrModel:", jsonutil.MarshalJson(clientDnsRrModel))
	start := time.Now()
	//if !checkDnsRr(rrModel) {
	//
	//	}

	if len(clientDnsRrModel.RrModels) != 1 {
		belogs.Error("digServerDnsRrs(): len(clientDnsRrModel.RrModels) isnot 1, clientDnsRrModel:", jsonutil.MarshalJson(clientDnsRrModel))
		return nil, errors.New("len(clientDnsRrModel.RrModels) isnot one")
	}

	id, err := clientcache.GetNewMessageId(dnsutil.DNS_OPCODE_QUERY)
	if err != nil {
		belogs.Error("digServerDnsRrs(): GetNewMessageId for DNS_OPCODE_QUERY fail:", err)
		return nil, err
	}
	belogs.Debug("digServerDnsRrs(): GetNewMessageId for DNS_OPCODE_QUERY, id:", id)

	queryModel, err := querymodel.NewQueryModelForQuestionByParametersAndRrModels(id, clientDnsRrModel.Ttl.ValueOrZero(),
		clientDnsRrModel.RrModels[0])
	if err != nil {
		belogs.Error("digServerDnsRrs(): NewQueryModelForQuestionByParametersAndRrModels fail:", err)
		return nil, err
	}
	belogs.Debug("digServerDnsRrs(): queryModel:", jsonutil.MarshalJson(queryModel))

	// send
	responseDnsModel, err = dnsclient.SendUdpDnsModel(queryModel, true)
	if err != nil {
		belogs.Error("digServerDnsRrs(): SendTcpDnsModel fail, updateModel:", jsonutil.MarshalJson(queryModel), err)
		return nil, err
	}
	belogs.Info("digServerDnsRrs(): send queryModel:", jsonutil.MarshalJson(queryModel),
		"  responseDnsModel:", jsonutil.MarshalJson(responseDnsModel), "  times:", time.Since(start))

	return responseDnsModel, nil
}

func queryClientDnsRrs(clientDnsRrModel *ClientDnsRrModel) (resultDnsRrs []*rr.RrModel, err error) {
	belogs.Debug("queryClientDnsRrs(): clientDnsRrModel:", jsonutil.MarshalJson(clientDnsRrModel))
	if len(clientDnsRrModel.RrModels) != 1 {
		belogs.Error("queryClientDnsRrs(): len(clientDnsRrModel.RrModels) isnot 1, clientDnsRrModel:", jsonutil.MarshalJson(clientDnsRrModel))
		return nil, errors.New("len(clientDnsRrModel.RrModels) isnot one")
	}
	return clientcache.QueryRrModels(clientDnsRrModel.RrModels[0])
}
func queryClientAllDnsRrs() (resultDnsRrs []*rr.RrModel, err error) {
	belogs.Debug("queryClientAllDnsRrs():")

	return clientcache.QueryAllRrModels()
}

func clearClientAllDnsRrs() (err error) {
	belogs.Debug("clearClientAllDnsRrs():")
	return clientcache.ClearAllRrModels()
}

func startKeepalive(clientKeepaliveModel ClientKeepaliveModel) (err error) {
	belogs.Debug("startKeepalive(): clientKeepaliveModel:", jsonutil.MarshalJson(clientKeepaliveModel))
	shouldSessionTimeout := false
	dnsConnect := dnsclient.GetTcpDnsConnect()
	belogs.Debug("startKeepalive(): dnsConnect.InactivityTimeout:", dnsConnect.InactivityTimeout)
	if dnsConnect.InactivityTimeout == 0 {
		shouldSessionTimeout = true
	}
	belogs.Debug("startKeepalive(): dnsConnect.InactivityTimeout ==0,  shouldSessionTimeout:", shouldSessionTimeout)

	err = keepaliveImpl(clientKeepaliveModel.InactivityTimeout, clientKeepaliveModel.KeepaliveInterval)
	if err != nil {
		belogs.Error("startKeepalive(): keepaliveImpl fail:", err)
		return err
	}
	belogs.Debug("startKeepalive(): shouldSessionTimeout:", shouldSessionTimeout)

	if shouldSessionTimeout {
		belogs.Debug("startKeepalive(): go keepSessionInClient")
		go keepSessionInClient(dnsConnect)
	}
	return nil
}

func keepaliveImpl(inactivityTimeout, keepaliveInterval uint32) (err error) {
	belogs.Debug("keepaliveImpl(): inactivityTimeout:", inactivityTimeout, "  keepaliveInterval:", keepaliveInterval)
	start := time.Now()
	// to dsoModel
	id, err := clientcache.GetNewMessageId(dnsutil.DNS_OPCODE_DSO)
	if err != nil {
		belogs.Error("keepaliveImpl(): GetNewMessageId for DNS_OPCODE_DSO fail:", err)
		return err
	}
	belogs.Debug("keepaliveImpl():GetNewMessageId for DNS_OPCODE_DSO, id:", id)
	err = clientcache.UpdateDsoMessageDsoTypeAndRrModel(id, dnsutil.DSO_TYPE_KEEPALIVE, nil)
	if err != nil {
		belogs.Error("keepaliveImpl(): UpdateDsoMessageDsoTypeAndRrModel fail:", err)
		return err
	}
	belogs.Debug("keepaliveImpl():UpdateDsoMessageDsoTypeAndRrModel id:", id)

	// dsomodel
	keepaliveModel := dsomodel.NewDsoModelWithKeepaliveTlvModel(id, dnsutil.DNS_QR_REQUEST,
		dnsutil.DNS_RCODE_NOERROR, inactivityTimeout,
		keepaliveInterval)
	belogs.Debug("keepaliveImpl():keepaliveModel:", jsonutil.MarshalJson(keepaliveModel))

	// send
	responseDnsModel, err := dnsclient.SendTcpDnsModel(keepaliveModel, true)
	if err != nil {
		belogs.Error("keepaliveImpl(): SendTcpDnsModel fail, keepaliveModel:", jsonutil.MarshalJson(keepaliveModel), err)
		return err
	}
	belogs.Info("keepaliveImpl(): send keepaliveModel:", jsonutil.MarshalJson(keepaliveModel),
		"  responseDnsModel:", jsonutil.MarshalJson(responseDnsModel), "  times:", time.Since(start))

	err = confirmMessageId(id, responseDnsModel)
	if err != nil {
		belogs.Error("keepaliveImpl(): confirmMessageId fail, id:", id, err)
		return err
	}

	belogs.Info("keepaliveImpl(): ok, responseDnsModel:", jsonutil.MarshalJson(responseDnsModel), "  times:", time.Since(start))
	return nil
}

func subscribeDnsRr(clientDnsRrModel *ClientDnsRrModel) (err error) {
	belogs.Debug("subscribeDnsRr(): clientDnsRrModel:", jsonutil.MarshalJson(clientDnsRrModel))
	start := time.Now()
	//if !checkDnsRr(clientDnsRrModel) {
	//	}

	// to dsoModel
	id, err := clientcache.GetNewMessageId(dnsutil.DNS_OPCODE_DSO)
	if err != nil {
		belogs.Error("subscribeDnsRr(): GetNewMessageId for DNS_OPCODE_DSO fail:", err)
		return err
	}
	belogs.Debug("subscribeDnsRr():GetNewMessageId for DNS_OPCODE_DSO, id:", id)

	err = clientcache.UpdateDsoMessageDsoTypeAndRrModel(id, dnsutil.DSO_TYPE_SUBSCRIBE, clientDnsRrModel.RrModels[0])
	if err != nil {
		belogs.Error("subscribeDnsRr(): UpdateDsoMessageDsoTypeAndRrModel fail:", err)
		return err
	}
	belogs.Debug("subscribeDnsRr():UpdateDsoMessageDsoTypeAndRrModel id:", id)

	// dsomodel
	subscribeModel, err := dsomodel.NewDsoModelWithSubscribeTlvModel(id, clientDnsRrModel.RrModels[0])
	if err != nil {
		belogs.Error("subscribeDnsRr(): NewDsoModelWithSubscribeTlvModel fail:", err)
		return err
	}
	belogs.Debug("keepaliveImpl():subscribeModel:", jsonutil.MarshalJson(subscribeModel))

	// send
	responseDnsModel, err := dnsclient.SendTcpDnsModel(subscribeModel, true)
	if err != nil {
		belogs.Error("subscribeDnsRr(): SendTcpDnsModel fail, subscribeModel:", jsonutil.MarshalJson(subscribeModel), err)
		return err
	}
	belogs.Debug("keepaliveImpl():responseDnsModel:", jsonutil.MarshalJson(responseDnsModel))

	err = confirmMessageId(id, responseDnsModel)
	if err != nil {
		belogs.Error("subscribeDnsRr(): confirmMessageId fail, id:", id, err)
		return err
	}
	belogs.Info("subscribeDnsRr(): send subscribeModel:", jsonutil.MarshalJson(subscribeModel),
		"  responseDnsModel:", jsonutil.MarshalJson(responseDnsModel), "  times:", time.Since(start))
	return nil
}

func unsubscribeDnsRr(clientDnsRrModel *ClientDnsRrModel) (err error) {
	belogs.Debug("unsubscribeDnsRr(): clientDnsRrModel:", jsonutil.MarshalJson(clientDnsRrModel))
	//if !checkDnsRr(rrModel) {
	//	}
	found, subscribeMessageId, err := clientcache.QueryDsoMessageIdByRrModelDb(clientDnsRrModel.RrModels[0])
	if err != nil {
		belogs.Error("unsubscribeDnsRr(): QueryDsoMessageIdByRrModelDb fail, RrModels[0]:", jsonutil.MarshalJson(clientDnsRrModel.RrModels[0]), err)
		return err
	} else if !found {
		belogs.Error("unsubscribeDnsRr(): QueryDsoMessageIdByRrModelDb ,not found, RrModels[0]:", jsonutil.MarshalJson(clientDnsRrModel.RrModels[0]), err)
		return errors.New("not found this RrModel")
	}
	belogs.Debug("unsubscribeDnsRr(): subscribeMessageId:", subscribeMessageId)

	// dsomodel
	unsubscribeModel, err := dsomodel.NewDsoModelWithUnsubscribeTlvModel(subscribeMessageId)
	if err != nil {
		belogs.Error("unsubscribeDnsRr(): NewDsoModelWithSubscribeTlvModel fail:", err)
		return err
	}
	belogs.Debug("unsubscribeDnsRr():subscribeModel:", jsonutil.MarshalJson(unsubscribeModel))

	// send
	_, err = dnsclient.SendTcpDnsModel(unsubscribeModel, false)
	if err != nil {
		belogs.Error("unsubscribeDnsRr(): SendTcpDnsModel fail, unsubscribeModel:", jsonutil.MarshalJson(unsubscribeModel), err)
		return err
	}
	belogs.Debug("unsubscribeDnsRr(): SendTcpDnsModel ok, unsubscribeModel:", jsonutil.MarshalJson(unsubscribeModel))

	err = clientcache.UpdateDsoMessageUnsubscribeTime(subscribeMessageId)
	if err != nil {
		belogs.Error("unsubscribeDnsRr(): UpdateDsoMessageUnsubscribeTime fail, subscribeMessageId:", subscribeMessageId, err)
		return err
	}
	belogs.Debug("unsubscribeDnsRr(): ok, clientDnsRrModel:", jsonutil.MarshalJson(clientDnsRrModel), " subscribeMessageId:", subscribeMessageId)
	return nil
}

func reconfirmDnsRr(rrModel *rr.RrModel) (err error) {
	belogs.Debug("reconfirmDnsRr(): rrModel:", jsonutil.MarshalJson(rrModel))
	if !checkDnsRr(rrModel) {

	}
	return nil
}
func checkDnsRrs(rrModels []*rr.RrModel) bool {
	for i := range rrModels {
		if !checkDnsRr(rrModels[i]) {
			belogs.Error("checkDnsRrs(): checkDnsRr, fail, rrModel:", rrModels[i])
			return false
		}
	}
	return true
}

func checkDnsRr(rrModel *rr.RrModel) bool {
	return true
}

func confirmMessageId(id uint16, responseDnsModel common.DnsModel) error {
	reponseMessageId := responseDnsModel.GetHeaderModel().GetIdOrMessageId()
	belogs.Debug("confirmMessageId(): id :", id, "  reponseMessageId:", reponseMessageId)
	if id != reponseMessageId {
		belogs.Error("confirmMessageId(): id isnot equal to responseMessageId, fail, id:", id, "   reponseMessageId:", reponseMessageId)
		return errors.New("send id is not equalt to response id")
	}
	confirmed, err := clientcache.ConfirmMessageId(reponseMessageId)
	if err != nil {
		belogs.Error("confirmMessageId(): ConfirmMessageId fail, reponseMessageId:", reponseMessageId, err)
		return err
	}
	if !confirmed {
		belogs.Error("confirmMessageId(): ConfirmMessageId confirmed is false, reponseMessageId:", reponseMessageId, err)
		return errors.New("message id cannot be confirmed")
	}
	belogs.Info("confirmMessageId(): confirmed, id :", id, "  reponseMessageId:", reponseMessageId)
	return nil
}

func receivePreceptRpki(preceptRpki *PreceptRpki) {
	start := time.Now()
	preceptId := preceptRpki.PreceptId
	belogs.Info("receivePreceptRpki(): preceptId:", preceptId)
	if len(preceptRpki.PreceptRpkiDomains) == 0 {
		belogs.Debug("receivePreceptRpki(): len(preceptRpki.PreceptRpkiDomains) is empty, preceptId:", preceptId)
		return
	}

	preceptRpkiType := conf.String("precept::rpkiType")
	belogs.Debug("receivePreceptRpki(): preceptRpkiType:", preceptRpkiType)
	for _, preceptRpkiDomain := range preceptRpki.PreceptRpkiDomains {
		snapshotUrl := preceptRpkiDomain.PreceptRpkiSnapshot.SnapshotUrl
		fullDomain, _ := urlutil.Host(preceptRpkiDomain.NotifyUrl)
		origin := dnsutil.DomainTrimFirstLabel(fullDomain)
		domain := dnsutil.DomainObtainFirstLabel(fullDomain)
		preceptRpkiDomainJson := jsonutil.MarshalJson(preceptRpkiDomain)
		belogs.Debug("receivePreceptRpki(): snapshotUrl:", snapshotUrl, "  fullDomain:", fullDomain,
			"  origin:", origin, "  domain:", domain, "  preceptRpkiDomain:", preceptRpkiDomainJson)

		clientDnsRrModel := NewClientDnsRrModel(origin, null.IntFrom(1000))
		// send url and serial
		if preceptRpkiType == "urlserial" {
			urlRrModel := rr.NewRrModel(origin, domain, dnsutil.DNS_TYPE_STR_TXT, dnsutil.DNS_CLASS_STR_IN,
				null.IntFrom(1000), preceptRpkiDomainJson)
			belogs.Debug("receivePreceptRpki(): notifyUrl:", preceptRpkiDomain.NotifyUrl, "  clientDnsRrModel:", jsonutil.MarshalJson(clientDnsRrModel))
			clientDnsRrModel.AddRrModel(urlRrModel)
			belogs.Debug("receivePreceptRpki(): preceptRpkiType is urlserial, notifyUrl:", preceptRpkiDomain.NotifyUrl, "  urlRrModel:", jsonutil.MarshalJson(urlRrModel))
			belogs.Debug("receivePreceptRpki(): range SnapshotPublishs, will send clientDnsRrModel:", jsonutil.MarshalJson(clientDnsRrModel))
			cloneAndSendAndResetClientDnsRrModel(clientDnsRrModel)
			belogs.Debug("receivePreceptRpki(): range SnapshotPublishs, after send clientDnsRrModel:", jsonutil.MarshalJson(clientDnsRrModel))
			continue
		}

		// send url and serial and Base64
		if preceptRpkiType == "urlserialbase64" {
			timeoutMins := conf.Int("rpstir2-rp::snapshotTimeoutSeconds") / 60
			httpConfig := httpclient.NewHttpClientConfigWithParam(uint64(timeoutMins), 3, "all")
			snapshotTime := time.Now()
			snapshotModel, err := rrdputil.GetRrdpSnapshotWithConfig(snapshotUrl, httpConfig)
			if err != nil {
				belogs.Error("receivePreceptRpki(): GetRrdpSnapshotWithConfig fail, notifyUrl:", preceptRpkiDomain.NotifyUrl,
					"   snapshotUrl:", snapshotUrl, err, "  time(s):", time.Since(snapshotTime))
				continue
			}
			belogs.Debug("receivePreceptRpki(): get snapshotModel, notifyUrl:", preceptRpkiDomain.NotifyUrl, "  snapshotUrl:", snapshotUrl,
				" snapshotModel:", jsonutil.MarshalJson(snapshotModel), "  time(s):", time.Since(snapshotTime))

			/*
				TODO
				1 遍历循环snapshotModel.SnapshotPublishs (SnapshotModel 和 SnapshotPublish定义见https://github.com/cpusoft/goutil/blob/master/rrdputil/rrdpmodel.go)
				2 将preceptRpkiDomainJson和snapshotPublish中值对应赋值给PreceptRpkiSnapshotPublishBase64，其中Serial取MaxSerial,rrdpType=snapshotPublish
				3 生成NewRrModel，前面各个参数参加前面urlRrModel，最后的txt用PreceptRpkiSnapshotPublishBase64的json值
				4 每20个NewRrModel放入c.RrModels = append(c.RrModels, urlRrModel)中，然后调用go addDnsRrs(&c)发送。注意每次新循环时，需要清空RrModels
			*/
			for _, preceptSnapshotPublish := range snapshotModel.SnapshotPublishs {
				preceptRpkiSnapshotPublishBase64 := PreceptRpkiSnapshotPublishBase64{
					RrdpType:    PRECEPT_RPKI_RRDP_TYPE_SNASHOT_PUBLISH,
					NotifyUrl:   preceptRpkiDomain.NotifyUrl,
					SessionId:   preceptRpkiDomain.SessionId,
					Serial:      preceptRpkiDomain.MaxSerial,
					SnapshotUrl: snapshotUrl,
					Url:         preceptSnapshotPublish.Uri,
					Base64:      preceptSnapshotPublish.Base64,
				}
				preceptRpkiSnapshotPublishBase64Json := jsonutil.MarshalJson(preceptRpkiSnapshotPublishBase64)
				snapshotRrModel := rr.NewRrModel(origin, domain, dnsutil.DNS_TYPE_STR_TXT, dnsutil.DNS_CLASS_STR_IN,
					null.IntFrom(1000), preceptRpkiSnapshotPublishBase64Json)
				belogs.Debug("receivePreceptRpki(): range SnapshotPublishs, get snapshotRrModel:", jsonutil.MarshalJson(snapshotRrModel))
				clientDnsRrModel.AddRrModel(snapshotRrModel)
				// if clientDnsRrModel.RrModels==**, will send and reset RrModels
				judgeSendClientDnsRrModel(clientDnsRrModel)
			}
			// if left clientDnsRrModel.RrModels, will send
			sendLeftClientDnsRrModel(clientDnsRrModel)

			for i := range preceptRpkiDomain.PreceptRpkiDeltas {
				deltaUrl := preceptRpkiDomain.PreceptRpkiDeltas[i].DeltaUrl
				belogs.Debug("receivePreceptRpki(): range PreceptRpkiDeltas, notifyUrl:", preceptRpkiDomain.NotifyUrl, "  deltaUrl:", deltaUrl)

				// if need delta base64
				timeoutMins := conf.Int("rpstir2-rp::deltaTimeoutSeconds") / 60
				httpConfig := httpclient.NewHttpClientConfigWithParam(uint64(timeoutMins), 3, "all")
				deltaTime := time.Now()
				deltaModel, err := rrdputil.GetRrdpDeltaWithConfig(deltaUrl, httpConfig)
				if err != nil {
					belogs.Error("receivePreceptRpki(): range PreceptRpkiDeltas, GetRrdpDeltaWithConfig fail, notifyUrl:", preceptRpkiDomain.NotifyUrl,
						"   deltaUrl:", deltaUrl, err, "  time(s):", time.Since(deltaTime))
					return
				}
				/*
					TODO
					1 分别遍历循环deltaModel.DeltaWithdraws和deltaModel.DeltaPublishs（定义也见https://github.com/cpusoft/goutil/blob/master/rrdputil/rrdpmodel.go)
					2 将对应的值赋值给PreceptRpkiDeltaWithdrawBase64 和 PreceptRpkiDeltaPublishBase64，注意rrdpType的值
					3 生成NewRrModel，前面各个参数参加前面urlRrModel，最后的txt用PreceptRpkiDeltaWithdrawBase64和PreceptRpkiDeltaPublishBase64的json值
					4 每20个NewRrModel放入c.RrModels = append(c.RrModels, urlRrModel)中，然后调用go addDnsRrs(&c)发送。注意每次新循环时，需要清空RrModels
				*/
				for _, preceptDeltaWithdraw := range deltaModel.DeltaWithdraws {
					preceptRpkiDeltaWithdrawBase64 := PreceptRpkiDeltaWithdrawBase64{
						RrdpType:  PRECEPT_RPKI_RRDP_TYPE_DELTA_WITHDRAW,
						NotifyUrl: preceptRpkiDomain.NotifyUrl,
						SessionId: preceptRpkiDomain.SessionId,
						Serial:    preceptRpkiDomain.MaxSerial,
						DeltaUrl:  deltaUrl,
						Url:       preceptDeltaWithdraw.Uri,
						Hash:      preceptDeltaWithdraw.Hash,
					}
					preceptRpkiDeltaWithdrawBase64Json := jsonutil.MarshalJson(preceptRpkiDeltaWithdrawBase64)
					belogs.Debug("receivePreceptRpki(): range DeltaWithdraws, notifyUrl:", preceptRpkiDomain.NotifyUrl,
						"  deltaUrl:", deltaUrl, "  preceptRpkiDeltaWithdrawBase64Json:", preceptRpkiDeltaWithdrawBase64Json)
					deltaWithdrawRrModel := rr.NewRrModel(origin, domain, dnsutil.DNS_TYPE_STR_TXT, dnsutil.DNS_CLASS_STR_IN,
						null.IntFrom(1000), preceptRpkiDeltaWithdrawBase64Json)
					belogs.Debug("receivePreceptRpki(): range DeltaWithdraws, get deltaWithdrawRrModel:", jsonutil.MarshalJson(deltaWithdrawRrModel))
					clientDnsRrModel.AddRrModel(deltaWithdrawRrModel)
					// if clientDnsRrModel.RrModels==**, will send and reset RrModels
					judgeSendClientDnsRrModel(clientDnsRrModel)
				}
				// if left clientDnsRrModel.RrModels, will send
				sendLeftClientDnsRrModel(clientDnsRrModel)

				for _, preceptDeltaPublish := range deltaModel.DeltaPublishs {
					preceptRpkiDeltaPublishBase64 := PreceptRpkiDeltaPublishBase64{
						RrdpType:  PRECEPT_RPKI_RRDP_TYPE_DELTA_PUBLISH,
						NotifyUrl: preceptRpkiDomain.NotifyUrl,
						SessionId: preceptRpkiDomain.SessionId,
						Serial:    preceptRpkiDomain.MaxSerial,
						DeltaUrl:  deltaUrl,
						Url:       preceptDeltaPublish.Uri,
						Hash:      preceptDeltaPublish.Hash,
						Base64:    preceptDeltaPublish.Base64,
					}
					preceptRpkiDeltaPublishBase64Json := jsonutil.MarshalJson(preceptRpkiDeltaPublishBase64)
					belogs.Debug("receivePreceptRpki(): range DeltaPublishs, notifyUrl:", preceptRpkiDomain.NotifyUrl,
						"  deltaUrl:", deltaUrl, "  preceptRpkiDeltaPublishBase64Json:", preceptRpkiDeltaPublishBase64Json)
					deltaPublishRrModel := rr.NewRrModel(origin, domain, dnsutil.DNS_TYPE_STR_TXT, dnsutil.DNS_CLASS_STR_IN,
						null.IntFrom(1000), preceptRpkiDeltaPublishBase64Json)
					belogs.Debug("receivePreceptRpki(): range DeltaPublishs, get deltaPublishRrModel:", jsonutil.MarshalJson(deltaPublishRrModel))
					clientDnsRrModel.AddRrModel(deltaPublishRrModel)
					// if clientDnsRrModel.RrModels==**, will send and reset RrModels
					judgeSendClientDnsRrModel(clientDnsRrModel)
				}
				// if left clientDnsRrModel.RrModels, will send
				sendLeftClientDnsRrModel(clientDnsRrModel)

				belogs.Debug("receivePreceptRpki(): get deltaModel, notifyUrl:", preceptRpkiDomain.NotifyUrl, "  deltaUrl:", deltaUrl,
					" deltaModel:", jsonutil.MarshalJson(deltaModel), "  time(s):", time.Since(deltaTime))
			}
		}

	}

	belogs.Info("receivePreceptRpki(): all done,  time(s):", time.Since(start))
	return
}

func cloneAndSendAndResetClientDnsRrModel(clientDnsRrModel *ClientDnsRrModel) {
	// send cloned clientDnsRrModel
	cSend := CloneClientDnsRrModel(clientDnsRrModel)
	belogs.Debug("cloneAndSendAndResetClientDnsRrModel():  send cSend:", jsonutil.MarshalJson(cSend))
	go addDnsRrs(cSend)
	// clear RrModels
	clientDnsRrModel.RrModels = make([]*rr.RrModel, 0)
}

func judgeSendClientDnsRrModel(clientDnsRrModel *ClientDnsRrModel) {
	rrModelsCountInOne := conf.Int("precept::rrModelsCountInOne")
	if len(clientDnsRrModel.RrModels) == rrModelsCountInOne {
		// send cloned clientDnsRrModel
		belogs.Debug("judgeSendClientDnsRrModel(): will send clientDnsRrModel:", jsonutil.MarshalJson(clientDnsRrModel))
		cloneAndSendAndResetClientDnsRrModel(clientDnsRrModel)
		belogs.Debug("judgeSendClientDnsRrModel():  after send clientDnsRrModel:", jsonutil.MarshalJson(clientDnsRrModel))
	}
}
func sendLeftClientDnsRrModel(clientDnsRrModel *ClientDnsRrModel) {
	if len(clientDnsRrModel.RrModels) > 0 {
		// send cloned clientDnsRrModel
		belogs.Debug("sendLeftClientDnsRrModel(): will send clientDnsRrModel:", jsonutil.MarshalJson(clientDnsRrModel))
		cloneAndSendAndResetClientDnsRrModel(clientDnsRrModel)
		belogs.Debug("sendLeftClientDnsRrModel(): after send clientDnsRrModel:", jsonutil.MarshalJson(clientDnsRrModel))
	}
}
