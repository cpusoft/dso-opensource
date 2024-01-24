package process

import (
	"errors"
	"time"

	connect "dns-connect"
	"dns-model/common"
	"dns-model/message"
	dsotransact "dso-core/transact"
	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/dnsutil"
	"github.com/cpusoft/goutil/jsonutil"
	querytransact "query-core/transact"
	updatetransact "update-core/transact"
)

// receive model --> response model
// dnsToProcessMsg: message.DNS_TRANSACT_SIDE_SERVER/CLIENT/RPOXY
func TransactDnsModel(dnsConnect *connect.DnsConnect, receiveDnsModel common.DnsModel, dnsToProcessMsg *message.DnsToProcessMsg) (responseDnsModel common.DnsModel, err error) {
	start := time.Now()
	belogs.Debug("TransactDnsModel():dnsConnect:", jsonutil.MarshalJson(dnsConnect),
		"   receiveDnsModel:", jsonutil.MarshalJson(receiveDnsModel), "   dnsToProcessMsg:", jsonutil.MarshalJson(dnsToProcessMsg))

	opCode := receiveDnsModel.GetHeaderModel().GetOpCode()
	belogs.Debug("TransactDnsModel():opCode:", opCode, "   dnsToProcessMsg:", jsonutil.MarshalJson(dnsToProcessMsg))
	switch opCode {
	case dnsutil.DNS_OPCODE_QUERY:
		responseDnsModel, err = querytransact.PerformQueryTransact(receiveDnsModel, dnsToProcessMsg)
	case dnsutil.DNS_OPCODE_UPDATE:
		responseDnsModel, err = updatetransact.PerformUpdateTransact(receiveDnsModel, dnsToProcessMsg)
	case dnsutil.DNS_OPCODE_DSO:
		responseDnsModel, err = dsotransact.PerformDsoTransact(dnsConnect, receiveDnsModel, dnsToProcessMsg)
	default:
		belogs.Error("TransactDnsModel(): opCode fail:", opCode)
		return nil, errors.New("not support DNS OPCODE")
	}
	if err != nil {
		belogs.Error("TransactDnsModel(): Perform***Transact fail, opCode:", opCode,
			"  receiveDsoModel: ", jsonutil.MarshalJson(receiveDnsModel))
		return nil, err
	}
	belogs.Info("TransactDnsModel(): TransactDnsModel , opCode:", opCode,
		"  receiveDnsModel:", jsonutil.MarshalJson(receiveDnsModel),
		"  responseDnsModel: ", jsonutil.MarshalJson(responseDnsModel), "   time(s):", time.Since(start))

	return responseDnsModel, nil

}
