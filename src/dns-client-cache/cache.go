package clientcache

import (
	"errors"

	"dns-model/rr"
	"github.com/cpusoft/goutil/belogs"
)

func InitCache() error {
	err := initDb()
	if err != nil {
		belogs.Error("InitCache(): initDb fail:", err)
		return err
	}
	err = initMem()
	if err != nil {
		belogs.Error("InitCache(): initMem fail:", err)
		return err
	}
	return nil
}
func ResetCache() error {
	return resetTableDb()
}

func UpdateRrModel(rrModel *rr.RrModel, justDel bool) error {
	if len(rrModel.RrFullDomain) == 0 && len(rrModel.RrType) == 0 {
		belogs.Error("UpdateRrModel(): rrFullDomain and rrType are all empty")
		return errors.New("rrFullDomain and rrType are all empty")
	}
	if !justDel {
		if len(rrModel.RrFullDomain) == 0 || len(rrModel.RrType) == 0 || len(rrModel.RrClass) == 0 ||
			rrModel.RrTtl.IsZero() || len(rrModel.RrData) == 0 {
			belogs.Error("UpdateRrModel(): is not justDel, rrFullDomain or rrType or rrClass or rrTtl or rrData is empty")
			return errors.New("rrFullDomain or rrType or rrClass or rrTtl or rrData is empty")
		}
	}
	return updateRrModelDb(rrModel, justDel)
}

// rrModel isnot nil
func QueryRrModels(rrModel *rr.RrModel) ([]*rr.RrModel, error) {
	if rrModel == nil {
		return nil, errors.New("rrModel is empty")
	}
	return queryRrModelsDb(rrModel)
}
func QueryAllRrModels() ([]*rr.RrModel, error) {
	return queryAllRrModelsDb()
}
func ClearAllRrModels() error {
	return clearAllRrModelsDb()
}

func GetNewMessageId(opCode uint8) (uint16, error) {
	return getNewMessageIdDb(opCode)
}
func UpdateDsoMessageDsoTypeAndRrModel(messageId uint16, dsoType uint8, rrModel *rr.RrModel) error {
	return updateDsoMessageDsoTypeAndRrModelDb(messageId, dsoType, rrModel)
}
func UpdateDsoMessageUnsubscribeTime(messageId uint16) error {
	return updateDsoMessageUnsubscribeTimeDb(messageId)
}

func QueryDsoMessageIdByRrModelDb(rrModel *rr.RrModel) (bool, uint16, error) {
	return queryDsoMessageIdByRrModelDb(rrModel)
}

func ConfirmMessageId(messageId uint16) (bool, error) {
	return confirmMessageIdDb(messageId)
}

/*
func AddSubscribeRrModel(messageId uint16, subsribeRr *rr.RrModel) (err error) {
	belogs.Debug("AddSubscribeRrModel(): messageId:", messageId, "   subsribeRr:", jsonutil.MarshalJson(subsribeRr))

	addRrKey := dnsmodel.GetRrModelKey(subsribeRr)
	belogs.Debug("AddSubscribeRrModel(): subsribeRr:", jsonutil.MarshalJson(subsribeRr), "   addRrKey:", addRrKey)
	_, ok := localResourceRecordCache.SubscribeRrModels[addRrKey]
	if ok {
		belogs.Info("AddSubscribeRrModel():found in cache, subsribeRr:", jsonutil.MarshalJson(subsribeRr), "   addRrKey:", addRrKey)
		return errors.New("already have subscribed '" + addRrKey + "'")
	}

	srr := NewSubscribeRrModel(messageId, subsribeRr)
	belogs.Debug("AddSubscribeRrModel(): addRrKey:", addRrKey, "  srr:", jsonutil.MarshalJson(srr))
	localResourceRecordCache.SubscribeRrModels[addRrKey] = srr
	belogs.Debug("AddSubscribeRrModel(): localResourceRecordCache.SubscribeRrModels:", jsonutil.MarshalJson(localResourceRecordCache.SubscribeRrModels))

	return nil
}

func DelSubscribeRrModel(messageId uint16) (err error) {
	belogs.Debug("DelSubscribeRrModel(): messageId:", messageId,
		"  localResourceRecordCache.SubscribeRrModels:", jsonutil.MarshalJson(localResourceRecordCache.SubscribeRrModels))

	for rrKey := range localResourceRecordCache.SubscribeRrModels {
		if messageId == localResourceRecordCache.SubscribeRrModels[rrKey].MessageId {
			belogs.Info("DelSubscribeRrModel():found in cache, messageId:", messageId, "  rrKey:", rrKey,
				"   localResourceRecordCache.SubscribeRrModels[i]:",
				jsonutil.MarshalJson(localResourceRecordCache.SubscribeRrModels[rrKey]))
			delete(localResourceRecordCache.SubscribeRrModels, rrKey)
			belogs.Debug("DelSubscribeRrModel(): new localResourceRecordCache.SubscribeRrModels:", jsonutil.MarshalJson(localResourceRecordCache.SubscribeRrModels))
			return nil
		}
	}
	belogs.Debug("DelSubscribeRrModel(): not found this messageId:", messageId)
	return nil

}

func UpdateSubscribeResourceRecordResult(messageId uint16, rCode uint8) (err error) {
	belogs.Debug("UpdateSubscribeResourceRecordResult(): messageId:", messageId, "  rCode:", rCode,
		"  localResourceRecordCache.SubscribeRrModels:", jsonutil.MarshalJson(localResourceRecordCache.SubscribeRrModels))

	for rrKey := range localResourceRecordCache.SubscribeRrModels {
		if messageId == localResourceRecordCache.SubscribeRrModels[rrKey].MessageId {
			belogs.Info("UpdateSubscribeResourceRecordResult():found in cache, messageId:", messageId, "  rrKey:", rrKey,
				"   localResourceRecordCache.SubscribeRrModels[i]:",
				jsonutil.MarshalJson(localResourceRecordCache.SubscribeRrModels[rrKey]))
			localResourceRecordCache.SubscribeRrModels[rrKey].RCode = rCode
			if rCode == dnsutil.DNS_RCODE_NOERROR {
				localResourceRecordCache.SubscribeRrModels[rrKey].Result = "ok"
			} else {
				localResourceRecordCache.SubscribeRrModels[rrKey].Result = "fail"
			}
			belogs.Debug("UpdateSubscribeResourceRecordResult(): new localResourceRecordCache.SubscribeRrModels:", jsonutil.MarshalJson(localResourceRecordCache.SubscribeRrModels))

			return nil
		}
	}
	belogs.Debug("UpdateSubscribeResourceRecordResult(): not found this messageId:", messageId)
	return nil
}

func AddKnownResourceRecord(knownRr *dnsmodel.RrModel) (exist bool, err error) {
	belogs.Debug("AddKnownResourceRecord(): knownRr:", jsonutil.MarshalJson(knownRr),
		"  localResourceRecordCache.KnownRrModels:", jsonutil.MarshalJson(localResourceRecordCache.KnownRrModels))

	for i := range localResourceRecordCache.KnownRrModels {
		if dnsmodel.EqualRr(knownRr, localResourceRecordCache.KnownRrModels[i]) {
			return true, nil
		}
	}
	localResourceRecordCache.KnownRrModels = append(localResourceRecordCache.KnownRrModels, knownRr)
	belogs.Debug("AddKnownResourceRecord(): added knownRr:", jsonutil.MarshalJson(knownRr),
		"  localResourceRecordCache.KnownRrModels:", jsonutil.MarshalJson(localResourceRecordCache.KnownRrModels))
	return false, nil
}

func DelKnownResourceRecord(knownRr *dnsmodel.RrModel) (err error) {
	belogs.Debug("DelKnownResourceRecord(): knownRr:", jsonutil.MarshalJson(knownRr),
		"  localResourceRecordCache.KnownRrModels:", jsonutil.MarshalJson(localResourceRecordCache.KnownRrModels))
	newKnownRrs := make([]*dnsmodel.RrModel, 0)
	if knownRr.RrClass == "ANY" || (knownRr.RrClass != "ANY" && knownRr.RrType == "ANY") {
		belogs.Debug("DelKnownResourceRecord(): collective rr and not rrValues, class:", knownRr.RrClass, "  type:", knownRr.RrType)
		for i := range localResourceRecordCache.KnownRrModels {
			if localResourceRecordCache.KnownRrModels[i].RrName != knownRr.RrName {
				newKnownRrs = append(newKnownRrs, localResourceRecordCache.KnownRrModels[i])
			}
		}
	} else if knownRr.RrClass != "ANY" && knownRr.RrType != "ANY" && len(knownRr.RrData) == 0 {
		belogs.Debug("DelKnownResourceRecord(): collective rr and rrValues, class:", knownRr.RrClass, "  type:", knownRr.RrType)
		for i := range localResourceRecordCache.KnownRrModels {
			if !(localResourceRecordCache.KnownRrModels[i].RrName == knownRr.RrName &&
				localResourceRecordCache.KnownRrModels[i].RrClass == knownRr.RrClass &&
				localResourceRecordCache.KnownRrModels[i].RrType == knownRr.RrType) {
				newKnownRrs = append(newKnownRrs, localResourceRecordCache.KnownRrModels[i])
			}
		}
	} else if knownRr.RrClass != "ANY" && knownRr.RrType != "ANY" && len(knownRr.RrData) != 0 {
		belogs.Debug("DelKnownResourceRecord(): specified rr and rrValues, class:", knownRr.RrClass, "  type:", knownRr.RrType)
		for i := range localResourceRecordCache.KnownRrModels {
			if !dnsmodel.EqualRr(localResourceRecordCache.KnownRrModels[i], knownRr) {
				newKnownRrs = append(newKnownRrs, localResourceRecordCache.KnownRrModels[i])
			}
		}
	}

	localResourceRecordCache.KnownRrModels = newKnownRrs
	belogs.Debug("DelKnownResourceRecord(): after del, new localResourceRecordCache.KnownRrModels:", jsonutil.MarshalJson(localResourceRecordCache.KnownRrModels))
	return nil

}

func QueryKnownResourceRecord(queryRr *dnsmodel.RrModel) (resultRrs []*dnsmodel.RrModel, err error) {
	belogs.Debug("QueryKnownResourceRecord(): queryRr:", jsonutil.MarshalJson(queryRr),
		"  localResourceRecordCache.KnownRrModels:", jsonutil.MarshalJson(localResourceRecordCache.KnownRrModels))
	resultRrs = make([]*dnsmodel.RrModel, 0)
	// check
	if queryRr == nil || len(queryRr.RrFullDomain) == 0 {
		return resultRrs, errors.New("rrDomain is empty")
	}
	if len(queryRr.RrType) == 0 {
		queryRr.RrType = "ANY"
	}

	for i := range localResourceRecordCache.KnownRrModels {
		if queryRr.RrFullDomain == localResourceRecordCache.KnownRrModels[i].RrFullDomain {
			if queryRr.RrType == "ANY" {
				resultRrs = append(resultRrs, localResourceRecordCache.KnownRrModels[i])
			} else {
				if queryRr.RrType == localResourceRecordCache.KnownRrModels[i].RrType {
					resultRrs = append(resultRrs, localResourceRecordCache.KnownRrModels[i])
				}
			}
		}
	}

	belogs.Info("QueryKnownResourceRecord(): queryRr:", jsonutil.MarshalJson(queryRr), "  resultRrs:", jsonutil.MarshalJson(resultRrs))
	return resultRrs, nil

}

func ShowAllKnownRrModels() (resultRrs []*dnsmodel.RrModel, err error) {
	belogs.Info("ShowAllKnownRrModels(): KnownRrModels:", jsonutil.MarshalJson(localResourceRecordCache.KnownRrModels))
	return localResourceRecordCache.KnownRrModels, nil

}
*/
