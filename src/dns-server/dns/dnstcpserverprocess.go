package dns

import (
	"sync"
	"time"

	connect "dns-connect"
	"dns-model/message"
	pushmodel "dns-model/push"
	process "dns-process"
	dsomodel "dso-core/model"
	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/convert"
	"github.com/cpusoft/goutil/dnsutil"
	"github.com/cpusoft/goutil/jsonutil"
	"github.com/cpusoft/goutil/transportutil"
)

type DnsTcpServerProcess struct {
	dnsConnects       map[string]*connect.DnsConnect
	dnsConnectsMutex  sync.RWMutex
	businessToConnMsg chan transportutil.BusinessToConnMsg
	dnsToProcessMsg   *message.DnsToProcessMsg
}

func NewDnsTcpServerProcess(businessToConnMsg chan transportutil.BusinessToConnMsg) *DnsTcpServerProcess {
	c := &DnsTcpServerProcess{}
	c.dnsConnects = make(map[string]*connect.DnsConnect, 16)
	c.businessToConnMsg = businessToConnMsg

	dnsMsg := make(chan message.DnsMsg, 15)
	dnsToProcessMsg := message.NewDnsToProcessMsg(message.DNS_TRANSACT_SIDE_SERVER, dnsMsg)
	c.dnsToProcessMsg = dnsToProcessMsg
	go c.waitTcpToProcessDnsMsg()
	return c
}

func (c *DnsTcpServerProcess) OnConnectProcess(tcpConn *transportutil.TcpConn) {
	c.dnsConnectsMutex.Lock()
	defer c.dnsConnectsMutex.Unlock()
	connKey := transportutil.GetTcpConnKey(tcpConn)
	dnsConnect := connect.NewDnsTcpConnect("all", tcpConn, c.businessToConnMsg)
	c.dnsConnects[connKey] = dnsConnect
}

func (c *DnsTcpServerProcess) OnReceiveAndSendProcess(tcpConn *transportutil.TcpConn,
	receiveData []byte) (nextConnectPolicy int, leftData []byte, err error) {
	start := time.Now()
	c.dnsConnectsMutex.Lock()
	defer c.dnsConnectsMutex.Unlock()
	belogs.Debug("DnsTcpServerProcess.OnReceiveAndSendProcess(): server len(receiveData):", len(receiveData), "   receiveData:", convert.PrintBytesOneLine(receiveData))

	// parse []byte --> receive model
	// not need recombine
	receiveDnsModel, newOffsetFromStart, err := process.ParseBytesToDnsModel(receiveData)
	if err != nil {
		belogs.Error("DnsTcpServerProcess.OnReceiveAndSendProcess(): server ParseBytesToDnsModel fail: tcpConn:", tcpConn.RemoteAddr().String(),
			convert.PrintBytesOneLine(receiveData), err)

		/*shaodebug: it is not response now, will response .
		er := process.SendErrorDsoModel(tcpConn, messageId, err)
		if er != nil {
			belogs.Error("OnReceiveAndSendProcess(): server CallParseToModel SendErrorDsoModel fail: tcpConn:", tcpConn.RemoteAddr().String(),
				er)
			return transportutil.NEXT_CONNECT_POLICY_CLOSE_FORCIBLE, nil, err
		}
		return model.GetNextConnectPolicy(err), nil, err
		*/
		return transportutil.NEXT_CONNECT_POLICY_KEEP, nil, err

	}
	belogs.Info("DnsTcpServerProcess.OnReceiveAndSendProcess(): server ParseBytesToDnsModel, receiveDnsModel:", jsonutil.MarshalJson(receiveDnsModel),
		"  newOffsetFromStart:", newOffsetFromStart, "  time(s):", time.Since(start))

	// save session
	belogs.Debug("DnsTcpServerProcess.OnReceiveAndSendProcess(): c.dnsConnects:", jsonutil.MarshalJson(c.dnsConnects), c.dnsConnects)
	connKey := transportutil.GetTcpConnKey(tcpConn)
	dnsConnect, ok := c.dnsConnects[connKey]
	if !ok {
		belogs.Error("DnsTcpServerProcess.OnReceiveAndSendProcess(): server dnsConnects fail: tcpConn:", tcpConn.RemoteAddr().String(),
			convert.PrintBytesOneLine(receiveData), err)
		err = SendErrorTcpDnsModel(transportutil.GetTcpConnKey(tcpConn), dnsutil.DNS_QR_RESPONSE, err)
		if err != nil {
			belogs.Error("DnsTcpServerProcess.OnReceiveAndSendProcess(): server dnsConnects SendErrorTcpDnsModel fail: tcpConn:", tcpConn.RemoteAddr().String(),
				err)
			return transportutil.NEXT_CONNECT_POLICY_CLOSE_FORCIBLE, nil, err
		}
		return transportutil.NEXT_CONNECT_POLICY_KEEP, nil, err
	}
	belogs.Info("DnsTcpServerProcess.OnReceiveAndSendProcess(): server  connKey:", connKey, "  dnsConnect:", jsonutil.MarshalJson(dnsConnect),
		"  c.dnsConnects:", jsonutil.MarshalJson(c.dnsConnects),
		"  time(s):", time.Since(start))

	// receive model --> transact --> response model
	// dnsError
	responseDnsModel, err := process.TransactDnsModel(dnsConnect, receiveDnsModel, c.dnsToProcessMsg)
	if err != nil {
		belogs.Error("DnsTcpServerProcess.OnReceiveAndSendProcess(): server TransactDnsModel fail: tcpConn:", tcpConn.RemoteAddr().String(),
			convert.PrintBytesOneLine(receiveData), err)

		err = SendErrorTcpDnsModel(transportutil.GetTcpConnKey(tcpConn), dnsutil.DNS_QR_RESPONSE, err)
		if err != nil {
			belogs.Error("DnsTcpServerProcess.OnReceiveAndSendProcess(): server TransactDnsModel SendErrorTcpDnsModel fail: tcpConn:", tcpConn.RemoteAddr().String(),
				err)
			return transportutil.NEXT_CONNECT_POLICY_CLOSE_FORCIBLE, nil, err
		}
		return transportutil.NEXT_CONNECT_POLICY_KEEP, nil, err

	}
	belogs.Debug("DnsTcpServerProcess.OnReceiveAndSendProcess(): server TransactDnsModel, responseDnsModel:", jsonutil.MarshalJson(responseDnsModel),
		"   tcpConn:", tcpConn.RemoteAddr().String(), "  time(s):", time.Since(start))
	receiveMessageId := receiveDnsModel.GetHeaderModel().GetIdOrMessageId()
	if receiveMessageId == 0 {
		// not send response
		belogs.Info("DnsTcpServerProcess.OnReceiveAndSendProcess(): server SendDsoModel,  receiveMessageId is 0, not send response:", jsonutil.MarshalJson(responseDnsModel),
			"   tcpConn:", tcpConn.RemoteAddr().String(), "  time(s):", time.Since(start))
	} else {
		err = SendTcpDnsModel(transportutil.GetTcpConnKey(tcpConn), responseDnsModel)
		if err != nil {
			belogs.Error("DnsTcpServerProcess.OnReceiveAndSendProcess(): server SendDsoModel fail: tcpConn:", tcpConn.RemoteAddr().String(),
				err)
			return transportutil.NEXT_CONNECT_POLICY_CLOSE_FORCIBLE, nil, err
		}
		belogs.Info("DnsTcpServerProcess.OnReceiveAndSendProcess(): server SendDsoModel, havd send responseDnsModel:", jsonutil.MarshalJson(responseDnsModel),
			"   tcpConn:", tcpConn.RemoteAddr().String(), "  time(s):", time.Since(start))
	}
	// continue to receive next receiveData
	return transportutil.NEXT_CONNECT_POLICY_KEEP, receiveData[newOffsetFromStart:], nil
}

func (c *DnsTcpServerProcess) OnCloseProcess(tcpConn *transportutil.TcpConn) {

	start := time.Now()
	// process func OnClose
	belogs.Debug("DnsTcpServerProcess.OnCloseProcess(): server tcpserver tcpConn: ", tcpConn.RemoteAddr().String())

	// remove tcpConn from tcpConns
	c.dnsConnectsMutex.Lock()
	defer c.dnsConnectsMutex.Unlock()
	belogs.Debug("DnsTcpServerProcess.OnCloseProcess(): server tcpserver will close old dnsConnects, tcpConn: ", tcpConn.RemoteAddr().String(),
		"   old len(dsp.dnsConnects): ", len(c.dnsConnects), "  connKey:", transportutil.GetTcpConnKey(tcpConn))
	dnsConnect, ok := c.dnsConnects[transportutil.GetTcpConnKey(tcpConn)]
	if ok {
		dnsConnect.Close()
		delete(c.dnsConnects, transportutil.GetTcpConnKey(tcpConn))
	}
	belogs.Info("DnsTcpServerProcess.OnCloseProcess(): server tcpserver new len(c.dnsConnects): ", len(c.dnsConnects), "  time(s):", time.Since(start))

}

func (c *DnsTcpServerProcess) waitTcpToProcessDnsMsg() (err error) {
	belogs.Debug("DnsTcpServerProcess.waitTcpToProcessDnsMsg(): server start")
	for {
		// wait next waitTcpToProcessDnsMsg
		select {
		case dnsMsg := <-c.dnsToProcessMsg.DnsMsg:
			belogs.Debug("DnsTcpServerProcess.waitTcpToProcessDnsMsg(): server dnsMsg:", jsonutil.MarshalJson(dnsMsg))

			switch dnsMsg.DnsMsgType {
			case message.DNS_MSG_TYPE_DSO_SERVER_PUSH_TO_CLIENT:
				start := time.Now()
				belogs.Info("DnsTcpServerProcess.waitTcpToProcessDnsMsg(): server pushtoclient, DnsMsgType is DNS_MSG_TYPE_DSO_SERVER_PUSH_TO_CLIENT, dnsMsg.DnsMsgData:", dnsMsg.DnsMsgData)
				dnsMsgData, ok := (dnsMsg.DnsMsgData).(string)
				if !ok {
					belogs.Error("DnsTcpServerProcess.waitTcpToProcessDnsMsg(): server push send is not string:", jsonutil.MarshalJson(dnsMsg), " wait next dnsmsg")
					break
				}
				belogs.Debug("DnsTcpServerProcess.waitTcpToProcessDnsMsg(): server pushtoclient, dnsMsgData:", dnsMsgData)

				pushResultRrModels := make([]*pushmodel.PushResultRrModel, 0)
				err := jsonutil.UnmarshalJson(dnsMsgData, &pushResultRrModels)
				if err != nil {
					belogs.Error("DnsTcpServerProcess.waitTcpToProcessDnsMsg(): server pushtoclient, push send is not pushResultRrModels, fail, dnsMsgData:", dnsMsgData, err, " wait next dnsmsg")
					break
				}
				belogs.Debug("DnsTcpServerProcess.waitTcpToProcessDnsMsg(): server pushtoclient, pushResultRrModels:", jsonutil.MarshalJson(pushResultRrModels))
				for i := range pushResultRrModels {
					connKey := pushResultRrModels[i].ConnKey
					rrModels := pushResultRrModels[i].RrModels
					dnsModel, err := dsomodel.NewDsoModelWithPushTlvModel(rrModels)
					if err != nil {
						belogs.Error("DnsTcpServerProcess.waitTcpToProcessDnsMsg(): server pushtoclient, NewDsoModelWithPushTlvModel fail, rrModels:", jsonutil.MarshalJson(rrModels), err, " wait next dnsmsg")
						continue
					}
					belogs.Debug("DnsTcpServerProcess.waitTcpToProcessDnsMsg(): server pushtoclient, dnsModel:", jsonutil.MarshalJson(dnsModel))

					err = SendTcpDnsModel(connKey, dnsModel)
					if err != nil {
						belogs.Error("DnsTcpServerProcess.waitTcpToProcessDnsMsg(): server pushtoclient, SendDsoModel fail: connKey:", connKey, err)
						continue
					}
					belogs.Info("DnsTcpServerProcess.waitTcpToProcessDnsMsg(): server pushtoclient, SendDsoModel ok,  dnsModel:", jsonutil.MarshalJson(dnsModel),
						"  time(s):", time.Since(start))

				}
			}
		}
	}

}
