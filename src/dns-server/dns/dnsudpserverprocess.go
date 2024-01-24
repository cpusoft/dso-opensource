package dns

import (
	"net"
	"time"

	connect "dns-connect"
	"dns-model/message"
	process "dns-process"
	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/convert"
	"github.com/cpusoft/goutil/jsonutil"
	"github.com/cpusoft/goutil/transportutil"
)

type DnsUdpServerProcess struct {
	dnsConnects       *connect.DnsConnect
	businessToConnMsg chan transportutil.BusinessToConnMsg
	dnsToProcessMsg   *message.DnsToProcessMsg
}

func NewDnsUdpServerProcess(businessToConnMsg chan transportutil.BusinessToConnMsg) *DnsUdpServerProcess {
	c := &DnsUdpServerProcess{}
	c.businessToConnMsg = businessToConnMsg
	dnsMsg := make(chan message.DnsMsg, 15)
	dnsToProcessMsg := message.NewDnsToProcessMsg(message.DNS_TRANSACT_SIDE_SERVER, dnsMsg)
	c.dnsToProcessMsg = dnsToProcessMsg
	go c.waitUdpToProcessDnsMsg()

	return c
}

func (c *DnsUdpServerProcess) OnReceiveAndSendProcess(udpConn *transportutil.UdpConn, clientUdpAddr *net.UDPAddr, receiveData []byte) (err error) {
	start := time.Now()
	belogs.Debug("DnsUdpServerProcess.OnReceiveAndSendProcess():", convert.PrintBytesOneLine(receiveData))
	// parse []byte --> receive model
	// not need recombine
	receiveDnsModel, newOffsetFromStart, err := process.ParseBytesToDnsModel(receiveData)
	if err != nil {
		belogs.Error("DnsUdpServerProcess.OnReceiveAndSendProcess(): server ParseBytesToDnsModel fail: udpConn:", clientUdpAddr,
			convert.PrintBytesOneLine(receiveData), err)
		return err
	}
	belogs.Info("DnsUdpServerProcess.OnReceiveAndSendProcess(): server ParseBytesToDnsModel, receiveDnsModel:", jsonutil.MarshalJson(receiveDnsModel),
		"  newOffsetFromStart:", newOffsetFromStart, "  time(s):", time.Since(start))

	// udp --> dnsconnect
	dnsConnect := connect.NewDnsUdpConnect("all", udpConn, clientUdpAddr, c.businessToConnMsg)
	belogs.Info("DnsUdpServerProcess.OnReceiveAndSendProcess(): dnsConnect:", dnsConnect)

	// receive model --> transact --> response model
	// dnsError
	responseDnsModel, err := process.TransactDnsModel(dnsConnect, receiveDnsModel, c.dnsToProcessMsg)
	if err != nil {
		belogs.Error("DnsUdpServerProcess.OnReceiveAndSendProcess(): server TransactDnsModel fail: clientUdpAddr:", clientUdpAddr,
			convert.PrintBytesOneLine(receiveData), err)
		return err
	}
	belogs.Debug("DnsUdpServerProcess.OnReceiveAndSendProcess(): server TransactDnsModel, responseDnsModel:", jsonutil.MarshalJson(responseDnsModel),
		"   clientUdpAddr:", clientUdpAddr, "  time(s):", time.Since(start))
	receiveMessageId := receiveDnsModel.GetHeaderModel().GetIdOrMessageId()
	if receiveMessageId == 0 {
		// not send response
		belogs.Info("DnsUdpServerProcess.OnReceiveAndSendProcess(): server SendDsoModel,  receiveMessageId is 0, not send response:", jsonutil.MarshalJson(responseDnsModel),
			"   clientUdpAddr:", clientUdpAddr, "  time(s):", time.Since(start))
	} else {
		err = SendUdpDnsModel(transportutil.GetUdpAddrKey(clientUdpAddr), responseDnsModel)
		if err != nil {
			belogs.Error("DnsUdpServerProcess.OnReceiveAndSendProcess(): server SendDsoModel fail: clientUdpAddr:", clientUdpAddr,
				err)
			return err
		}
		belogs.Info("DnsUdpServerProcess.OnReceiveAndSendProcess(): server SendDsoModel, have send responseDnsModel:", jsonutil.MarshalJson(responseDnsModel),
			"   clientUdpAddr:", clientUdpAddr, "  time(s):", time.Since(start))
	}
	return nil
}
func (c *DnsUdpServerProcess) waitUdpToProcessDnsMsg() (err error) {
	belogs.Debug("DnsUdpServerProcess.waitUdpToProcessDnsMsg(): server start")
	for {
		// wait next waitUdpToProcessDnsMsg
		select {
		case dnsMsg := <-c.dnsToProcessMsg.DnsMsg:
			belogs.Debug("DnsUdpServerProcess.waitUdpToProcessDnsMsg(): server dnsMsg:", jsonutil.MarshalJson(dnsMsg))
		}
	}

}
