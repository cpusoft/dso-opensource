package transact

import (
	"dns-model/common"
	"dns-model/message"
	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/dnsutil"
	"github.com/cpusoft/goutil/jsonutil"
	"github.com/cpusoft/goutil/transportutil"
	querymodel "query-core/model"
)

func PerformQueryTransact(receiveDnsModel common.DnsModel, dnsToProcessMsg *message.DnsToProcessMsg) (responseDnsModel common.DnsModel, err error) {

	belogs.Debug("PerformQueryTransact():  receiveDnsModel:", jsonutil.MarshalJson(receiveDnsModel), "   dnsToProcessMsg:", jsonutil.MarshalJson(dnsToProcessMsg))
	receiveQueryModel, ok := (receiveDnsModel).(*querymodel.QueryModel)
	if !ok {
		belogs.Error("PerformQueryTransact():receiveDnsModel to receiveQueryModel, fail:", jsonutil.MarshalJson(receiveDnsModel))
		return nil, dnsutil.NewDnsError("fail to convert model type",
			0, dnsutil.DNS_OPCODE_QUERY, dnsutil.DNS_RCODE_SERVFAIL, transportutil.NEXT_CONNECT_POLICY_KEEP)
	}
	belogs.Debug("PerformQueryTransact(): receiveDnsModel to receiveQueryModel,  receiveQueryModel:", jsonutil.MarshalJson(receiveQueryModel))

	// get querymodel header (have no tlv)
	id := receiveQueryModel.GetHeaderModel().GetIdOrMessageId()
	belogs.Debug("PerformQueryTransact(): id:", id, " dnsToProcessMsg:", jsonutil.MarshalJson(dnsToProcessMsg))
	switch dnsToProcessMsg.DnsTransactSide {
	case message.DNS_TRANSACT_SIDE_SERVER:
		// query db and return response
		responseQueryModel, err := performQueryTransactInServer(receiveQueryModel)
		if err != nil {
			belogs.Error("PerformQueryTransact(): performQueryTransactInServer fail, receiveQueryModel:", jsonutil.MarshalJson(receiveQueryModel),
				"   dnsToProcessMsg:", jsonutil.MarshalJson(dnsToProcessMsg), err)
			return nil, err
		}
		return responseQueryModel, nil
	case message.DNS_TRANSACT_SIDE_CLIENT:
		// receiveQueryModel --> receiveQueryRrModel
		responseQueryRrModel, err := performQueryTransactInClient(receiveQueryModel)
		if err != nil {
			belogs.Error("PerformQueryTransact(): performQueryTransactInClient fail, receiveQueryModel:", jsonutil.MarshalJson(receiveQueryModel),
				"   dnsToProcessMsg:", jsonutil.MarshalJson(dnsToProcessMsg), err)
			return nil, err
		}
		return responseQueryRrModel, nil
	default:
		belogs.Error("PerformQueryTransact(): dnsTransactSide for header is not supported, fail:", dnsToProcessMsg.DnsTransactSide)
		return nil, dnsutil.NewDnsError("dnsTransactSide is not supported",
			id, dnsutil.DNS_OPCODE_QUERY, dnsutil.DNS_RCODE_NOTIMP, transportutil.NEXT_CONNECT_POLICY_KEEP)
	}

}

func performQueryTransactInServer(receiveQueryModel *querymodel.QueryModel) (responseQueryModel *querymodel.QueryModel, err error) {

	id := receiveQueryModel.GetHeaderModel().GetIdOrMessageId()
	questionModels := receiveQueryModel.QueryDataModel.QuestionModels
	if len(questionModels) != 1 {
		belogs.Error("performQueryTransactInServer(): len(questionModels) is not 1:", len(questionModels))
		return nil, dnsutil.NewDnsError("number of questionModels is too much",
			id, dnsutil.DNS_OPCODE_QUERY, dnsutil.DNS_RCODE_SERVFAIL, transportutil.NEXT_CONNECT_POLICY_KEEP)
	}

	questionRrModel, answerRrModels, authorityRrModels, additonalRrModels, err := queryQuestionDb(questionModels[0])
	if err != nil {
		belogs.Error("PerformQueryTransact(): queryQuestionDb fail, questionModels:", jsonutil.MarshalJson(questionModels[0]), err)
		return nil, dnsutil.NewDnsError(err.Error(),
			id, dnsutil.DNS_OPCODE_QUERY, dnsutil.DNS_RCODE_SERVFAIL, transportutil.NEXT_CONNECT_POLICY_KEEP)
	}
	belogs.Debug("PerformQueryTransact():answerRrModels:", jsonutil.MarshalJson(answerRrModels),
		"   authorityRrModels:", jsonutil.MarshalJson(authorityRrModels), "  additonalRrModels:", jsonutil.MarshalJson(additonalRrModels))

	responseQueryModel, err = querymodel.NewQueryModelForAnswerByParametersAndRrModels(id, 0,
		questionRrModel, answerRrModels, authorityRrModels, additonalRrModels)
	if err != nil {
		belogs.Error("PerformQueryTransact(): NewQueryModelForAnswerByParametersAndRrModels fail, questionModels:", jsonutil.MarshalJson(questionModels[0]), err)
		return nil, dnsutil.NewDnsError(err.Error(),
			id, dnsutil.DNS_OPCODE_QUERY, dnsutil.DNS_RCODE_SERVFAIL, transportutil.NEXT_CONNECT_POLICY_KEEP)
	}
	belogs.Info("PerformQueryTransact():questionModels:", jsonutil.MarshalJson(questionModels), "  responseQueryModel:", jsonutil.MarshalJson(responseQueryModel))
	return responseQueryModel, nil
}

func performQueryTransactInClient(receiveQueryModel *querymodel.QueryModel) (responseQueryRrModel *querymodel.QueryRrModel, err error) {
	belogs.Debug("performQueryTransactInClient():receiveQueryModel:", jsonutil.MarshalJson(receiveQueryModel))
	return querymodel.ConvertQueryModelToQueryRrModel(receiveQueryModel)
}
