package dns

import (
	"errors"
	"strings"

	clientcache "dns-client-cache"
	connect "dns-connect"
	"dns-model/common"
	process "dns-process"
	dsomodel "dso-core/model"
	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/conf"
	"github.com/cpusoft/goutil/dnsutil"
	"github.com/cpusoft/goutil/jsonutil"
	"github.com/cpusoft/goutil/transportutil"
	querymodel "query-core/model"
	updatemodel "update-core/model"
)

var dnsClient *DnsClient

//////////////////////////////////
//

type DnsClient struct {
	// tcp/tls server and callback Func
	TcpClient           *transportutil.TcpClient
	dnsTcpClientProcess *DnsTcpClientProcess

	UdpClient           *transportutil.UdpClient
	dnsUdpClientProcess *DnsUdpClientProcess
}

// serverProtocol: "tcp" or "udp" or "tcp;udp" or "tls;udp"
func StartDnsClient(serverProtocol, serverHost, serverTcpPort, serverUdpPort string) (err error) {
	belogs.Info("StartDnsClient(): serverProtocol:", serverProtocol,
		"  serverHost:", serverHost,
		"  serverTcpPort:", serverTcpPort,
		"  serverUdpPort:", serverUdpPort)

	// no :=
	dnsClient = &DnsClient{}
	if (strings.Contains(serverProtocol, "tcp") || strings.Contains(serverProtocol, "tls")) &&
		len(serverTcpPort) > 0 {

		err = clientcache.InitCache()
		if err != nil {
			belogs.Error("StartDnsClient(): InitCache fail:", err)
			return err
		}

		tcpBusinessToConnMsg := make(chan transportutil.BusinessToConnMsg, 15)
		belogs.Debug("StartDnsClient(): tcpBusinessToConnMsg:", tcpBusinessToConnMsg)

		// process
		dnsClient.dnsTcpClientProcess = NewDnsTcpClientProcess(tcpBusinessToConnMsg)
		belogs.Debug("StartDnsClient(): NewDnsTcpClientProcess:", dnsClient.dnsTcpClientProcess)

		// tcpClient
		tcpClient, err := NewTcpClient(serverProtocol, dnsClient.dnsTcpClientProcess, tcpBusinessToConnMsg)
		if err != nil {
			belogs.Error("StartDnsClient(): NewTcpClient fail:", err)
			return err
		}
		dnsClient.TcpClient = tcpClient
		belogs.Debug("StartDnsClient(): tcpClient:", dnsClient.TcpClient)
		// set to global dnsClient
		if strings.Contains(serverProtocol, "tls") {
			err = dnsClient.TcpClient.StartTlsClient(serverHost + ":" + serverTcpPort)
		} else if strings.Contains(serverProtocol, "tcp") {
			err = dnsClient.TcpClient.StartTcpClient(serverHost + ":" + serverTcpPort)
		} else {
			return errors.New("protocols is not supported")
		}
		if err != nil {
			belogs.Error("StartDnsClient(): StartTlsClient or  StartTcpClient fail, serverProtocol:", serverProtocol,
				"  serverHost:", serverHost, "  serverTcpPort:", serverTcpPort, "  serverUdpPort:", serverUdpPort, err)
			return err
		}
		belogs.Info("StartDnsClient(): start tcpClient to serverHost:", serverHost, " serverTcpPort:", serverTcpPort)
	}
	if strings.Contains(serverProtocol, "udp") &&
		len(serverUdpPort) > 0 {
		udpBusinessToConnMsg := make(chan transportutil.BusinessToConnMsg, 15)
		belogs.Debug("StartDnsClient(): udpBusinessToConnMsg:", udpBusinessToConnMsg)

		// process
		dnsClient.dnsUdpClientProcess = NewDnsUdpClientProcess(udpBusinessToConnMsg)
		belogs.Debug("StartDnsClient(): NewDnsUdpClientProcess:", dnsClient.dnsUdpClientProcess)

		// udpClient
		udpClient, err := NewUdpClient(dnsClient.dnsUdpClientProcess, udpBusinessToConnMsg)
		if err != nil {
			belogs.Error("StartDnsClient(): NewUdpClient fail:", err)
			return err
		}
		dnsClient.UdpClient = udpClient
		belogs.Debug("StartDnsClient(): udpClient:", dnsClient.UdpClient)
		err = dnsClient.UdpClient.StartUdpClient(serverHost + ":" + serverUdpPort)
		if err != nil {
			belogs.Error("StartDnsClient(): StartUdpClient fail, serverProtocol:", serverProtocol,
				"  serverHost:", serverHost, "  serverTcpPort:", serverTcpPort, "  serverUdpPort:", serverUdpPort, err)
			return err
		}
		belogs.Info("StartDnsClient(): start udpClient to serverHost:", serverHost, " serverUdpPort:", serverUdpPort)
	}

	return nil

}

func NewTcpClient(serverProtocol string, dnsTcpClientProcess *DnsTcpClientProcess,
	tcpBusinessToConnMsg chan transportutil.BusinessToConnMsg) (c *transportutil.TcpClient, err error) {
	belogs.Debug("NewTcpClient(): serverProtocol:", serverProtocol)

	if strings.Contains(serverProtocol, "tls") {
		certsPath := conf.String("dns-client::programDir") + "/conf/cert/"
		caTlsRoot := conf.String("dns-client::caTlsRoot")
		clientTlsKey := conf.String("dns-client::clientTlsKey")
		clientTlsCrt := conf.String("dns-client::clientTlsCrt")
		receiveOnePacketLength := conf.Int("dns-client::tcptlsReceiveOnePacketLength")
		belogs.Debug("NewTcpClient():tls certsPath:", certsPath, " caTlsRoot:", caTlsRoot,
			" clientTlsKey:", clientTlsKey, " clientTlsCrt:", clientTlsCrt, "  receiveOnePacketLength:", receiveOnePacketLength)

		var err error
		c, err = transportutil.NewTlsClient(certsPath+caTlsRoot,
			certsPath+clientTlsCrt, certsPath+clientTlsKey,
			dnsTcpClientProcess, tcpBusinessToConnMsg, receiveOnePacketLength)
		if err != nil {
			belogs.Error("NewTcpClient(): NewTlsClient fail, certsPath:", certsPath,
				"   caTlsRoot:", caTlsRoot, "  clientTlsKey:", clientTlsKey,
				"   clientTlsCrt:", clientTlsCrt, "   tlsVerifyClient:", true,
				"   dnsTcpClientProcess:", dnsTcpClientProcess, "  receiveOnePacketLength:", receiveOnePacketLength,
				err)
			return nil, err
		}
		return c, nil
	} else if strings.Contains(serverProtocol, "tcp") {
		receiveOnePacketLength := conf.Int("dns-client::tcptlsReceiveOnePacketLength")
		belogs.Debug("NewTcpClient(): tcp  receiveOnePacketLength:", receiveOnePacketLength)
		c = transportutil.NewTcpClient(dnsTcpClientProcess, tcpBusinessToConnMsg, receiveOnePacketLength)
		return c, nil
	} else {
		return nil, errors.New("protocol is not supported")
	}

}
func NewUdpClient(dnsUdpClientProcess *DnsUdpClientProcess, udpBusinessToConnMsg chan transportutil.BusinessToConnMsg) (*transportutil.UdpClient, error) {
	receiveOnePacketLength := conf.Int("dns-client::udpReceiveOnePacketLength")
	belogs.Debug("NewUdpClient():udp receiveOnePacketLength:", receiveOnePacketLength)
	return transportutil.NewUdpClient(dnsUdpClientProcess, udpBusinessToConnMsg, receiveOnePacketLength), nil
}
func StopDnsClient() (err error) {
	if CheckDnsClientIsConnected() {
		belogs.Debug("StopDnsClient(): CheckDnsClientIsConnected, will send BUSINESS_TO_CONN_MSG_TYPE_CLIENT_CLOSE_CONNECT")
		businessToConnMsg := &transportutil.BusinessToConnMsg{
			BusinessToConnMsgType: transportutil.BUSINESS_TO_CONN_MSG_TYPE_CLIENT_CLOSE_CONNECT,
		}
		belogs.Debug("StopDnsClient(): CheckDnsClientIsConnected, businessToConnMsg:", jsonutil.MarshalJson(businessToConnMsg))
		dnsClient.TcpClient.SendAndReceiveMsg(businessToConnMsg)
	}
	dnsClient = nil
	return nil
}

func CheckDnsClientIsConnected() bool {
	if dnsClient != nil && dnsClient.TcpClient != nil {
		return true
	}
	return false
}

func SendTcpDnsModel(dnsModel common.DnsModel, needClientWaitForServerResponse bool) (responseDnsModel common.DnsModel, err error) {
	connToBusinessMsg, err := process.ClientSendTcpDnsModel(dnsClient.TcpClient, dnsModel, needClientWaitForServerResponse)
	if err != nil {
		belogs.Error("SendTcpDnsModel(): ClientSendTcpDnsModel fail:", err)
		return nil, err
	}
	belogs.Debug("SendTcpDnsModel(): ClientSendTcpDnsModel ok, connToBusinessMsg(may be nil):", jsonutil.MarshalJson(connToBusinessMsg))
	if connToBusinessMsg == nil {
		return nil, nil
	}
	return convertConnBusinessMsgToResponseDnsModel(connToBusinessMsg)
}

func SendUdpDnsModel(dnsModel common.DnsModel, needClientWaitForServerResponse bool) (responseDnsModel common.DnsModel, err error) {
	connToBusinessMsg, err := process.ClientSendUdpDnsModel(dnsClient.UdpClient, dnsModel, needClientWaitForServerResponse)
	if err != nil {
		belogs.Error("SendUdpDnsModel(): ClientSendUdpDnsModel fail:", err)
		return nil, err
	}
	belogs.Debug("SendUdpDnsModel(): ClientSendUdpDnsModel ok, connToBusinessMsg(may be nil):", jsonutil.MarshalJson(connToBusinessMsg))
	if connToBusinessMsg == nil {
		return nil, nil
	}
	return convertConnBusinessMsgToResponseDnsModel(connToBusinessMsg)
}

func convertConnBusinessMsgToResponseDnsModel(connToBusinessMsg *transportutil.ConnToBusinessMsg) (responseDnsModel common.DnsModel, err error) {
	opCode := dnsutil.DnsStrOpCodes[connToBusinessMsg.ConnToBusinessMsgType]

	switch opCode {
	case dnsutil.DNS_OPCODE_QUERY:
		/*
			receiveData := jsonutil.MarshalJson(connToBusinessMsg.ReceiveData)
			var responseQueryRrModel querymodel.QueryRrModel
			err = jsonutil.UnmarshalJson(receiveData, &responseQueryRrModel)
			if err != nil {
				belogs.Error("convertConnBusinessMsgToResponseDnsModel(): DNS_OPCODE_QUERY UnmarshalJson responseQueryRrModel fail, connToBusinessMsg.ReceiveData:", connToBusinessMsg.ReceiveData, err)
				return nil, err
			}
		*/
		responseQueryRrModel, ok := (connToBusinessMsg.ReceiveData).(*querymodel.QueryRrModel)
		if !ok {
			belogs.Error("convertConnBusinessMsgToResponseDnsModel(): DNS_OPCODE_QUERY ReceiveData to QueryRrModel fail:", jsonutil.MarshalJson(connToBusinessMsg.ReceiveData))
			return nil, errors.New("ReceiveData is not query type")
		}
		belogs.Info("convertConnBusinessMsgToResponseDnsModel(): DNS_OPCODE_QUERY responseQueryRrModel:", jsonutil.MarshalJson(responseQueryRrModel))
		return responseQueryRrModel, nil
	case dnsutil.DNS_OPCODE_UPDATE:
		// try convert to responseUpdateRrModel
		responseUpdateRrModel, ok := (connToBusinessMsg.ReceiveData).(*updatemodel.UpdateRrModel)
		if ok {
			belogs.Info("convertConnBusinessMsgToResponseDnsModel(): DNS_OPCODE_UPDATE responseUpdateRrModel:", jsonutil.MarshalJson(responseUpdateRrModel))
			return responseUpdateRrModel, nil
		}
		// if cannot convert, then using json
		receiveData := jsonutil.MarshalJson(connToBusinessMsg.ReceiveData)
		err = jsonutil.UnmarshalJson(receiveData, &responseUpdateRrModel)
		if err != nil {
			belogs.Error("convertConnBusinessMsgToResponseDnsModel(): DNS_OPCODE_UPDATE ReceiveData UnmarshalJson responseUpdateRrModel fail, receiveData:", receiveData, err)
			return nil, err
		}
		belogs.Info("convertConnBusinessMsgToResponseDnsModel(): DNS_OPCODE_UPDATE ReceiveData to responseUpdateRrModel:", jsonutil.MarshalJson(responseUpdateRrModel))
		return responseUpdateRrModel, nil

	case dnsutil.DNS_OPCODE_DSO:
		responseDsoModel, ok := (connToBusinessMsg.ReceiveData).(*dsomodel.DsoModel)
		if !ok {
			belogs.Error("convertConnBusinessMsgToResponseDnsModel(): DNS_OPCODE_UPDATE ReceiveData to dsoModel fail:", jsonutil.MarshalJson(connToBusinessMsg.ReceiveData))
			return nil, errors.New("ReceiveData is not dso type")
		}
		belogs.Info("convertConnBusinessMsgToResponseDnsModel(): DNS_OPCODE_DSO responseDsoModel:", jsonutil.MarshalJson(responseDsoModel))
		return responseDsoModel, nil
	}
	return nil, errors.New("opCode is not supported")

}

func GetTcpDnsConnect() *connect.DnsConnect {
	return dnsClient.dnsTcpClientProcess.DnsConnect
}
