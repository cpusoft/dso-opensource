package transact

import (
	"dns-model/common"
	"dns-model/message"
	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/dnsutil"
	"github.com/cpusoft/goutil/jsonutil"
	"github.com/cpusoft/goutil/transportutil"
	updatemodel "update-core/model"
)

// err: dnserror
func PerformUpdateTransact(receiveDnsModel common.DnsModel,
	dnsToProcessMsg *message.DnsToProcessMsg) (responseDnsModel common.DnsModel, err error) {
	belogs.Debug("PerformUpdateTransact(): receiveDnsModel:", jsonutil.MarshalJson(receiveDnsModel), "   dnsToProcessMsg:", jsonutil.MarshalJson(dnsToProcessMsg))
	receiveUpdateModel, ok := (receiveDnsModel).(*updatemodel.UpdateModel)
	if !ok {
		belogs.Error("PerformUpdateTransact():receiveDnsModel to receiveUpdateModel, fail:", jsonutil.MarshalJson(receiveDnsModel))
		return nil, dnsutil.NewDnsError("fail to convert model type",
			0, dnsutil.DNS_OPCODE_UPDATE, dnsutil.DNS_RCODE_SERVFAIL, transportutil.NEXT_CONNECT_POLICY_KEEP)
	}
	belogs.Debug("PerformUpdateTransact(): receiveDnsModel to receiveUpdateModel,  receiveUpdateModel:", jsonutil.MarshalJson(receiveUpdateModel))

	id := receiveUpdateModel.GetHeaderModel().GetIdOrMessageId()
	// header transact
	responseUpdateModel, err := performHeaderTransact(receiveUpdateModel, dnsToProcessMsg)
	if err != nil {
		belogs.Error("PerformUpdateTransact(): performHeaderTransact fail, receiveUpdateModel:", jsonutil.MarshalJson(receiveUpdateModel), "   dnsToProcessMsg:", jsonutil.MarshalJson(dnsToProcessMsg), err)
		return nil, err
	}
	belogs.Debug("PerformUpdateTransact(): performHeaderTransact,  responseUpdateModel:", jsonutil.MarshalJson(*responseUpdateModel))

	// zone transact
	err = performZoneSectionTransact(id, receiveUpdateModel.UpdateDataModel.ZoneModel, dnsToProcessMsg)
	if err != nil {
		belogs.Error("PerformUpdateTransact(): performZoneSectionTransact fail, receiveUpdateModel:",
			jsonutil.MarshalJson(receiveUpdateModel), "   dnsToProcessMsg:", jsonutil.MarshalJson(dnsToProcessMsg), err)
		return nil, err
	}

	// prerequiste transact
	err = performPrerequisiteSectionTransact(receiveUpdateModel, dnsToProcessMsg)
	if err != nil {
		belogs.Error("PerformUpdateTransact(): performPrerequisiteSectionTransact fail, receiveUpdateModel:",
			jsonutil.MarshalJson(receiveUpdateModel), "   dnsToProcessMsg:", jsonutil.MarshalJson(dnsToProcessMsg), err)
		return nil, err
	}

	// update transact
	err = performUpdateSectionTransact(receiveUpdateModel, dnsToProcessMsg)
	if err != nil {
		belogs.Error("PerformUpdateTransact(): performUpdateSectionTransact fail, receiveUpdateModel:",
			jsonutil.MarshalJson(receiveUpdateModel), "   dnsToProcessMsg:", jsonutil.MarshalJson(dnsToProcessMsg), err)
		return nil, err
	}
	belogs.Info("PerformUpdateTransact(): ok,  responseUpdateModel:", jsonutil.MarshalJson(*responseUpdateModel))
	return responseUpdateModel, nil
}

// err: dnserror
func performHeaderTransact(receiveUpdateModel *updatemodel.UpdateModel, dnsToProcessMsg *message.DnsToProcessMsg) (responseUpdateModel *updatemodel.UpdateModel, err error) {
	belogs.Debug("performHeaderTransact(): receiveUpdateModel:", jsonutil.MarshalJson(receiveUpdateModel), "   dnsToProcessMsg:", jsonutil.MarshalJson(dnsToProcessMsg))

	id := receiveUpdateModel.HeaderForUpdateModel.Id
	qr := uint8(dnsutil.DNS_QR_RESPONSE)
	rCode := uint8(dnsutil.DNS_RCODE_NOERROR)
	responseUpdateModel, _ = updatemodel.NewUpdateModelByParameters(id, qr, rCode)
	belogs.Debug("performHeaderTransact():responseUpdateModel:", jsonutil.MarshalJson(responseUpdateModel))

	// get updatemodel header (have no tlv)
	belogs.Debug("performHeaderTransact():dnsToProcessMsg:", jsonutil.MarshalJson(dnsToProcessMsg))
	switch dnsToProcessMsg.DnsTransactSide {
	case message.DNS_TRANSACT_SIDE_SERVER:
		err = performHeaderTransactInServer(receiveUpdateModel, responseUpdateModel)
	case message.DNS_TRANSACT_SIDE_CLIENT:
		err = performHeaderTransactInClient(receiveUpdateModel, responseUpdateModel)
	default:
		belogs.Error("performHeaderTransact(): dnsTransactSide for header is not supported, fail:", dnsToProcessMsg.DnsTransactSide)
		return nil, dnsutil.NewDnsError("dnsTransactSide for header is not supported",
			id, dnsutil.DNS_OPCODE_UPDATE, dnsutil.DNS_RCODE_NOTIMP, transportutil.NEXT_CONNECT_POLICY_KEEP)
	}
	if err != nil {
		belogs.Error("performHeaderTransact(): performDsoHeaderTransactIn*** fail:, receiveUpdateModel:", jsonutil.MarshalJson(receiveUpdateModel), "   responseUpdateModel:", jsonutil.MarshalJson(responseUpdateModel),
			"   dnsToProcessMsg:", jsonutil.MarshalJson(dnsToProcessMsg), err)
		return nil, err
	}
	belogs.Debug("performHeaderTransact():ok responseUpdateModel:", jsonutil.MarshalJson(responseUpdateModel))
	return responseUpdateModel, nil
}

// err: dnserror
func performHeaderTransactInServer(receiveUpdateModel *updatemodel.UpdateModel, responseUpdateModel *updatemodel.UpdateModel) (err error) {
	id := receiveUpdateModel.GetHeaderModel().GetIdOrMessageId()
	if receiveUpdateModel.CountZPUAModel.ZoCount == 0 ||
		receiveUpdateModel.CountZPUAModel.PrCount == 0 ||
		receiveUpdateModel.CountZPUAModel.UpCount == 0 {
		belogs.Error("performHeaderTransactInServer(): receiveUpdateModel.CountZPUAModel have 0 count,",
			"    receiveUpdateModel.CountZPUAModel:", jsonutil.MarshalJson(receiveUpdateModel.CountZPUAModel))
		return dnsutil.NewDnsError("receiveUpdateModel.CountZPUAModel have 0 count",
			id,
			dnsutil.DNS_OPCODE_UPDATE,
			dnsutil.DNS_RCODE_FORMERR, transportutil.NEXT_CONNECT_POLICY_KEEP)
	}
	if receiveUpdateModel.CountZPUAModel.ZoCount != 1 ||
		receiveUpdateModel.UpdateDataModel.ZoneModel == nil {
		belogs.Error("performHeaderTransactInServer(): ZoCount not equal to 1 or ZoneModel is nil,",
			"    ZoCount:", receiveUpdateModel.CountZPUAModel.ZoCount, "  ZoneModel:", receiveUpdateModel.UpdateDataModel.ZoneModel)
		return dnsutil.NewDnsError("ZoCount not equal to 1 or ZoneModel is nil",
			id,
			dnsutil.DNS_OPCODE_UPDATE,
			dnsutil.DNS_RCODE_FORMERR, transportutil.NEXT_CONNECT_POLICY_KEEP)
	}
	if receiveUpdateModel.CountZPUAModel.PrCount != uint16(len(receiveUpdateModel.UpdateDataModel.PrerequisiteModels)) {
		belogs.Error("performHeaderTransactInServer(): PrCount not equal to len(PrerequisiteModels),",
			"    PrCount:", receiveUpdateModel.CountZPUAModel.PrCount, "  PrerequisiteModels:", jsonutil.MarshalJson(receiveUpdateModel.UpdateDataModel.PrerequisiteModels))
		return dnsutil.NewDnsError("PrCount not equal to len(PrerequisiteModels)",
			id,
			dnsutil.DNS_OPCODE_UPDATE,
			dnsutil.DNS_RCODE_FORMERR, transportutil.NEXT_CONNECT_POLICY_KEEP)
	}

	if receiveUpdateModel.CountZPUAModel.UpCount != uint16(len(receiveUpdateModel.UpdateDataModel.UpdateModels)) {
		belogs.Error("performHeaderTransactInServer(): UpCount not equal to len(UpdateModels),",
			"    UpCount:", receiveUpdateModel.CountZPUAModel.UpCount, "  UpdateModels:", jsonutil.MarshalJson(receiveUpdateModel.UpdateDataModel.UpdateModels))
		return dnsutil.NewDnsError("UpCount not equal to len(UpdateModels)",
			id,
			dnsutil.DNS_OPCODE_UPDATE,
			dnsutil.DNS_RCODE_FORMERR, transportutil.NEXT_CONNECT_POLICY_KEEP)
	}

	if receiveUpdateModel.CountZPUAModel.AdCount != uint16(len(receiveUpdateModel.UpdateDataModel.AdditionalDataModels)) {
		belogs.Error("performHeaderTransactInServer(): AdCount not equal to len(AdditionalDataModels),",
			"    AdCount:", receiveUpdateModel.CountZPUAModel.AdCount, "  AdditionalDataModels:", jsonutil.MarshalJson(receiveUpdateModel.UpdateDataModel.AdditionalDataModels))
		return dnsutil.NewDnsError("AdCount not equal to len(AdditionalDataModels)",
			id,
			dnsutil.DNS_OPCODE_UPDATE,
			dnsutil.DNS_RCODE_FORMERR, transportutil.NEXT_CONNECT_POLICY_KEEP)
	}
	return nil
}

func performHeaderTransactInClient(receiveUpdateModel *updatemodel.UpdateModel, responseUpdateModel *updatemodel.UpdateModel) (err error) {
	// ok ,if all are 0
	if receiveUpdateModel.CountZPUAModel.ZoCount == 0 &&
		receiveUpdateModel.CountZPUAModel.PrCount == 0 &&
		receiveUpdateModel.CountZPUAModel.UpCount == 0 &&
		receiveUpdateModel.CountZPUAModel.AdCount == 0 {
		belogs.Debug("performHeaderTransactInServer(): receiveUpdateModel.CountZPUAModel all are 0 count,",
			"    receiveUpdateModel.CountZPUAModel:", jsonutil.MarshalJson(receiveUpdateModel.CountZPUAModel))
		return nil
	}
	return nil
}
