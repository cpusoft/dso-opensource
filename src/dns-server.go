package main

import (
	"fmt"
	"strings"
	"time"

	dnsserver "dns-server/dns"
	driveserver "dns-server/drive"
	sys "dns-sys"
	zonefile "dns-zonefile"
	push "dso-push"
	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/conf"
	"github.com/cpusoft/goutil/ginserver"
	_ "github.com/cpusoft/goutil/logs"
	"github.com/cpusoft/goutil/xormdb"
	"github.com/gin-gonic/gin"
	"golang.org/x/sync/errgroup"
)

func startDnsServer() {
	start := time.Now()

	err := xormdb.InitMySql()
	if err != nil {
		belogs.Error("startDnsServer(): start InitMySql failed:", err)
		fmt.Println("dns-server failed to start, ", err)
		return
	}
	defer xormdb.XormEngine.Close()
	xormdb.XormEngine.ShowSQL(true)

	var g errgroup.Group
	err = startDnsTcpUdpServer(&g)
	if err != nil {
		belogs.Error("startDnsServer(): startDnsTcpUdpServer fail, end:", err)
		return
	}

	err = startDnsHttpServer(&g)
	if err != nil {
		belogs.Error("startDnsServer(): startDnsHttpServer fail, end:", err)
		return
	}

	belogs.Info("startDnsServer(): server started, time(s):", time.Since(start))

	if err := g.Wait(); err != nil {
		belogs.Error("main(): server fail, will exit, err:", err)
	}
	belogs.Info("startDnsServer(): server tcp/http(s)  server end, time(s):", time.Since(start))

}

func startDnsTcpUdpServer(g *errgroup.Group) (err error) {

	start := time.Now()
	serverProtocol := conf.String("dns-server::serverProtocol")
	tcpPort := conf.String("dns-server::serverTcpPort")
	tlsPort := conf.String("dns-server::serverTlsPort")
	udpPort := conf.String("dns-server::serverUdpPort")
	belogs.Debug("startDnsTcpUdpServer(): serverProtocol:", serverProtocol,
		"  tcpPort:", tcpPort, "  tlsPort:", tlsPort, "  udpPort:", udpPort)

	g.Go(func() error {
		belogs.Info("startDnsTcpUdpServer(): server run tcp/tls on :", serverProtocol,
			" tcpPort:", tcpPort, "  tlsPort:", tlsPort, "  udpPort:", udpPort)
		var tcpTlsPort string
		if strings.Contains(serverProtocol, "tcp") {
			tcpTlsPort = tcpPort
		} else if strings.Contains(serverProtocol, "tls") {
			tcpTlsPort = tlsPort
		}
		err = dnsserver.StartDnsServer(serverProtocol, tcpTlsPort, udpPort)
		if err != nil {
			belogs.Error("startDnsTcpUdpServer(): startDnsTcpUdpServer fail:", serverProtocol, tcpPort, tlsPort, err)
		}
		return err
	})
	belogs.Info("startDnsTcpUdpServer():tcp/tls server will start, time(s):", time.Since(start))

	return nil
}

func startDnsHttpServer(g *errgroup.Group) (err error) {
	start := time.Now()

	certsPath := conf.String("dns-server::programDir") + "/conf/cert/"
	serverHttpPort := conf.String("dns-server::serverHttpPort")
	serverHttpsPort := conf.String("dns-server::serverHttpsPort")
	serverHttpsCrt := conf.String("dns-server::serverHttpsCrt")
	serverHttpsKey := conf.String("dns-server::serverHttpsKey")
	belogs.Info("startDnsHttpServer(): start server: certsPath:", certsPath,
		"   serverHttpPort:", serverHttpPort, "   serverHttpsPort:", serverHttpsPort,
		"   serverHttpsCrt:", serverHttpsCrt, "  serverHttpsKey:", serverHttpsKey)

	//gin.SetMode(gin.DebugMode)
	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()
	engine.Use(gin.Logger())
	engine.Use(gin.Recovery())

	// sys
	engine.POST("/sys/initreset", sys.InitReset)

	// server
	engine.POST("/server/activepushall", driveserver.ActivePushAll)

	// zonefile
	engine.POST("/zonefile/import", zonefile.ImportZoneFile)
	engine.POST("/zonefile/export", zonefile.ExportZoneFile)

	// push
	engine.POST("/push/subscribe", push.Subscribe)
	engine.POST("/push/unsubscribe", push.Unsubscribe)
	engine.POST("/push/delconn", push.DelConn)
	engine.POST("/push/queryrrmodelsshouldpush", push.QueryRrModelsShouldPush)
	engine.POST("/push/activepushall", push.ActivePushAll)
	engine.POST("/push/getallsubscribedrrs", push.GetAllSubscribedRrs)

	engine.POST("/server/queryserverdnsrrs", driveserver.QueryServerDnsRrs)
	engine.POST("/server/queryserveralldnsrrs", driveserver.QueryServerAllDnsRrs)

	engine.POST("/server/queryrpkirepos", driveserver.QueryRpkiRepos)

	if serverHttpPort != "" {
		g.Go(func() error {
			belogs.Info("startDnsHttpServer(): server run http on :", serverHttpPort)
			err := engine.Run(":" + serverHttpPort)
			if err != nil {
				belogs.Error("startDnsHttpServer(): http fail:", serverHttpPort, err)
			}
			return err
		})
	}

	if serverHttpsPort != "" {
		g.Go(func() error {
			belogs.Info("startDnsHttpServer(): server run https on :", serverHttpsPort, certsPath+serverHttpsCrt, certsPath+serverHttpsKey)
			err := ginserver.RunTLSEx(engine, ":"+serverHttpsPort, certsPath+serverHttpsCrt, certsPath+serverHttpsKey)
			if err != nil {
				belogs.Error("startDnsHttpServer(): https fail:", serverHttpsPort, err)
			}
			return err
		})
	}
	belogs.Info("startDnsHttpServer(): http(s) server will start, time(s):", time.Since(start))

	return
}
