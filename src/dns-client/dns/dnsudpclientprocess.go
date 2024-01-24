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

type DnsUdpClientProcess struct {
	dnsConnects       *connect.DnsConnect
	businessToConnMsg chan transportutil.BusinessToConnMsg
	dnsToProcessMsg   *message.DnsToProcessMsg
}

func NewDnsUdpClientProcess(businessToConnMsg chan transportutil.BusinessToConnMsg) *DnsUdpClientProcess {
	c := &DnsUdpClientProcess{}
	c.businessToConnMsg = businessToConnMsg
	dnsMsg := make(chan message.DnsMsg, 15)
	dnsToProcessMsg := message.NewDnsToProcessMsg(message.DNS_TRANSACT_SIDE_CLIENT, dnsMsg)
	c.dnsToProcessMsg = dnsToProcessMsg
	go c.waitUdpToProcessDnsMsg()
	return c
}
func (c *DnsUdpClientProcess) OnReceiveProcess(udpConn *transportutil.UdpConn, receiveData []byte) (connToBusinessMsg *transportutil.ConnToBusinessMsg, err error) {
	start := time.Now()
	belogs.Debug("DnsUdpClientProcess.OnReceiveProcess(): client len(receiveData):", len(receiveData), "   receiveData:", convert.PrintBytesOneLine(receiveData))

	// parse []byte --> receive model
	// not need recombine
	receiveDnsModel, newOffsetFromStart, err := process.ParseBytesToDnsModel(receiveData)
	if err != nil {
		belogs.Error("DnsUdpClientProcess.OnReceiveProcess(): client ParseBytesToDnsModel fail: udpConn:", udpConn,
			convert.PrintBytesOneLine(receiveData), err)
		return nil, err
	}
	belogs.Info("DnsUdpClientProcess.OnReceiveProcess(): client ParseBytesToDnsModel, receiveDnsModel:", jsonutil.MarshalJson(receiveDnsModel),
		"  newOffsetFromStart:", newOffsetFromStart, "  time(s):", time.Since(start))

	// udp --> dnsconnect
	dnsConnect := connect.NewDnsUdpConnect("all", udpConn, nil, c.businessToConnMsg)
	belogs.Info("DnsUdpClientProcess.OnReceiveProcess(): dnsConnect:", dnsConnect)

	// receive model --> transact --> end this read
	// dnsError
	responseDnsModel, err := process.TransactDnsModel(dnsConnect, receiveDnsModel, c.dnsToProcessMsg)
	if err != nil {
		belogs.Error("DnsUdpClientProcess.OnReceiveProcess(): client TransactDnsModel fail: udpConn:", udpConn,
			convert.PrintBytesOneLine(receiveData), err)
		return nil, err
	}
	belogs.Debug("DnsUdpClientProcess.OnReceiveProcess(): responseDnsModel:", jsonutil.MarshalJson(responseDnsModel))

	receiveMessageId := receiveDnsModel.GetHeaderModel().GetIdOrMessageId()
	isActiveSendFromServer := (receiveMessageId == 0) // when id is 0, is active from server
	opCode := responseDnsModel.GetHeaderModel().GetOpCode()
	msgType := dnsutil.DnsIntOpCodes[opCode]
	belogs.Debug("DnsUdpClientProcess.OnReceiveProcess(): receiveMessageId:", receiveMessageId, "  isActiveSendFromServer:", isActiveSendFromServer,
		" opCode:", opCode, "  msgType:", msgType)
	connToBusinessMsg = &transportutil.ConnToBusinessMsg{
		IsActiveSendFromServer: isActiveSendFromServer,
		ConnToBusinessMsgType:  msgType,
		ReceiveData:            responseDnsModel, //jsonutil.MarshalJson(responseDnsModel),
	}
	belogs.Info("DnsUdpClientProcess.OnReceiveProcess(): client TransactDnsModel, connToBusinessMsg:", jsonutil.MarshalJson(connToBusinessMsg),
		"   udpConn:", udpConn, "  time(s):", time.Since(start))
	return connToBusinessMsg, nil
}

func (c *DnsUdpClientProcess) waitUdpToProcessDnsMsg() (err error) {
	belogs.Debug("DnsUdpClientProcess.waitUdpToProcessDnsMsg():client start")
	for {
		select {
		case dnsMsg := <-c.dnsToProcessMsg.DnsMsg:
			belogs.Info("DnsUdpClientProcess.waitUdpToProcessDnsMsg(): client dnsMsg:", jsonutil.MarshalJson(dnsMsg))
		}
	}

}
