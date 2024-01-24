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
	querymodel "query-core/model"
	updatemodel "update-core/model"
)

// receive bytes --> receive model
// return normal error(not dnsError)
func ParseBytesToDnsModel(receiveBytes []byte) (receiveDnsModel common.DnsModel, newOffsetFromStart uint16, err error) {
	start := time.Now()
	belogs.Debug("ParseBytesToDnsModel(): receiveBytes:", convert.PrintBytesOneLine(receiveBytes))
	if len(receiveBytes) < common.DNS_HEADER_LENGTH+common.DNS_COUNT_LENGTH {
		belogs.Error("ParseBytesToDnsModel(): len(receiveBytes) < DNS_HEADER_LENGTH+DNS_COUNT_LENGTH:", len(receiveBytes))
		return nil, 0, errors.New("receiveBytes is illegal DNS format")
	}
	belogs.Info("#接收到DNS消息(二进制格式):" + osutil.GetNewLineSep() + convert.PrintBytes(receiveBytes, 8))

	// header
	headerType, headerModel, err := common.ParseBytesToHeaderModel(receiveBytes[:common.DNS_HEADER_LENGTH])
	if err != nil {
		belogs.Error("ParseBytesToDnsModel(): ParseBytesToHeaderModel fail:", convert.PrintBytesOneLine(receiveBytes[:common.DNS_HEADER_LENGTH]))
		return nil, 0, errors.New("receiveBytes is illegal headerModel format")
	}
	offsetFromStart := uint16(common.DNS_HEADER_LENGTH)
	belogs.Debug("ParseBytesToDnsModel(): headerType:", headerType, "   headerModel:", jsonutil.MarshalJson(headerModel), "  offsetFromStart:", offsetFromStart)

	// count
	countModel, err := common.ParseBytesToCountModel(receiveBytes[offsetFromStart:offsetFromStart+common.DNS_COUNT_LENGTH], headerType)
	if err != nil {
		belogs.Error("ParseBytesToDnsModel(): ParseBytesToCountModel fail:", convert.PrintBytesOneLine(receiveBytes[:common.DNS_HEADER_LENGTH]))
		return nil, 0, errors.New("receiveBytes is illegal countModel format")
	}
	offsetFromStart += uint16(common.DNS_COUNT_LENGTH)
	belogs.Debug("ParseBytesToDnsModel(): headerType:", headerType, "   countModel:", jsonutil.MarshalJson(countModel), "  offsetFromStart:", offsetFromStart)

	// datamodel
	opCode := headerModel.GetOpCode()
	switch opCode {
	case dnsutil.DNS_OPCODE_QUERY:
		receiveDnsModel, newOffsetFromStart, err = querymodel.ParseBytesToQueryModel(headerModel,
			countModel, receiveBytes[offsetFromStart:], offsetFromStart)
	case dnsutil.DNS_OPCODE_UPDATE:
		receiveDnsModel, newOffsetFromStart, err = updatemodel.ParseBytesToUpdateModel(headerModel,
			countModel, receiveBytes[offsetFromStart:], offsetFromStart)
	case dnsutil.DNS_OPCODE_DSO:
		receiveDnsModel, newOffsetFromStart, err = dsomodel.ParseBytesToDsoModel(headerModel,
			countModel, receiveBytes[offsetFromStart:], offsetFromStart)
	default:
		belogs.Error("ParseBytesToDnsModel(): opCode fail:", opCode)
		return nil, 0, errors.New("not support DNS OPCODE")
	}
	if err != nil {
		belogs.Error("ParseBytesToDnsModel(): ParseBytesTo***Model fail, opCode:", opCode,
			"  receiveBytes[offsetFromStart:]: ", convert.PrintBytesOneLine(receiveBytes[offsetFromStart:]), "  offsetFromStart:", offsetFromStart)
		return nil, 0, err
	}
	belogs.Debug("ParseBytesToDnsModel(): ParseBytesToDnsModel , opCode:", opCode,
		"   receiveDnsModel:", jsonutil.MarshalJson(receiveDnsModel),
		"   receiveBytes[offsetFromStart:]: ", convert.PrintBytesOneLine(receiveBytes[offsetFromStart:]),
		"   newOffsetFromStart:", newOffsetFromStart,
		"   time(s):", time.Since(start))

	belogs.Info("#接收到DNS消息(json格式):" + osutil.GetNewLineSep() + jsonutil.MarshallJsonIndent(receiveDnsModel))
	belogs.Info("#接收到DNS类型: " + dnsutil.DnsIntOpCodes[opCode])
	return receiveDnsModel, newOffsetFromStart, nil
}
