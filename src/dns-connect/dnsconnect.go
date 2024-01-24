package connect

import (
	"errors"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/conf"
	"github.com/cpusoft/goutil/dnsutil"
	"github.com/cpusoft/goutil/transportutil"
)

type DnsConnect struct {
	// all;query;update;dso
	SupportType string `json:"supportType"`

	// tcp/tls
	TcpConn *transportutil.TcpConn `json:"tcpConn"`
	// udp
	UdpConn       *transportutil.UdpConn `json:"udpConn"`
	clientUdpAddr *net.UDPAddr

	// close outside
	businessToConnMsg chan transportutil.BusinessToConnMsg `json:"-"`

	// DSO
	LastSendDsoType      uint16       `json:"lastSendDsoType"`
	LastIsUnidirectional bool         `json:"lastIsUnidirectional"`
	InactivityTimeout    uint32       `json:"inactivityTimeout"`
	KeepaliveInterval    uint32       `json:"keepaliveInterval"`
	SessionTimeoutTicker *time.Ticker `json:"sessionTimeoutTimer"`
	canKeepSession       bool         `json:"canKeepSession"`
	sessionState         string       `json:"sessionState"`
	sessionMutex         sync.RWMutex `json:"-"`
}

func NewDnsTcpConnect(supportType string, tcpConn *transportutil.TcpConn,
	businessToConnMsg chan transportutil.BusinessToConnMsg) *DnsConnect {
	c := &DnsConnect{
		SupportType: supportType,

		TcpConn:           tcpConn,
		businessToConnMsg: businessToConnMsg,

		InactivityTimeout: uint32(conf.Int("keepalive::inactivityTimeout")),
		KeepaliveInterval: uint32(conf.Int("keepalive::keepaliveInterval")),

		sessionState:   dnsutil.DSO_SESSION_STATE_CONNECTED_SESSIONLESS,
		canKeepSession: false,
	}
	return c
}
func NewDnsUdpConnect(supportType string, udpConn *transportutil.UdpConn, clientUdpAddr *net.UDPAddr,
	businessToConnMsg chan transportutil.BusinessToConnMsg) *DnsConnect {
	c := &DnsConnect{
		SupportType: supportType,

		UdpConn:           udpConn,
		clientUdpAddr:     clientUdpAddr,
		businessToConnMsg: businessToConnMsg,
	}
	return c
}
func (c *DnsConnect) EstablishedSession() error {
	if c.SupportDso() {
		c.sessionMutex.Lock()
		defer c.sessionMutex.Unlock()
		c.sessionState = dnsutil.DSO_SESSION_STATE_ESTABLISHED_SESSION
		c.canKeepSession = true
		belogs.Debug("DnsConnect.EstablishedSession(): set established_session to sessionState:", c.sessionState)
		return nil
	} else {
		belogs.Error("DnsConnect.EstablishedSession(): not support dso:", c.SupportType)
		return errors.New("this connect is not support dso for EstablishedSession, it is " + c.SupportType)
	}
}

func (c *DnsConnect) UnestablishedSession() error {
	if c.SupportDso() {
		c.sessionMutex.Lock()
		defer c.sessionMutex.Unlock()
		c.sessionState = dnsutil.DSO_SESSION_STATE_CONNECTED_SESSIONLESS
		c.canKeepSession = false
		return nil
	} else {
		belogs.Error("DnsConnect.UnestablishedSession(): not support dso:", c.SupportType)
		return errors.New("this connect is not support dso for UnestablishedSession, it is " + c.SupportType)
	}
}

// return current canKeepSession and set new canKeepSession
func (c *DnsConnect) GetCanKeepSession() (bool, error) {
	if c.SupportDso() {
		c.sessionMutex.RLock()
		defer c.sessionMutex.RUnlock()
		return c.canKeepSession, nil
	} else {
		belogs.Error("DnsConnect.GetCanKeepSession(): not support dso:", c.SupportType)
		return false, errors.New("this connect is not support dso for GetCanKeepSession, it is " + c.SupportType)
	}
}
func (c *DnsConnect) SetCanKeepSession(k bool) error {
	if c.SupportDso() {
		c.sessionMutex.Lock()
		defer c.sessionMutex.Unlock()
		c.canKeepSession = k
		return nil
	} else {
		belogs.Error("DnsConnect.GetCanKeepSession(): not support dso:", c.SupportType)
		return errors.New("this connect is not support dso for SetCanKeepSession, it is " + c.SupportType)
	}

}
func (c *DnsConnect) GetSessionState() (string, error) {
	if c.SupportDso() {
		c.sessionMutex.RLock()
		defer c.sessionMutex.RUnlock()
		belogs.Debug("DnsConnect.GetSessionState(): get sessionState:", c.sessionState)
		return c.sessionState, nil
	} else {
		belogs.Error("GetSessionState(): not support dso:", c.SupportType)
		return "", errors.New("this connect is not support dso for GetSessionState, it is " + c.SupportType)
	}
}

func (c *DnsConnect) Close() {
	if c.SupportDso() {
		c.sessionMutex.Lock()
		defer c.sessionMutex.Unlock()
		c.canKeepSession = false
		c.sessionState = dnsutil.DSO_SESSION_STATE_CONNECTED_SESSIONLESS
		belogs.Debug("DnsConnect.Close(): set connected_sessionless to sessionState:", c.sessionState)
		if c.SessionTimeoutTicker != nil {
			c.SessionTimeoutTicker.Stop()
		}
	} else {
		// do nothing
	}
}

func (c *DnsConnect) SendBusinessToConnMsg(businessToConnMsg transportutil.BusinessToConnMsg) {
	c.businessToConnMsg <- businessToConnMsg
}

func (c *DnsConnect) SupportQuery() bool {
	belogs.Debug("DnsConnect.SupportQuery():SupportType:", c.SupportType)
	if c.TcpConn != nil || c.UdpConn != nil {
		if strings.Contains(c.SupportType, "all") || strings.Contains(c.SupportType, "query") {
			return true
		}
	}
	return false
}
func (c *DnsConnect) SupportUpdate() bool {
	belogs.Debug("DnsConnect.SupportUpdate():SupportType:", c.SupportType)
	if c.TcpConn != nil {
		if strings.Contains(c.SupportType, "all") || strings.Contains(c.SupportType, "update") {
			return true
		}
	}
	return false
}

func (c *DnsConnect) SupportDso() bool {
	belogs.Debug("DnsConnect.SupportDso():SupportType:", c.SupportType)
	if c.TcpConn != nil {
		if strings.Contains(c.SupportType, "all") || strings.Contains(c.SupportType, "dso") {
			return true
		}
	}
	return false
}
