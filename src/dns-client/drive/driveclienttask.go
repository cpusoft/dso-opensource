package drive

import (
	"time"

	connect "dns-connect"
	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/transportutil"
)

// will call go KeepSessionInClient
func keepSessionInClient(dnsConnect *connect.DnsConnect) {

	// rfc8490:  keepaliveInterval
	dnsConnect.SessionTimeoutTicker = time.NewTicker(time.Duration(dnsConnect.KeepaliveInterval) * time.Second) // start ticker
	belogs.Info("keepSessionInClient(): will send keepalive to keep session:", transportutil.GetTcpConnKey(dnsConnect.TcpConn),
		"  ever KeepaliveInterval(s):", dnsConnect.KeepaliveInterval)

	for {
		select {
		case t := <-dnsConnect.SessionTimeoutTicker.C:
			belogs.Debug("keepSessionInClient(): inactivityTimeout:", dnsConnect.InactivityTimeout,
				" keepaliveInterval:", dnsConnect.KeepaliveInterval, " t:", t)
			keepaliveImpl(dnsConnect.InactivityTimeout, dnsConnect.KeepaliveInterval)

		}
	}
}
