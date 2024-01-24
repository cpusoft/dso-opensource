package process

import (
	"errors"
	"time"

	"dns-model/common"
	dsomodel "dso-core/model"
	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/convert"
	"github.com/cpusoft/goutil/dnsutil"
	"github.com/cpusoft/goutil/jsonutil"
	"github.com/cpusoft/goutil/osutil"
	"github.com/cpusoft/goutil/transportutil"
	querymodel "query-core/model"
	updatemodel "update-core/model"
)

// dnsModel --> bytes
func ServerSendTcpDnsModel(ts *transportutil.TcpServer, serverConnKey string, dnsModel common.DnsModel) (err error) {
	start := time.Now()
	belogs.Debug("ServerSendTcpDnsModel():serverConnKey:", serverConnKey, "  dnsModel:", jsonutil.MarshalJson(dnsModel))

	sendBytes := dnsModel.Bytes()
	belogs.Info("#服务器发送TCP/TLS消息(二进制格式):" + osutil.GetNewLineSep() + convert.PrintBytes(sendBytes, 8))
	belogs.Info("#服务器发送TCP/TLS消息(json格式):" + osutil.GetNewLineSep() + jsonutil.MarshallJsonIndent(dnsModel))
	businessToConnMsg := &transportutil.BusinessToConnMsg{
		BusinessToConnMsgType: transportutil.BUSINESS_TO_CONN_MSG_TYPE_COMMON_SEND_DATA,
		SendData:              sendBytes,
		ServerConnKey:         serverConnKey,
	}
	ts.SendBusinessToConnMsg(businessToConnMsg)
	belogs.Debug("ServerSendTcpDnsModel(): businessToConnMsg:", jsonutil.MarshalJson(businessToConnMsg),
		"   time(s):", time.Since(start))
	return nil
}
func ServerSendUdpDnsModel(us *transportutil.UdpServer, serverConnKey string, dnsModel common.DnsModel) (err error) {
	start := time.Now()
	belogs.Debug("ServerSendUdpDnsModel():serverConnKey:", serverConnKey, "  dnsModel:", jsonutil.MarshalJson(dnsModel))

	sendBytes := dnsModel.Bytes()
	belogs.Info("#服务器发送UDP消息(二进制格式):" + osutil.GetNewLineSep() + convert.PrintBytes(sendBytes, 8))
	belogs.Info("#服务器发送UDP消息(json格式):" + osutil.GetNewLineSep() + jsonutil.MarshallJsonIndent(dnsModel))
	businessToConnMsg := &transportutil.BusinessToConnMsg{
		BusinessToConnMsgType: transportutil.BUSINESS_TO_CONN_MSG_TYPE_COMMON_SEND_DATA,
		SendData:              sendBytes,
		ServerConnKey:         serverConnKey,
	}
	us.SendBusinessToConnMsg(businessToConnMsg)
	belogs.Debug("ServerSendUdpDnsModel(): businessToConnMsg:", jsonutil.MarshalJson(businessToConnMsg),
		"   time(s):", time.Since(start))
	return nil
}

// dnsModel --> bytes
func ClientSendTcpDnsModel(tc *transportutil.TcpClient, dnsModel common.DnsModel, needClientWaitForServerResponse bool) (connToBusinessMsg *transportutil.ConnToBusinessMsg, err error) {
	start := time.Now()
	sendBytes := dnsModel.Bytes()
	belogs.Debug("ClientSendTcpDnsModel(): dnsModel:", jsonutil.MarshalJson(dnsModel),
		"   sendBytes:", convert.PrintBytesOneLine(sendBytes), " needClientWaitForServerResponse:", needClientWaitForServerResponse)
	belogs.Info("#客户端往服务器发送DNS消息(json格式):" + osutil.GetNewLineSep() + jsonutil.MarshallJsonIndent(dnsModel))
	belogs.Info("#客户端往服务器发送DNS消息(二进制格式):" + osutil.GetNewLineSep() + convert.PrintBytes(sendBytes, 8))
	businessToConnMsg := &transportutil.BusinessToConnMsg{
		BusinessToConnMsgType:           transportutil.BUSINESS_TO_CONN_MSG_TYPE_COMMON_SEND_DATA,
		SendData:                        sendBytes,
		NeedClientWaitForServerResponse: needClientWaitForServerResponse,
	}
	connToBusinessMsg, err = tc.SendAndReceiveMsg(businessToConnMsg)
	if err != nil {
		belogs.Error("ClientSendTcpDnsModel(): SendAndReceiveMsg fail, businessToConnMsg:", jsonutil.MarshalJson(businessToConnMsg), err)
		return nil, err
	}
	belogs.Info("ClientSendTcpDnsModel(): businessToConnMsg:", jsonutil.MarshalJson(businessToConnMsg),
		"   time(s):", time.Since(start))
	return connToBusinessMsg, nil
}

func ClientSendUdpDnsModel(uc *transportutil.UdpClient, dnsModel common.DnsModel, needClientWaitForServerResponse bool) (connToBusinessMsg *transportutil.ConnToBusinessMsg, err error) {
	start := time.Now()
	sendBytes := dnsModel.Bytes()
	belogs.Debug("ClientSendUdpDnsModel(): dnsModel:", jsonutil.MarshalJson(dnsModel),
		"   sendBytes:", convert.PrintBytesOneLine(sendBytes), "  needClientWaitForServerResponse:", needClientWaitForServerResponse)
	belogs.Info("#客户端往服务器发送DNS消息(json格式):" + osutil.GetNewLineSep() + jsonutil.MarshallJsonIndent(dnsModel))
	belogs.Info("#客户端往服务器发送DNS消息(二进制格式):" + osutil.GetNewLineSep() + convert.PrintBytes(sendBytes, 8))
	businessToConnMsg := &transportutil.BusinessToConnMsg{
		BusinessToConnMsgType:           transportutil.BUSINESS_TO_CONN_MSG_TYPE_COMMON_SEND_AND_RECEIVE_DATA,
		SendData:                        sendBytes,
		NeedClientWaitForServerResponse: needClientWaitForServerResponse,
	}
	connToBusinessMsg, err = uc.SendAndReceiveMsg(businessToConnMsg)
	if err != nil {
		belogs.Error("ClientSendUdpDnsModel(): SendAndReceiveMsg fail, businessToConnMsg:", jsonutil.MarshalJson(businessToConnMsg), err)
		return nil, err
	}
	belogs.Info("ClientSendUdpDnsModel(): SendAndReceiveMsg ok, connToBusinessMsg:", jsonutil.MarshalJson(connToBusinessMsg),
		"   time(s):", time.Since(start))
	return connToBusinessMsg, nil
}
func ServerSendErrorTcpDnsModel(ts *transportutil.TcpServer, serverConnKey string, qr uint8, err1 error) (err error) {
	belogs.Debug("SendErrorDnsModel():  serverConnKey:", serverConnKey, "  dnsError:", jsonutil.MarshalJson(err1))

	id := dnsutil.GetDnsErrorId(err1)
	opCode := dnsutil.GetDnsErrorOpCode(err1)
	rCode := dnsutil.GetDnsErrorRCode(err1)
	belogs.Debug("SendErrorDnsModel(): id:", id, "  opCode:", opCode, "  rCode:", rCode, "  qr:", qr)
	switch opCode {
	case dnsutil.DNS_OPCODE_QUERY:
		queryModel, err := querymodel.NewQueryModelByParameters(id, qr, 0, 0, 0, 0, rCode)
		if err != nil {
			belogs.Error("SendErrorDnsModel(): NewQueryModelByParameters fail:", err)
			return err
		}
		return ServerSendTcpDnsModel(ts, serverConnKey, queryModel)
	case dnsutil.DNS_OPCODE_UPDATE:
		updateModel, err := updatemodel.NewUpdateModelByParameters(id, qr, rCode)
		if err != nil {
			belogs.Error("SendErrorDnsModel(): NewUpdateModelByParameters fail:", err)
			return err
		}
		return ServerSendTcpDnsModel(ts, serverConnKey, updateModel)
	case dnsutil.DNS_OPCODE_DSO:
		dsoModel, err := dsomodel.NewDsoModelByParameters(id, qr, rCode)
		if err != nil {
			belogs.Error("SendErrorDnsModel(): NewDsoModelByParameters fail:", err)
			return err
		}
		return ServerSendTcpDnsModel(ts, serverConnKey, dsoModel)
	default:
		belogs.Error("SendErrorDnsModel(): not support opCode: fail:", opCode)
		return errors.New("SendErrorDnsModel(): not support opCode")
	}

}
