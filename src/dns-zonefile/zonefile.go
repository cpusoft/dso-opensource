package zonefile

import (
	"dns-model/rr"
	"dns-model/zonefile"
	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/jsonutil"
)

func importZoneFile(receiveFile string) (err error) {
	belogs.Debug("importZoneFile(): receiveFile:", receiveFile)

	originModel, err := zonefile.LoadFromZoneFile(receiveFile)
	if err != nil {
		belogs.Error("importZoneFile(): LoadFromZoneFile fail:", receiveFile, err)
		return err
	}
	belogs.Debug("importZoneFile(): receiveFile:", receiveFile, "   originModel:", jsonutil.MarshalJson(originModel))

	err = saveOriginModelDb(originModel)
	if err != nil {
		belogs.Error("importZoneFile(): saveOriginModelDb fail:", jsonutil.MarshalJson(originModel), err)
		return err
	}
	belogs.Info("importZoneFile():importZoneFile ok, receiveFile:", receiveFile, "   originModel:", jsonutil.MarshalJson(originModel))
	return nil
}

func exportZoneFile(exportOrigin ExportOrigin) (originModel *rr.OriginModel, err error) {
	return getOriginModelDb(exportOrigin)
}

/*
var zoneFileServer *ZoneFileServer

func init() {
	var err error
	zoneFilePath := conf.String("dns-server::dataDir") + osutil.GetPathSeparator() + "zonefile" + osutil.GetPathSeparator()
	belogs.Info("init(): zoneFilePath:", zoneFilePath)
	zoneFileServer, err = NewZoneFileServer(zoneFilePath)
	if err != nil {
		belogs.Error("init(): NewZoneFileServer fail:", zoneFilePath, err)
		return
	}
	belogs.Info("init(): zoneFilePath:", jsonutil.MarshalJson(zoneFileServer.zoneFileModels))
}

func queryResourceRecord(zoneFileServer *ZoneFileServer, queryRr *zonefileutil.ResourceRecord) (resultRrs []*zonefileutil.ResourceRecord, err error) {
	belogs.Debug("queryResourceRecord(): zonefile queryRr:", jsonutil.MarshalJson(queryRr))
	err = zonefileutil.CheckNameAndTypeAndValues(queryRr)
	if err != nil {
		belogs.Error("queryResourceRecord(): CheckNameAndTypeAndValues fail:", jsonutil.MarshalJson(queryRr), err)
		return nil, err
	}
	resultRrs = make([]*zonefileutil.ResourceRecord, 0)
	var rrName string
	for origin, zoneFileModel := range zoneFileServer.zoneFileModels {
		if strings.HasSuffix(queryRr.RrDomain, origin) {
			rrName = strings.TrimSuffix(queryRr.RrDomain, origin)
			belogs.Debug("queryResourceRecord(): old queryRr:", jsonutil.MarshalJson(queryRr),
				" will found set rrName:", rrName, "   rrDomain:", queryRr.RrDomain, "   origin:", origin)
			queryRr.RrName = zonefileutil.FormatRrName(rrName)
			belogs.Debug("queryResourceRecord(): new queryRr:", jsonutil.MarshalJson(queryRr), "   origin:", origin)
			resultRrs, err = zonefileutil.QueryResourceRecords(zoneFileModel, queryRr)
			if err != nil {
				belogs.Error("queryResourceRecord(): QueryResourceRecords fail:", queryRr, err)
				return nil, err
			}
			break
		}
	}
	belogs.Info("queryResourceRecord():found, queryRr:", jsonutil.MarshalJson(queryRr),
		"  resultRrs:", jsonutil.MarshalJson(resultRrs))
	return resultRrs, nil
}

func addResourceRecord(zoneFileServer *ZoneFileServer, addRr *zonefileutil.ResourceRecord) (err error) {
	belogs.Debug("addResourceRecord(): addRr:", jsonutil.MarshalJson(addRr))
	err = zonefileutil.CheckDomainOrNameAndTypeAndValues(addRr, true)
	if err != nil {
		belogs.Error("addResourceRecord(): CheckNameAndTypeAndValues fail,addRr:", jsonutil.MarshalJson(addRr), err)
		return err
	}
	err = zonefileutil.CheckClassTypeShouldNoAny(addRr)
	if err != nil {
		belogs.Error("addResourceRecord(): CheckClassTypeShouldNoAny addRr fail:", jsonutil.MarshalJson(addRr), err)
		return err
	}

	for origin, zoneFileModel := range zoneFileServer.zoneFileModels {
		if strings.HasSuffix(addRr.RrDomain, origin) {
			rrName := strings.TrimSuffix(addRr.RrDomain, origin)
			belogs.Debug("addRr():found rrName:", rrName, "   origin:", origin)
			addRr.RrName = zonefileutil.FormatRrName(rrName)
			belogs.Debug("addResourceRecord():added rrName, addRr:", jsonutil.MarshalJson(addRr))
			err = zonefileutil.AddResourceRecord(zoneFileModel, nil, addRr)
			if err != nil {
				belogs.Error("addResourceRecord(): AddResourceRecord fail,addRr:", jsonutil.MarshalJson(addRr), err)
				return err
			}
			belogs.Info("addResourceRecord(): added addRr:", jsonutil.MarshalJson(addRr))
			return nil
		}
	}
	belogs.Error("addResourceRecord(): not found by origin, fail,  addRr:", jsonutil.MarshalJson(addRr))
	return errors.New("It is failed to add, because there cannot found zonefile by origin of domain")
}

// cannot update CLASS/TYPE=ANY, must specified RR
func updateResourceRecord(zoneFileServer *ZoneFileServer, oldRr, newRr *zonefileutil.ResourceRecord) (err error) {
	belogs.Debug("updateResourceRecord(): oldRr:", jsonutil.MarshalJson(oldRr),
		"  newRr:", jsonutil.MarshalJson(newRr))
	err = zonefileutil.CheckNameAndTypeAndValues(oldRr)
	if err != nil {
		belogs.Error("updateResourceRecord(): CheckNameAndTypeAndValues oldRr fail:", jsonutil.MarshalJson(oldRr), err)
		return err
	}
	err = zonefileutil.CheckDomainOrNameAndTypeAndValues(newRr, true)
	if err != nil {
		belogs.Error("updateResourceRecord(): CheckNameAndTypeAndValues newRr fail:", jsonutil.MarshalJson(newRr), err)
		return err
	}
	err = zonefileutil.CheckClassTypeShouldNoAny(oldRr)
	if err != nil {
		belogs.Error("updateResourceRecord(): CheckClassTypeShouldNoAny oldRr fail:", jsonutil.MarshalJson(oldRr), err)
		return err
	}
	err = zonefileutil.CheckClassTypeShouldNoAny(newRr)
	if err != nil {
		belogs.Error("updateResourceRecord(): CheckClassTypeShouldNoAny newRr fail:", jsonutil.MarshalJson(newRr), err)
		return err
	}

	for origin, zoneFileModel := range zoneFileServer.zoneFileModels {
		// have same origin
		if strings.HasSuffix(oldRr.RrDomain, origin) &&
			strings.HasSuffix(newRr.RrDomain, origin) {

			rrName := strings.TrimSuffix(oldRr.RrDomain, origin)
			belogs.Debug("updateResourceRecord():found rrName:", rrName, "   origin:", origin)
			oldRr.RrName = zonefileutil.FormatRrName(rrName)
			newRr.RrName = zonefileutil.FormatRrName(rrName)
			belogs.Debug("updateResourceRecord():added rrName, oldRr:", jsonutil.MarshalJson(oldRr),
				"  newRr:", jsonutil.MarshalJson(newRr))

			err = zonefileutil.UpdateResourceRecord(zoneFileModel,
				oldRr, newRr)
			if err != nil {
				belogs.Error("updateResourceRecord(): UpdateResourceRecord fail: oldRr:", jsonutil.MarshalJson(oldRr),
					"  newRr:", jsonutil.MarshalJson(newRr), err)
				return err
			}
			belogs.Info("updateResourceRecord(): update UpdateResourceRecord: oldRr:", jsonutil.MarshalJson(oldRr),
				"  newRr:", jsonutil.MarshalJson(newRr))
			return nil
		}
	}
	belogs.Error("updateResourceRecord(): not found by origin, fail, oldRr:", jsonutil.MarshalJson(oldRr),
		"  newRr:", jsonutil.MarshalJson(newRr))
	return errors.New("It is failed to update, because there cannot found zonefile by origin of domain")
}

// can del CLASS/TYPE=ANY
func delResourceRecord(zoneFileServer *ZoneFileServer, delRr *zonefileutil.ResourceRecord) (newDelRr *zonefileutil.ResourceRecord, err error) {
	belogs.Debug("delResourceRecord(): delRr:", jsonutil.MarshalJson(delRr))
	err = zonefileutil.CheckNameAndTypeAndValues(delRr)
	if err != nil {
		belogs.Error("delResourceRecord(): CheckNameAndTypeAndValues fail:", jsonutil.MarshalJson(delRr), err)
		return nil, err
	}

	for origin, zoneFileModel := range zoneFileServer.zoneFileModels {
		if strings.HasSuffix(delRr.RrDomain, origin) {
			rrName := strings.TrimSuffix(delRr.RrDomain, origin)
			belogs.Debug("delResourceRecord():found rrName:", rrName, "   origin:", origin)
			delRr.RrName = zonefileutil.FormatRrName(rrName)
			belogs.Debug("updateResourceRecord():added rrName, delRr:", jsonutil.MarshalJson(delRr))
			newDelRr, err = zonefileutil.DelResourceRecord(zoneFileModel, delRr)
			if err != nil {
				belogs.Error("delResourceRecord(): DelResourceRecord fail:", jsonutil.MarshalJson(delRr), err)
				return nil, err
			}
			belogs.Info("delResourceRecord(): del delRr:", jsonutil.MarshalJson(delRr), "  newDelRr:", jsonutil.MarshalJson(newDelRr))
			return newDelRr, nil
		}
	}
	belogs.Error("delResourceRecord(): not found by origin, fail:", jsonutil.MarshalJson(delRr))
	return nil, errors.New("It is failed to del, because there cannot found zonefile by origin of domain")
}

func getZoneFileModelByOrigin(zoneFileServer *ZoneFileServer, rr *zonefileutil.ResourceRecord) *zonefileutil.ZoneFileModel {
	for origin, zoneFileModel := range zoneFileServer.zoneFileModels {
		belogs.Debug("getZoneFileModelByOrigin(): rr.RrDomain:", rr.RrDomain, "   origin:", origin)
		if strings.HasSuffix(rr.RrDomain, origin) {
			belogs.Debug("getZoneFileModelByOrigin(): found rr.RrDomain:", rr.RrDomain, "   origin:", origin)
			return zoneFileModel
		}
	}
	return nil
}

func query(zoneFileServer *ZoneFileServer, queryRr *zonefileutil.ResourceRecord) (resultRrs []*zonefileutil.ResourceRecord, err error) {
	belogs.Debug("query(): queryRr :", jsonutil.MarshalJson(queryRr))

	// get zoneFileModel
	zoneFileModel := getZoneFileModelByOrigin(zoneFileServer, queryRr)
	if zoneFileModel == nil {
		belogs.Error("query(): getZoneFileModelByOrigin,  fail:", zoneFileServer, err)
		return nil, err
	}

	// query get result rr
	resultRrs, err = queryResourceRecord(zoneFileServer, queryRr)
	if err != nil {
		belogs.Error("query(): queryResourceRecord, fail:", queryRr, err)
		return nil, err
	}
	belogs.Info("query(): query :", jsonutil.MarshalJson(queryRr),
		"  result resultRrs:", jsonutil.MarshalJson(resultRrs))

	return resultRrs, nil

}

// addRr: added Rr: maybe nil //
// delRr: delete Rr: maybe nil
func saveAndQueryAndPush(zoneFileServer *ZoneFileServer,
	addRr, delRr *zonefileutil.ResourceRecord,
	needSave bool) (err error) {
	belogs.Debug("saveAndQueryAndPush(): addRr :", jsonutil.MarshalJson(addRr),
		"  delRr:", jsonutil.MarshalJson(delRr),
		"  needSave:", needSave)

	if needSave {
		// get zoneFileModel
		var zoneFileModel *zonefileutil.ZoneFileModel
		if addRr != nil {
			zoneFileModel = getZoneFileModelByOrigin(zoneFileServer, addRr)
		} else if delRr != nil {
			zoneFileModel = getZoneFileModelByOrigin(zoneFileServer, delRr)
		} else {
			belogs.Error("saveAndQueryAndPush(): addRr and delRr are all nil")
			return errors.New("addRr and delRr are all nil")
		}
		if zoneFileModel == nil {
			belogs.Error("saveAndQueryAndPush(): getZoneFileModelByOrigin,  fail:", zoneFileServer, err)
			return err
		}

		// save zone file
		err = zonefileutil.SaveZoneFile(zoneFileModel, "")
		if err != nil {
			belogs.Error("saveAndQueryAndPush(): SaveZoneFile,  fail:", zoneFileModel.ZoneFileName, err)
			return err
		}
	}

	// query get affected rr
	pushRrs := make([]*zonefileutil.ResourceRecord, 0)
	// when del, add delRr to affectedRrs
	if delRr != nil {
		// add delRr as index 0
		pushRrs = append(pushRrs, delRr)
		belogs.Info("saveAndQueryAndPush():add delRr to pushRrs, pushRrs :", jsonutil.MarshalJson(pushRrs))
	}
	// add or update, query affectedRrs
	if addRr != nil {
		affectedRrs, err := queryResourceRecord(zoneFileServer, addRr)
		if err != nil {
			belogs.Error("saveAndQueryAndPush(): queryResourceRecord, fail:", affectedRrs, err)
			return err
		}
		pushRrs = append(pushRrs, affectedRrs...)
		belogs.Info("saveAndQueryAndPush():add affectedRrs to pushRrs, affectedRrs :", jsonutil.MarshalJson(affectedRrs),
			"   pushRrs:", jsonutil.MarshalJson(pushRrs))
	}

	// if result rr is empty, so not send
	if len(pushRrs) == 0 {
		belogs.Error("saveAndQueryAndPush(): pushRrs is emtpy, fail, addRr :", jsonutil.MarshalJson(addRr), "  delRr:", jsonutil.MarshalJson(delRr))
		return errors.New("pushRrs is empty")
	}

	err = pushResourceRecoreds(pushRrs)
	if err != nil {
		belogs.Error("saveAndQueryAndPush(): pushResourceRecoreds, fail, pushRrs:", jsonutil.MarshalJson(pushRrs), err)
		return err
	}
	belogs.Info("saveAndQueryAndPush(): addRr :", jsonutil.MarshalJson(addRr),
		"  delRr:", jsonutil.MarshalJson(delRr),
		"  pushRrs:", jsonutil.MarshalJson(pushRrs))
	return nil

}

func pushResourceRecoreds(resourceRecords []*zonefileutil.ResourceRecord) (err error) {

	host := "https://" + conf.String("dns-server::serverHost") + ":" + conf.String("dns-server::serverHttpsPort")
	belogs.Info("sendToZoneFile(): start,  host:", host, "   json:", jsonutil.MarshalJson(resourceRecords))
	// send to zonefile
	go httpclient.Post(host+"/push/triggerpush", jsonutil.MarshalJson(resourceRecords), false)
	return nil
}
*/
