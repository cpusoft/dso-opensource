package dns

import (
	"errors"
	"strings"

	"dns-model/common"
	process "dns-process"
	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/conf"
	"github.com/cpusoft/goutil/transportutil"
)

//////////////////////////////////
// DnsServer
var dnsServer *DnsServer

type DnsServer struct {
	// tcp/tls server and callback Func
	TcpServer           *transportutil.TcpServer
	dnsTcpServerProcess *DnsTcpServerProcess

	UdpServer           *transportutil.UdpServer
	dnsUdpServerProcess *DnsUdpServerProcess
}

// serverProtocol: "tcp" or "udp" or "tcp;udp" or "tls;udp"
func StartDnsServer(serverProtocol, serverTcpPort, serverUdpPort string) (err error) {
	belogs.Debug("StartDnsServer(): serverProtocol:", serverProtocol,
		"   serverTcpPort:", serverTcpPort, "  serverUdpPort:", serverUdpPort)

	// no :=
	dnsServer = &DnsServer{}

	// set to global dnsServer
	belogs.Debug("StartDnsServer(): serverProtocol:", serverProtocol)
	if (strings.Contains(serverProtocol, "tcp") || strings.Contains(serverProtocol, "tls")) &&
		len(serverTcpPort) > 0 {
		// msg
		tcpBusinessToConnMsg := make(chan transportutil.BusinessToConnMsg, 15)
		belogs.Debug("StartDnsServer(): tcpBusinessToConnMsg:", tcpBusinessToConnMsg)

		// process
		dnsServer.dnsTcpServerProcess = NewDnsTcpServerProcess(tcpBusinessToConnMsg)
		belogs.Debug("StartDnsServer(): dnsTcpServerProcess:", dnsServer.dnsTcpServerProcess)

		// tcpServer
		tcpServer, err := NewTcpServer(serverProtocol, dnsServer.dnsTcpServerProcess, tcpBusinessToConnMsg)
		if err != nil {
			belogs.Error("StartDnsServer(): NewTcpServer fail:", err)
			return err
		}
		dnsServer.TcpServer = tcpServer
		belogs.Info("StartDnsServer(): will  StartTlsServer or StartTcpServer ", dnsServer.TcpServer, " on serverTcpPort:", serverTcpPort)

		if strings.Contains(serverProtocol, "tls") {
			go dnsServer.TcpServer.StartTlsServer(serverTcpPort)
		} else if strings.Contains(serverProtocol, "tcp") {
			go dnsServer.TcpServer.StartTcpServer(serverTcpPort)
		}
	}
	if strings.Contains(serverProtocol, "udp") &&
		len(serverUdpPort) > 0 {

		// msg
		udpBusinessToConnMsg := make(chan transportutil.BusinessToConnMsg, 15)
		belogs.Debug("StartDnsServer(): udpBusinessToConnMsg:", udpBusinessToConnMsg)

		dnsServer.dnsUdpServerProcess = NewDnsUdpServerProcess(udpBusinessToConnMsg)
		belogs.Debug("StartDnsServer(): dnsUdpServerProcess:", dnsServer.dnsUdpServerProcess)

		udpServer, err := NewUdpServer(dnsServer.dnsUdpServerProcess, udpBusinessToConnMsg)
		if err != nil {
			belogs.Error("StartDnsServer(): NewUdpServer fail:", err)
			return err
		}
		dnsServer.UdpServer = udpServer
		belogs.Info("StartDnsServer(): will  StartUdpServer ", dnsServer.UdpServer, " on ", serverUdpPort)
		go dnsServer.UdpServer.StartUdpServer(serverUdpPort)
	}

	return nil
}

func NewTcpServer(serverProtocol string, dnsTcpServerProcess *DnsTcpServerProcess, tcpBusinessToConnMsg chan transportutil.BusinessToConnMsg) (*transportutil.TcpServer, error) {

	belogs.Debug("NewTcpServer(): serverProtocol:", serverProtocol)
	if strings.Contains(serverProtocol, "tls") {
		certsPath := conf.String("dns-server::programDir") + "/conf/cert/"
		caTlsRoot := conf.String("dns-server::caTlsRoot")
		serverTlsKey := conf.String("dns-server::serverTlsKey")
		serverTlsCrt := conf.String("dns-server::serverTlsCrt")
		receiveOnePacketLength := conf.Int("dns-server::tcptlsReceiveOnePacketLength")

		belogs.Debug("NewTcpServer(): certsPath:", certsPath, " caTlsRoot:", caTlsRoot, " serverTlsKey:", serverTlsKey,
			" serverTlsCrt:", serverTlsCrt, "  receiveOnePacketLength:", receiveOnePacketLength)

		var err error
		// no verify client cert
		c, err := transportutil.NewTlsServer(certsPath+caTlsRoot,
			certsPath+serverTlsCrt, certsPath+serverTlsKey,
			false, dnsTcpServerProcess, tcpBusinessToConnMsg, receiveOnePacketLength)
		if err != nil {
			belogs.Error("NewTcpServer(): NewTlsServer fail, certsPath:", certsPath,
				"   caTlsRoot:", caTlsRoot, "  serverTlsKey:", serverTlsKey,
				"   serverTlsCrt:", serverTlsCrt, "   tlsVerifyClient:", true,
				"   dnsTcpServerProcess:", dnsTcpServerProcess, " receiveOnePacketLength:", receiveOnePacketLength, err)
			return nil, err
		}
		return c, nil
	} else if strings.Contains(serverProtocol, "tcp") {
		receiveOnePacketLength := conf.Int("dns-server::tcptlsReceiveOnePacketLength")
		belogs.Debug("NewTcpServer(): tcp receiveOnePacketLength:", receiveOnePacketLength)
		c := transportutil.NewTcpServer(dnsTcpServerProcess, tcpBusinessToConnMsg, receiveOnePacketLength)
		return c, nil
	} else {
		return nil, errors.New("not support protocol")
	}
}

func NewUdpServer(dnsUdpServerProcess *DnsUdpServerProcess, udpBusinessToConnMsg chan transportutil.BusinessToConnMsg) (*transportutil.UdpServer, error) {
	receiveOnePacketLength := conf.Int("dns-server::udpReceiveOnePacketLength")
	belogs.Debug("NewUdpServer():udp receiveOnePacketLength:", receiveOnePacketLength)
	return transportutil.NewUdpServer(dnsUdpServerProcess, udpBusinessToConnMsg, receiveOnePacketLength), nil
}
func SendTcpDnsModel(serverConnKey string, dnsModel common.DnsModel) error {
	return process.ServerSendTcpDnsModel(dnsServer.TcpServer, serverConnKey, dnsModel)
}
func SendUdpDnsModel(serverConnKey string, dnsModel common.DnsModel) error {
	return process.ServerSendUdpDnsModel(dnsServer.UdpServer, serverConnKey, dnsModel)
}
func SendErrorTcpDnsModel(serverConnKey string, qr uint8, err1 error) error {
	return process.ServerSendErrorTcpDnsModel(dnsServer.TcpServer, serverConnKey, qr, err1)
}
