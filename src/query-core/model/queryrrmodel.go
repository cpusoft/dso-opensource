package model

import (
	"errors"

	"dns-model/common"
	dnsconvert "dns-model/convert"
	"dns-model/rr"
	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/jsonutil"
)

type QueryRrModel struct {
	HeaderForQueryModel common.HeaderForQueryModel `json:"headerForQueryModel"`
	CountQANAModel      common.CountQANAModel      `json:"countQANAModel"`
	QueryDataRrModel    QueryDataRrModel           `json:"queryDataRrModel"`
}

func (c QueryRrModel) GetHeaderModel() common.HeaderModel {
	return c.HeaderForQueryModel
}
func (c QueryRrModel) GetCountModel() common.CountModel {
	return c.CountQANAModel
}
func (c QueryRrModel) GetDataModel() interface{} {
	return c.QueryDataRrModel
}
func (c QueryRrModel) Bytes() []byte {
	return []byte(jsonutil.MarshalJson(c))
}
func (c QueryRrModel) GetDnsModelType() string {
	return "rr"
}

type QueryDataRrModel struct {
	QuestionRrModel   *rr.RrModel   `json:"questionRrModels"`
	AnswerRrModels    []*rr.RrModel `json:"answerRrModels"`
	AuthorityRrModels []*rr.RrModel `json:"authorityRrModels"`
	AdditonalRrModels []*rr.RrModel `json:"additonalRrModels"`
}

func ConvertQueryModelToQueryRrModel(receiveQueryModel *QueryModel) (queryRrModel *QueryRrModel, err error) {
	belogs.Debug("ConvertQueryModelToQueryRrModel(): receiveQueryModel:", jsonutil.MarshalJson(receiveQueryModel))

	c := &QueryRrModel{
		HeaderForQueryModel: receiveQueryModel.HeaderForQueryModel,
		CountQANAModel:      receiveQueryModel.CountQANAModel,
	}
	belogs.Debug("ConvertQueryModelToQueryRrModel(): QueryRrModel header and count:", jsonutil.MarshalJson(c))

	belogs.Debug("ConvertQueryModelToQueryRrModel(): len(QuestionModels):", len(receiveQueryModel.QueryDataModel.QuestionModels))
	if len(receiveQueryModel.QueryDataModel.QuestionModels) == 1 {
		c.QueryDataRrModel.QuestionRrModel, err = dnsconvert.ConvertPacketToRr("", receiveQueryModel.QueryDataModel.QuestionModels[0])
		if err != nil {
			belogs.Error("ConvertQueryModelToQueryRrModel(): QuestionModels ConvertPacketToRr fail:",
				" QuestionModels[0]:", jsonutil.MarshalJson(receiveQueryModel.QueryDataModel.QuestionModels[0]), err)
			return nil, err
		}
		belogs.Debug("ConvertQueryModelToQueryRrModel():c.QueryDataRrModel.QuestionRrModel:", jsonutil.MarshalJson(c.QueryDataRrModel.QuestionRrModel))
	} else {
		belogs.Error("ConvertQueryModelToQueryRrModel(): len(QuestionModels) is not 1:",
			" QuestionModels:", jsonutil.MarshalJson(receiveQueryModel.QueryDataModel.QuestionModels))
		return nil, errors.New("len(QuestionModels) is not 1")
	}
	belogs.Debug("ConvertQueryModelToQueryRrModel():c.QueryDataRrModel.QuestionRrModel:", jsonutil.MarshalJson(c.QueryDataRrModel.QuestionRrModel))

	belogs.Debug("ConvertQueryModelToQueryRrModel(): len(AnswerModels):", len(receiveQueryModel.QueryDataModel.AnswerModels))
	if len(receiveQueryModel.QueryDataModel.AnswerModels) > 0 {
		c.QueryDataRrModel.AnswerRrModels = make([]*rr.RrModel, 0)
		for i := range receiveQueryModel.QueryDataModel.AnswerModels {
			answerRrModel, err := dnsconvert.ConvertPacketToRr("", receiveQueryModel.QueryDataModel.AnswerModels[i])
			if err != nil {
				belogs.Error("ConvertQueryModelToQueryRrModel(): AnswerModels ConvertPacketToRr fail:",
					" AnswerModels[i]:", jsonutil.MarshalJson(receiveQueryModel.QueryDataModel.AnswerModels[i]), err)
				return nil, err
			}
			belogs.Debug("ConvertQueryModelToQueryRrModel():answerRrModel:", jsonutil.MarshalJson(answerRrModel))
			c.QueryDataRrModel.AnswerRrModels = append(c.QueryDataRrModel.AnswerRrModels, answerRrModel)
		}
	}
	belogs.Debug("ConvertQueryModelToQueryRrModel():c.QueryDataRrModel.AnswerRrModels:", jsonutil.MarshalJson(c.QueryDataRrModel.AnswerRrModels))

	belogs.Debug("ConvertQueryModelToQueryRrModel(): len(AuthorityModels):", len(receiveQueryModel.QueryDataModel.AuthorityModels))
	if len(receiveQueryModel.QueryDataModel.AuthorityModels) > 0 {
		c.QueryDataRrModel.AuthorityRrModels = make([]*rr.RrModel, 0)
		for i := range receiveQueryModel.QueryDataModel.AuthorityModels {
			authorityRrModel, err := dnsconvert.ConvertPacketToRr("", receiveQueryModel.QueryDataModel.AuthorityModels[i])
			if err != nil {
				belogs.Error("ConvertQueryModelToQueryRrModel(): AuthorityModels ConvertPacketToRr fail:",
					" AuthorityModels[i]:", jsonutil.MarshalJson(receiveQueryModel.QueryDataModel.AuthorityModels[i]), err)
				return nil, err
			}
			belogs.Debug("ConvertQueryModelToQueryRrModel():authorityRrModel:", jsonutil.MarshalJson(authorityRrModel))
			c.QueryDataRrModel.AuthorityRrModels = append(c.QueryDataRrModel.AuthorityRrModels, authorityRrModel)
		}
	}
	belogs.Debug("ConvertQueryModelToQueryRrModel():c.QueryDataRrModel.AuthorityRrModels:", jsonutil.MarshalJson(c.QueryDataRrModel.AuthorityRrModels))

	belogs.Debug("ConvertQueryModelToQueryRrModel(): len(AdditonalModels):", len(receiveQueryModel.QueryDataModel.AdditonalModels))
	if len(receiveQueryModel.QueryDataModel.AdditonalModels) > 0 {
		c.QueryDataRrModel.AdditonalRrModels = make([]*rr.RrModel, 0)
		for i := range receiveQueryModel.QueryDataModel.AdditonalModels {
			additonalRrModel, err := dnsconvert.ConvertPacketToRr("", receiveQueryModel.QueryDataModel.AdditonalModels[i])
			if err != nil {
				belogs.Error("ConvertQueryModelToQueryRrModel(): AdditonalModels ConvertPacketToRr fail:",
					" AdditonalModels[i]:", jsonutil.MarshalJson(receiveQueryModel.QueryDataModel.AdditonalModels[i]), err)
				return nil, err
			}
			belogs.Debug("ConvertQueryModelToQueryRrModel():additonalRrModel:", jsonutil.MarshalJson(additonalRrModel))
			c.QueryDataRrModel.AdditonalRrModels = append(c.QueryDataRrModel.AdditonalRrModels, additonalRrModel)
		}
	}
	belogs.Debug("ConvertQueryModelToQueryRrModel():c.QueryDataRrModel.AdditonalRrModels:", jsonutil.MarshalJson(c.QueryDataRrModel.AdditonalRrModels))

	belogs.Debug("ConvertQueryModelToQueryRrModel(): queryRrModel:", jsonutil.MarshalJson(c))
	return c, nil

}
