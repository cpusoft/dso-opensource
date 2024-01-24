package main

import (
	"strings"
	"time"

	dnsclient "dns-client/dns"
	driveclient "dns-client/drive"
	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/conf"
	"github.com/cpusoft/goutil/ginserver"
	_ "github.com/cpusoft/goutil/logs"
	"github.com/gin-gonic/gin"
	"golang.org/x/sync/errgroup"
)

func startDnsClient() {
	start := time.Now()
	var g errgroup.Group
	err := startDnsTcpUdpClient(&g)
	if err != nil {
		belogs.Error("startDnsClient(): startDnsTcpUdpClient fail, end:", err)
		return
	}

	err = startDnsHttpClient(&g)
	if err != nil {
		belogs.Error("startDnsClient(): startDnsHttpClient fail, end:", err)
		return
	}

	belogs.Info("startDnsClient(): client started:", time.Since(start))

	if err := g.Wait(); err != nil {
		belogs.Error("startDnsClient(): client fail, will exit, err:", err)
	}
	belogs.Info("startDnsClient(): client tcp/http(s)  client end, time(s):", time.Since(start))

}

func startDnsTcpUdpClient(g *errgroup.Group) (err error) {
	start := time.Now()
	// should get server port
	serverProtocol := conf.String("dns-server::serverProtocol")
	serverHost := conf.String("dns-server::serverHost")
	tcpPort := conf.String("dns-server::serverTcpPort")
	tlsPort := conf.String("dns-server::serverTlsPort")
	udpPort := conf.String("dns-server::serverUdpPort")
	belogs.Debug("startDnsTcpUdpClient(): serverProtocol:", serverProtocol, "  serverHost:", serverHost,
		"  tcpPort:", tcpPort, "  tlsPort:", tlsPort, "  udpPort:", udpPort)

	g.Go(func() error {
		belogs.Debug("startDnsTcpUdpClient(): server run tcp/tls on :", serverProtocol, " tcpPort:", tcpPort,
			"  tlsPort:", tlsPort)
		var tcpTlsPort string
		if strings.Contains(serverProtocol, "tcp") {
			tcpTlsPort = tcpPort
		} else if strings.Contains(serverProtocol, "tls") {
			tcpTlsPort = tlsPort
		}
		// serverProtocol string, serverHost string, serverPort string, udpPort string
		err = dnsclient.StartDnsClient(serverProtocol, serverHost, tcpTlsPort, udpPort)
		if err != nil {
			belogs.Error("startDnsTcpUdpClient(): StartDnsClient fail, serverProtocol:", serverProtocol,
				"  tlsPort:", tlsPort, "  tcpPort:", tcpPort, "  udpPort:", udpPort, err)
		}
		belogs.Info("startDnsTcpUdpClient(): StartDnsClient ok, serverProtocol:", serverProtocol,
			"  tlsPort:", tlsPort, "  tcpPort:", tcpPort, "  udpPort:", udpPort, " time(s):", time.Since(start))
		return nil
	})
	belogs.Info("startDnsTcpUdpClient():tcp/tls/udp client will start, time(s):", time.Since(start))
	return nil
}

func startDnsHttpClient(g *errgroup.Group) (err error) {

	start := time.Now()

	certsPath := conf.String("dns-client::programDir") + "/conf/cert/"
	clientHttpPort := conf.String("dns-client::clientHttpPort")
	clientHttpsPort := conf.String("dns-client::clientHttpsPort")
	clientHttpsCrt := conf.String("dns-client::clientHttpsCrt")
	clientHttpsKey := conf.String("dns-client::clientHttpsKey")
	belogs.Info("startDnsHttpClient(): start client: certsPath:", certsPath,
		"   clientHttpPort:", clientHttpPort, "   clientHttpsPort:", clientHttpsPort,
		"   clientHttpsCrt:", clientHttpsCrt, "  clientHttpsKey:", clientHttpsKey)

	//gin.SetMode(gin.DebugMode)
	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()
	engine.Use(gin.Logger())
	engine.Use(gin.Recovery())

	// connect
	engine.POST("/client/creatednsconnect", driveclient.CreateDnsConnect)
	engine.POST("/client/closednsconnect", driveclient.CloseDnsConnect)

	// update/srp
	engine.POST("/client/adddnsrrs", driveclient.AddDnsRrs)
	engine.POST("/client/deldnsrrs", driveclient.DelDnsRrs)
	engine.POST("/client/digserverdnsrrs", driveclient.DigServerDnsRrs)
	engine.POST("/client/queryclientdnsrrs", driveclient.QueryClientDnsRrs)
	engine.POST("/client/queryclientalldnsrrs", driveclient.QueryClientAllDnsRrs)
	engine.POST("/client/clearclientalldnsrrs", driveclient.ClearClientAllDnsRrs)

	// dso
	engine.POST("/client/startkeepalive", driveclient.StartKeepalive)
	engine.POST("/client/subscribednsrr", driveclient.SubscribeDnsRr)
	engine.POST("/client/unsubscribednsrr", driveclient.UnsubscribeDnsRr)
	//engine.POST("/dsoclient/triggerreconfirm", dsoclient.TriggerReconfirm)

	// recept rpki
	engine.POST("/client/receiverreceptrpki", driveclient.ReceivePreceptRpki)

	if clientHttpPort != "" {
		g.Go(func() error {
			belogs.Info("startDnsHttpClient(): client run http on :", clientHttpPort)
			err := engine.Run(":" + clientHttpPort)
			if err != nil {
				belogs.Error("startDnsHttpClient(): http fail:", clientHttpPort, err)
			}
			return err
		})
	}

	if clientHttpsPort != "" {
		g.Go(func() error {
			belogs.Info("startDnsHttpClient(): client run https on :", clientHttpsPort, certsPath+clientHttpsCrt, certsPath+clientHttpsKey)
			err := ginserver.RunTLSEx(engine, ":"+clientHttpsPort, certsPath+clientHttpsCrt, certsPath+clientHttpsKey)
			if err != nil {
				belogs.Error("startDnsHttpClient(): https fail:", clientHttpsPort, err)
			}
			return err
		})
	}
	belogs.Info("startDnsHttpClient(): http(s) client will start, time(s):", time.Since(start))

	return
}
