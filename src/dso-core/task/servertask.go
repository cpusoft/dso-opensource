package task

import (
	"time"

	connect "dns-connect"
	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/jsonutil"
	"github.com/cpusoft/goutil/transportutil"
)

// will call go CheckSessionTimeoutInServer
func CheckSessionTimeoutInServer(dnsConnect *connect.DnsConnect) {

	// rfc8490: twice the current inactivity timeout value
	belogs.Debug("CheckSessionTimeoutInServer():dnsConnect.InactivityTimeout:", dnsConnect.InactivityTimeout)
	dnsConnect.SessionTimeoutTicker = time.NewTicker(time.Duration(dnsConnect.InactivityTimeout*2) * time.Second) // start ticker
	belogs.Info("CheckSessionTimeoutInServer(): will check session:", transportutil.GetTcpConnKey(dnsConnect.TcpConn),
		"  ever 2*InactivityTimeout(s):", 2*dnsConnect.InactivityTimeout)

	for {
		select {
		case t := <-dnsConnect.SessionTimeoutTicker.C:
			belogs.Debug("CheckSessionTimeoutInServer(): SessionTimeoutTicker: t:", t)
			canKeepSession, err := dnsConnect.GetCanKeepSession()
			if err == nil && canKeepSession {
				//c.SessionTimeoutTicker.Reset(time.Second * c.InactivityTimeout * 2)
				belogs.Debug("CheckSessionTimeoutInServer(): canKeepSession is true , 2*InactivityTimeout(s):", 2*dnsConnect.InactivityTimeout)
				// reset it to false, only receive data can set it to true
				dnsConnect.SetCanKeepSession(false)
			} else {
				// will stop connect
				if err != nil {
					belogs.Error("CheckSessionTimeoutInServer(): GetCanKeepSession, fail:", err)
				}
				// will close connection
				// stop ticker
				belogs.Debug("CheckSessionTimeoutInServer(): canKeepSession is false")
				dnsConnect.SessionTimeoutTicker.Stop()
				businessToConnMsg := transportutil.BusinessToConnMsg{
					BusinessToConnMsgType: transportutil.BUSINESS_TO_CONN_MSG_TYPE_SERVER_CLOSE_ONE_CONNECT_FORCIBLE,
					ServerConnKey:         transportutil.GetTcpConnKey(dnsConnect.TcpConn),
				}
				belogs.Info("CheckSessionTimeoutInServer(): canKeepSession is false so timeout, 2*InactivityTimeout(s):", 2*dnsConnect.InactivityTimeout,
					"   will send businessToConnMsg to close this connect:", jsonutil.MarshalJson(businessToConnMsg))
				dnsConnect.SendBusinessToConnMsg(businessToConnMsg)

				return
			}

		}
	}
}
