package message

import "github.com/cpusoft/goutil/dnsutil"

const (
	DNS_TRANSACT_SIDE_SERVER = "server"
	DNS_TRANSACT_SIDE_CLIENT = "client"

	DNS_MSG_TYPE_DSO_SERVER_PUSH_TO_CLIENT = "dsoServerPushToClient"
	DNS_MSG_TYPE_DSO_CLIENT_PUSH           = "dsoClientPush"

	DNS_MSG_TYPE_OP_CODE_QUERY  = "opCodeQuery"
	DNS_MSG_TYPE_OP_CODE_UPDATE = "opCodeUpdate"
	DNS_MSG_TYPE_OP_CODE_DSO    = "opCodeDso"
)

var DnsMsgTypeIntOpCodes map[uint8]string = map[uint8]string{
	dnsutil.DNS_OPCODE_QUERY:  DNS_MSG_TYPE_OP_CODE_QUERY,
	dnsutil.DNS_OPCODE_UPDATE: DNS_MSG_TYPE_OP_CODE_UPDATE,
	dnsutil.DNS_OPCODE_DSO:    DNS_MSG_TYPE_OP_CODE_DSO,
}

type DnsMsg struct {
	// DNS_MSG_TYPE_***
	DnsMsgType string      `json:"dnsMsgType"`
	DnsMsgData interface{} `json:"dnsMsgData"`
}

func NewDnsMsg(dnsMsgType string, dnsMsgData interface{}) *DnsMsg {
	c := &DnsMsg{
		DnsMsgType: dnsMsgType,
		DnsMsgData: dnsMsgData,
	}
	return c
}

type DnsToProcessMsg struct {
	// DNS_TRANSACT_SIDE_SERVER/DNS_TRANSACT_SIDE_CLIENT
	DnsTransactSide string `json:"dnsTransactSide"`
	// outside make/close
	DnsMsg chan DnsMsg `json:"-"`
}

func NewDnsToProcessMsg(dnsTransactSide string, dnsMsg chan DnsMsg) *DnsToProcessMsg {
	c := &DnsToProcessMsg{
		DnsTransactSide: dnsTransactSide,
		DnsMsg:          dnsMsg,
	}
	return c
}

func (c *DnsToProcessMsg) SendDnsToProcessMsgChan(dnsMsg *DnsMsg) {
	c.DnsMsg <- *dnsMsg
}

// from process, such as 'receive data, will send to tcp/udp client'
// used in client only
type DnsFromProcessMsg struct {
	ReceiveDnsMsg bool `json:"receiveDnsMsg"`
	// outside make/close
	DnsMsg chan DnsMsg `json:"-"`
}

// default false
func NewDnsFromProcessMsg() *DnsFromProcessMsg {
	c := &DnsFromProcessMsg{
		ReceiveDnsMsg: false,
	}
	return c
}
