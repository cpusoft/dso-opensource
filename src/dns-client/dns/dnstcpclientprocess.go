package dns

import (
	"time"

	connect "dns-connect"
	"dns-model/message"
	process "dns-process"
	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/convert"
	"github.com/cpusoft/goutil/dnsutil"
	"github.com/cpusoft/goutil/jsonutil"
	"github.com/cpusoft/goutil/transportutil"
)

type DnsTcpClientProcess struct {
	DnsConnect        *connect.DnsConnect
	businessToConnMsg chan transportutil.BusinessToConnMsg
	dnsToProcessMsg   *message.DnsToProcessMsg
}

func NewDnsTcpClientProcess(businessToConnMsg chan transportutil.BusinessToConnMsg) *DnsTcpClientProcess {
	c := &DnsTcpClientProcess{}
	c.businessToConnMsg = businessToConnMsg
	dnsMsg := make(chan message.DnsMsg, 15)
	dnsToProcessMsg := message.NewDnsToProcessMsg(message.DNS_TRANSACT_SIDE_CLIENT, dnsMsg)
	c.dnsToProcessMsg = dnsToProcessMsg
	go c.waitTcpToProcessDnsMsg()
	return c
}

func (c *DnsTcpClientProcess) OnConnectProcess(tcpConn *transportutil.TcpConn) {
	c.DnsConnect = connect.NewDnsTcpConnect("all", tcpConn, c.businessToConnMsg)
	return
}

func (c *DnsTcpClientProcess) OnCloseProcess(tcpConn *transportutil.TcpConn) {
	c.DnsConnect.Close()
	return
}

func (c *DnsTcpClientProcess) OnReceiveProcess(tcpConn *transportutil.TcpConn,
	receiveData []byte) (nextRwPolicy int, leftData []byte, connToBusinessMsg *transportutil.ConnToBusinessMsg, err error) {
	start := time.Now()
	belogs.Debug("DnsTcpClientProcess.OnReceiveProcess(): client len(receiveData):", len(receiveData), "   receiveData:", convert.PrintBytesOneLine(receiveData))

	// parse []byte --> receive model
	// not need recombine
	receiveDnsModel, newOffsetFromStart, err := process.ParseBytesToDnsModel(receiveData)
	if err != nil {
		belogs.Error("DnsTcpClientProcess.OnReceiveProcess(): client ParseBytesToDnsModel fail: tcpConn:", tcpConn.RemoteAddr().String(),
			convert.PrintBytesOneLine(receiveData), err)

		/*shaodebug: it is not response now, will response .
		er := process.SendErrorDsoModel(tcpConn, messageId, err)
		if er != nil {
			belogs.Error("OnReceiveProcess(): server CallParseToModel SendErrorDsoModel fail: tcpConn:", tcpConn.RemoteAddr().String(),
				er)
			return transportutil.NEXT_CONNECT_POLICY_CLOSE_FORCIBLE, nil, err
		}
		return model.GetNextConnectPolicy(err), nil, err
		*/
		return transportutil.NEXT_CONNECT_POLICY_CLOSE_FORCIBLE, nil, nil, err
	}
	belogs.Info("DnsTcpClientProcess.OnReceiveProcess(): client ParseBytesToDnsModel, receiveDnsModel:", jsonutil.MarshalJson(receiveDnsModel),
		"  newOffsetFromStart:", newOffsetFromStart, "  time(s):", time.Since(start))

	// save session and convert dnsConnect

	// receive model --> transact --> end this read
	// dnsError
	dnsConnect := c.DnsConnect
	responseDnsModel, err := process.TransactDnsModel(dnsConnect, receiveDnsModel, c.dnsToProcessMsg)
	if err != nil {
		belogs.Error("DnsTcpClientProcess.OnReceiveProcess(): client TransactDnsModel fail: tcpConn:", tcpConn.RemoteAddr().String(),
			convert.PrintBytesOneLine(receiveData), err)
		return transportutil.NEXT_CONNECT_POLICY_CLOSE_FORCIBLE, nil, nil, err

	}
	belogs.Info("DnsTcpClientProcess.OnReceiveProcess(): client TransactDnsModel, responseDnsModel:", jsonutil.MarshalJson(responseDnsModel),
		"   tcpConn:", tcpConn.RemoteAddr().String(), "  time(s):", time.Since(start))

	receiveMessageId := receiveDnsModel.GetHeaderModel().GetIdOrMessageId()
	isActiveSendFromServer := (receiveMessageId == 0) // when id is 0, is active from server
	opCode := responseDnsModel.GetHeaderModel().GetOpCode()
	msgType := dnsutil.DnsIntOpCodes[opCode]
	belogs.Debug("DnsTcpClientProcess.OnReceiveProcess():receiveMessageId:", receiveMessageId, "  isActiveSendFromServer:", isActiveSendFromServer,
		" opCode:", opCode, "  msgType:", msgType)
	connToBusinessMsg = &transportutil.ConnToBusinessMsg{
		IsActiveSendFromServer: isActiveSendFromServer,
		ConnToBusinessMsgType:  msgType,
		ReceiveData:            responseDnsModel, // jsonutil.MarshalJson(responseDnsModel),
	}
	belogs.Info("DnsTcpClientProcess.OnReceiveProcess(): client TransactDnsModel, connToBusinessMsg:", jsonutil.MarshalJson(connToBusinessMsg),
		"  time(s):", time.Since(start))
	// continue to receive next receiveData
	return transportutil.NEXT_RW_POLICY_WAIT_READ, leftData, connToBusinessMsg, nil
}

func (c *DnsTcpClientProcess) waitTcpToProcessDnsMsg() (err error) {
	belogs.Debug("DnsTcpClientProcess.waitTcpToProcessDnsMsg():client start")
	for {
		// wait next waitTcpToProcessDnsMsg
		select {
		case dnsMsg := <-c.dnsToProcessMsg.DnsMsg:
			belogs.Info("DnsTcpClientProcess.waitTcpToProcessDnsMsg(): client dnsMsg:", jsonutil.MarshalJson(dnsMsg))

			switch dnsMsg.DnsMsgType {
			case message.DNS_MSG_TYPE_DSO_CLIENT_PUSH:
				belogs.Info("DnsTcpClientProcess.waitTcpToProcessDnsMsg(): client DnsMsgType is DNS_MSG_TYPE_DSO_CLIENT_PUSH")
			}
		}
	}

}
