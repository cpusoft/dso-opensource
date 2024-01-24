package model

import (
	"errors"

	"dns-model/common"
	dnsconvert "dns-model/convert"
	"dns-model/rr"
	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/dnsutil"
	"github.com/cpusoft/goutil/jsonutil"
	"github.com/guregu/null"
)

type UpdateRrModel struct {
	HeaderForUpdateModel common.HeaderForUpdateModel `json:"headerForUpdateModel"`
	CountZPUAModel       common.CountZPUAModel       `json:"countZPUAModel"`
	UpdateDataRrModel    UpdateDataRrModel           `json:"updateDataRrModel"`
}

func (c UpdateRrModel) GetHeaderModel() common.HeaderModel {
	return c.HeaderForUpdateModel
}
func (c UpdateRrModel) GetCountModel() common.CountModel {
	return c.CountZPUAModel
}
func (c UpdateRrModel) GetDataModel() interface{} {
	return c.UpdateDataRrModel
}
func (c UpdateRrModel) Bytes() []byte {
	return []byte(jsonutil.MarshalJson(c))
}
func (c UpdateRrModel) GetDnsModelType() string {
	return "rr"
}

type UpdateDataRrModel struct {
	ZoneRrModel            *rr.RrModel   `json:"zoneRrModel"`
	PrerequisiteRrModels   []*rr.RrModel `json:"prerequisiteRrModel"`
	UpdateRrModels         []*rr.RrModel `json:"updateRrModel"`
	AdditionalDataRrModels []*rr.RrModel `json:"additionalDataRrModel"`
}

func ConvertUpdateModelToUpdateRrModel(receiveUpdateModel *UpdateModel) (updateRrModel *UpdateRrModel, err error) {
	c := &UpdateRrModel{
		HeaderForUpdateModel: receiveUpdateModel.HeaderForUpdateModel,
		CountZPUAModel:       receiveUpdateModel.CountZPUAModel,
	}
	belogs.Debug("ConvertUpdateModelToUpdateRrModel(): UpdateRrModel header and count:", jsonutil.MarshalJson(c))

	belogs.Debug("ConvertUpdateModelToUpdateRrModel(): ZoneModel:", jsonutil.MarshalJson(receiveUpdateModel.UpdateDataModel.ZoneModel))
	origin := string(receiveUpdateModel.UpdateDataModel.ZoneModel.ZNamePacketDomain.FullDomain)
	belogs.Debug("ConvertUpdateModelToUpdateRrModel(): origin:", origin)
	rrType, ok := dnsutil.DnsIntTypes[receiveUpdateModel.UpdateDataModel.ZoneModel.ZType]
	if !ok {
		belogs.Error("ConvertUpdateModelToUpdateRrModel(): ZType to rrType fail:",
			" zType:", receiveUpdateModel.UpdateDataModel.ZoneModel.ZType)
		return nil, errors.New("ZType to rrType fail")
	}
	belogs.Debug("ConvertUpdateModelToUpdateRrModel(): rrType:", rrType)
	rrClass, ok := dnsutil.DnsIntClasses[receiveUpdateModel.UpdateDataModel.ZoneModel.ZClass]
	if !ok {
		belogs.Error("ConvertUpdateModelToUpdateRrModel(): ZClass to rrClass fail:",
			" ZClass:", receiveUpdateModel.UpdateDataModel.ZoneModel.ZClass)
		return nil, errors.New("ZClass to rrClass fail")
	}
	belogs.Debug("ConvertUpdateModelToUpdateRrModel(): rrClass:", rrClass)
	c.UpdateDataRrModel.ZoneRrModel = rr.NewRrModel(origin, "", rrType, rrClass, null.NewInt(0, false), "")
	belogs.Debug("ConvertUpdateModelToUpdateRrModel():c.UpdateDataRrModel.ZoneRrModel:", jsonutil.MarshalJson(c.UpdateDataRrModel.ZoneRrModel))

	belogs.Debug("ConvertUpdateModelToUpdateRrModel(): len(PrerequisiteModels):", len(receiveUpdateModel.UpdateDataModel.PrerequisiteModels))
	if len(receiveUpdateModel.UpdateDataModel.PrerequisiteModels) > 0 {
		c.UpdateDataRrModel.PrerequisiteRrModels = make([]*rr.RrModel, 0)
		for i := range receiveUpdateModel.UpdateDataModel.PrerequisiteModels {
			prerequisiteRrModel, err := dnsconvert.ConvertPacketToRr(origin, receiveUpdateModel.UpdateDataModel.PrerequisiteModels[i])
			if err != nil {
				belogs.Error("ConvertUpdateModelToUpdateRrModel(): PrerequisiteModels ConvertPacketToRr fail:",
					" PrerequisiteModels[i]:", jsonutil.MarshalJson(receiveUpdateModel.UpdateDataModel.PrerequisiteModels[i]), err)
				return nil, err
			}
			belogs.Debug("ConvertUpdateModelToUpdateRrModel():prerequisiteRrModel:", jsonutil.MarshalJson(prerequisiteRrModel))
			c.UpdateDataRrModel.PrerequisiteRrModels = append(c.UpdateDataRrModel.PrerequisiteRrModels, prerequisiteRrModel)
		}
	}
	belogs.Debug("ConvertUpdateModelToUpdateRrModel():c.UpdateDataRrModel.PrerequisiteRrModels:", jsonutil.MarshalJson(c.UpdateDataRrModel.PrerequisiteRrModels))

	belogs.Debug("ConvertUpdateModelToUpdateRrModel(): len(UpdateRrModels):", len(receiveUpdateModel.UpdateDataModel.UpdateModels))
	if len(receiveUpdateModel.UpdateDataModel.UpdateModels) > 0 {
		c.UpdateDataRrModel.UpdateRrModels = make([]*rr.RrModel, 0)
		for i := range receiveUpdateModel.UpdateDataModel.UpdateModels {
			updateRrModel, err := dnsconvert.ConvertPacketToRr(origin, receiveUpdateModel.UpdateDataModel.UpdateModels[i])
			if err != nil {
				belogs.Error("ConvertUpdateModelToUpdateRrModel(): UpdateModels ConvertPacketToRr fail:",
					" UpdateModels[i]:", jsonutil.MarshalJson(receiveUpdateModel.UpdateDataModel.UpdateModels[i]), err)
				return nil, err
			}
			belogs.Debug("ConvertUpdateModelToUpdateRrModel():updateRrModel:", jsonutil.MarshalJson(updateRrModel))
			c.UpdateDataRrModel.UpdateRrModels = append(c.UpdateDataRrModel.UpdateRrModels, updateRrModel)
		}
	}
	belogs.Debug("ConvertUpdateModelToUpdateRrModel():c.UpdateDataRrModel.UpdateRrModels:", jsonutil.MarshalJson(c.UpdateDataRrModel.UpdateRrModels))

	belogs.Debug("ConvertUpdateModelToUpdateRrModel(): len(AdditionalDataModels):", len(receiveUpdateModel.UpdateDataModel.AdditionalDataModels))
	if len(receiveUpdateModel.UpdateDataModel.AdditionalDataModels) > 0 {
		c.UpdateDataRrModel.AdditionalDataRrModels = make([]*rr.RrModel, 0)
		for i := range receiveUpdateModel.UpdateDataModel.AdditionalDataModels {
			additonalRrModel, err := dnsconvert.ConvertPacketToRr(origin, receiveUpdateModel.UpdateDataModel.AdditionalDataModels[i])
			if err != nil {
				belogs.Error("ConvertUpdateModelToUpdateRrModel(): AdditionalDataModels ConvertPacketToRr fail:",
					" AdditionalDataModels[i]:", jsonutil.MarshalJson(receiveUpdateModel.UpdateDataModel.AdditionalDataModels[i]), err)
				return nil, err
			}
			belogs.Debug("ConvertUpdateModelToUpdateRrModel():additonalRrModel:", jsonutil.MarshalJson(additonalRrModel))
			c.UpdateDataRrModel.AdditionalDataRrModels = append(c.UpdateDataRrModel.AdditionalDataRrModels, additonalRrModel)
		}
	}
	belogs.Debug("ConvertUpdateModelToUpdateRrModel():c.UpdateDataRrModel.AdditionalDataRrModels:", jsonutil.MarshalJson(c.UpdateDataRrModel.AdditionalDataRrModels))

	return c, nil

}
