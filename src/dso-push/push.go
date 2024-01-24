package push

import (
	"errors"
	"strings"

	pushmodel "dns-model/push"
	"dns-model/rr"
	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/dnsutil"
	"github.com/cpusoft/goutil/jsonutil"
)

var pushCache *PushCache

func init() {
	pushCache = NewPushCache()
}

// pushResultRrModels may be empty
func subscribe(pushRrModel *pushmodel.PushRrModel) (pushResultRrModels []*pushmodel.PushResultRrModel, err error) {
	belogs.Debug("subscribe(): pushRrModel:", jsonutil.MarshalJson(pushRrModel))
	// subscribe to pushCache
	connKey := pushRrModel.ConnKey
	rrFullDomain := pushRrModel.RrFullDomain
	rrType := pushRrModel.RrType
	rrClass := pushRrModel.RrClass
	subscribeMessageId := pushRrModel.SubscribeMessageId
	connAndRrKey, connAndRrAnyKey, connAndRrDelKey := getConnAndRrKeys(connKey, rrFullDomain, rrType, rrClass)
	belogs.Debug("subscribe(): connKey:", connKey, "  connAndRrKey:", connAndRrKey,
		"  connAndRrAnyKey:", connAndRrAnyKey, "  connAndRrDelKey:", connAndRrDelKey,
		"  subscribeMessageId:", subscribeMessageId)

	pushCache.subscribe(connAndRrKey, connAndRrAnyKey, connAndRrDelKey, subscribeMessageId)
	belogs.Debug("subscribe(): subscribe, connKey:", connKey, "  rrFullDomain:", rrFullDomain, "  rrType:", rrType, "  rrClass:", rrClass)

	// query to get current resultRr
	rrModels, err := queryDb(rrFullDomain, rrType)
	if err != nil {
		belogs.Error("subscribe(): queryDb, fail, rrFullDomain:", rrFullDomain, "  rrType:", rrType, "  rrClass:", rrClass, err)
		return nil, err
	}
	if len(rrModels) == 0 {
		belogs.Debug("subscribe(): queryDb, rrModels is empty, rrFullDomain:", rrFullDomain, "  rrType:", rrType)
		return make([]*pushmodel.PushResultRrModel, 0), nil
	}
	belogs.Debug("subscribe(): resultRrModels ok, connKey:", connKey, "  rrModels:", jsonutil.MarshalJson(rrModels))
	pushResultRrModel := pushmodel.NewPushResultModel(connKey)
	pushResultRrModel.AddRrModels(false, rrModels)
	belogs.Debug("subscribe():pushResultRrModel, pushResultRrModel:", jsonutil.MarshalJson(pushResultRrModel))

	pushResultRrModels = make([]*pushmodel.PushResultRrModel, 0)
	pushResultRrModels = append(pushResultRrModels, pushResultRrModel)
	return pushResultRrModels, nil
}

func unsubscribe(unpushRrModel *pushmodel.UnpushRrModel) (err error) {
	belogs.Debug("unsubscribe(): unpushRrModel:", jsonutil.MarshalJson(unpushRrModel))
	// subscribe to pushCache
	connKey := unpushRrModel.ConnKey
	subscribeMessageId := unpushRrModel.SubscribeMessageId

	// subscribe to pushCache
	pushCache.unsubscribe(subscribeMessageId)
	belogs.Debug("unsubscribe(): unsubscribe, connKey:", connKey, "  subscribeMessageId:", subscribeMessageId)
	return nil
}

/*
func reconfirmResourceRecordInConn(pushServer *PushServer, connKey string,
	resourceRecord *dnsmodel.RrModel) (err error) {
	belogs.Debug("reconfirmResourceRecordInConn(): connKey:", connKey, "  resourceRecord:", jsonutil.MarshalJson(resourceRecord))

	// send to zonefile, to active queryandpush
	sendToZoneFile("queryandpush", resourceRecord)
	return
}
*/

func delConn(connKey string) {
	belogs.Debug("delConn(): connKey:", connKey)
	pushCache.delConnKey(connKey)

}

// Some subscribers maybe are only interested in some results, others are interested in others
// so get ConnKeys
/*
rfc8765 6.3.1. PUSH Message



push del:
 if TTL==0xFFFFFFFF and Class!=ANY and Type!=ANY , then del by FullDomain and Class and Type and RData
 if TTL==0xFFFFFFFE
	if Class!=ANY and Type!=ANY, then del by FullDomain and Type and Class
 	if Class!=ANY and Type==ANY, then del by FullDomain and Class
 	if Class==ANY , delete by fullDomain

rrModels : already be convert to push from update

pushResultRrModels may be empty
*/
func queryRrModelsShouldPush(pushRrModels []*pushmodel.PushRrModel) (pushResultRrModels []*pushmodel.PushResultRrModel, err error) {
	belogs.Debug("queryRrModelsShouldPush(): pushRrModels:", jsonutil.MarshalJson(pushRrModels))
	// connKey should be from subscribedRrs, not from pushRrModel
	mapPushResultRrModels := make(map[string]*pushmodel.PushResultRrModel)
	for i := range pushRrModels {
		rrKey := rr.GetRrKey(pushRrModels[i].RrFullDomain, pushRrModels[i].RrType, pushRrModels[i].RrClass)
		rrAnyKey := rr.GetRrKey(pushRrModels[i].RrFullDomain, dnsutil.DNS_TYPE_STR_ANY, pushRrModels[i].RrClass)
		var rrDelKey string
		if rr.IsDelRrModelForDso(pushRrModels[i].RrTtl) {
			rrDelKey = rr.GetRrKey(pushRrModels[i].RrFullDomain, dnsutil.DNS_RR_DEL_KEY, pushRrModels[i].RrClass)
		}
		belogs.Debug("queryRrModelsShouldPush(): rrKey:", rrKey, "   rrAnyKey:", rrAnyKey, "  rrDelKey:", rrDelKey)
		connAndRrKeys, _ := pushCache.foundInSubscribedRrs(rrKey, rrAnyKey, rrDelKey)
		belogs.Debug("queryRrModelsShouldPush(): foundInSubscribedRrs, connAndRrKeys:", connAndRrKeys)
		if len(connAndRrKeys) == 0 {
			belogs.Debug("queryRrModelsShouldPush(): foundInSubscribedRrs, connAndRrKeys is empty:", connAndRrKeys)
			continue
		}

		rrModels, err := queryDb(pushRrModels[i].RrFullDomain, pushRrModels[i].RrType)
		if err != nil {
			belogs.Error("queryRrModelsShouldPush(): queryDb, fail, pushRrModels[i]:", jsonutil.MarshalJson(pushRrModels[i]), err)
			return nil, err
		}
		if len(rrModels) == 0 {
			belogs.Debug("queryRrModelsShouldPush(): queryDb, rrModels is empty, pushRrModels[i]:", jsonutil.MarshalJson(pushRrModels[i]))
			continue
		}
		belogs.Debug("queryRrModelsShouldPush(): queryDb, get rrModels:", jsonutil.MarshalJson(rrModels))

		for j := range connAndRrKeys {
			split := strings.Split(connAndRrKeys[j], "_")
			if len(split) != 2 {
				belogs.Error("queryRrModelsShouldPush(): connAndRrKeys split fail, connAndRrKeys[j]:", connAndRrKeys[j])
				return nil, errors.New("connAndRrKeys split fail")
			}
			connKey := split[0]
			rrKey := split[1]
			isDel := rr.IsDelTypeKey(rrKey)
			belogs.Debug("queryRrModelsShouldPush(): connAndRrKeys[j]:", connAndRrKeys[j], "  connKey:", connKey,
				"  rrKey:", rrKey, "  isDel:", isDel)
			if pushResultModel, ok := mapPushResultRrModels[connKey]; !ok {
				pushResultModel = pushmodel.NewPushResultModel(connKey)
				pushResultModel.AddRrModels(isDel, rrModels)
				belogs.Debug("queryRrModelsShouldPush():!ok, AddRrModels connKey:", connKey, "  pushResultModel:", jsonutil.MarshalJson(pushResultModel))
				mapPushResultRrModels[connKey] = pushResultModel
			} else {
				pushResultModel.AddRrModels(isDel, rrModels)
				belogs.Debug("queryRrModelsShouldPush():ok, AddRrModels connKey:", connKey, "  pushResultModel:", jsonutil.MarshalJson(pushResultModel))
				mapPushResultRrModels[connKey] = pushResultModel
			}
			belogs.Debug("queryRrModelsShouldPush(): connKey:", connKey, "  mapPushResultRrModels:", jsonutil.MarshalJson(mapPushResultRrModels))
		}
		belogs.Debug("PushCache.queryRrModelsShouldPush(): mapPushResultRrModels:", jsonutil.MarshalJson(mapPushResultRrModels))
	}

	pushResultRrModels = make([]*pushmodel.PushResultRrModel, 0)
	for connKey, pushResultModel := range mapPushResultRrModels {
		belogs.Debug("queryRrModelsShouldPush():range, connKey:", connKey, "  pushResultModel:", jsonutil.MarshalJson(pushResultModel))
		pushResultRrModels = append(pushResultRrModels, pushResultModel)
	}
	belogs.Debug("queryRrModelsShouldPush(): pushResultRrModels:", jsonutil.MarshalJson(pushResultRrModels))
	return pushResultRrModels, nil
}

func activePushAll() (pushResultRrModels []*pushmodel.PushResultRrModel, err error) {
	belogs.Debug("activePushAll():")
	pushResultRrModels = make([]*pushmodel.PushResultRrModel, 0)
	mapPushResultRrModels := make(map[string]*pushmodel.PushResultRrModel)

	connAndRrKeys, _ := pushCache.foundAllConnAndRrKeys()
	belogs.Debug("activePushAll(): foundAllConnAndRrKeys, connAndRrKeys:", connAndRrKeys)
	if len(connAndRrKeys) == 0 {
		belogs.Debug("activePushAll(): connAndRrKeys is empty")
		return pushResultRrModels, nil
	}
	rrModels, err := queryAllDb()
	if err != nil {
		belogs.Error("activePushAll(): queryAllDb, fail:", err)
		return nil, err
	}
	belogs.Debug("activePushAll(): queryDb, get rrModels:", jsonutil.MarshalJson(rrModels))

	for j := range connAndRrKeys {
		split := strings.Split(connAndRrKeys[j], "_")
		if len(split) != 2 {
			belogs.Error("activePushAll(): connAndRrKeys split fail, connAndRrKeys[j]:", connAndRrKeys[j])
			return nil, errors.New("connAndRrKeys split fail")
		}
		connKey := split[0]
		rrKey := split[1]
		isDel := rr.IsDelTypeKey(rrKey)
		belogs.Debug("activePushAll(): connAndRrKeys[j]:", connAndRrKeys[j], "  connKey:", connKey,
			"  rrKey:", rrKey, "  isDel:", isDel)
		if pushResultModel, ok := mapPushResultRrModels[connKey]; !ok {
			pushResultModel = pushmodel.NewPushResultModel(connKey)
			pushResultModel.AddRrModels(isDel, rrModels)
			belogs.Debug("activePushAll():!ok, AddRrModels connKey:", connKey, "  pushResultModel:", jsonutil.MarshalJson(pushResultModel))
			mapPushResultRrModels[connKey] = pushResultModel
		} else {
			pushResultModel.AddRrModels(isDel, rrModels)
			belogs.Debug("activePushAll():ok, AddRrModels connKey:", connKey, "  pushResultModel:", jsonutil.MarshalJson(pushResultModel))
			mapPushResultRrModels[connKey] = pushResultModel
		}
		belogs.Debug("queryRrModelsShouldPush(): connKey:", connKey, "  mapPushResultRrModels:", jsonutil.MarshalJson(mapPushResultRrModels))
	}

	for connKey, pushResultModel := range mapPushResultRrModels {
		belogs.Debug("activePushAll():range, connKey:", connKey, "  pushResultModel:", jsonutil.MarshalJson(pushResultModel))
		pushResultRrModels = append(pushResultRrModels, pushResultModel)
	}
	belogs.Debug("activePushAll(): pushResultRrModels:", jsonutil.MarshalJson(pushResultRrModels))
	return pushResultRrModels, nil
}

func getAllSubscribedRrs() map[string]uint16 {
	return pushCache.getAllSubscribedRrs()
}

/*
func sendToZoneFile(urlPath string, resourceRecord *dnsmodel.RrModel) {
	host := "https://" + conf.String("dns-server::serverHost") + ":" + conf.String("dns-server::serverHttpsPort")
	belogs.Info("sendToZoneFile(): start,  host:", host, "   json:", jsonutil.MarshalJson(resourceRecord))
	// send to zonefile
	go httpclient.Post(host+"/zonefile/"+urlPath, jsonutil.MarshalJson(resourceRecord), false)

}

func sendToDsoServer(urlPath string, rrsAndConnKeysModel RrsAndConnKeysModel) {
	host := "https://" + conf.String("dns-server::serverHost") + ":" + conf.String("dns-server::serverHttpsPort")
	belogs.Info("sendToZoneFile(): start,  host:", host, "   json:", jsonutil.MarshalJson(rrsAndConnKeysModel))
	// send to zonefile
	go httpclient.Post(host+"/dsoserver/"+urlPath, jsonutil.MarshalJson(rrsAndConnKeysModel), false)

}
*/

func getConnAndRrKeys(connKey string, rrFullDomain, rrType, rrClass string) (connAndRrKey, connAndRrAnyKey, connAndRrDelKey string) {
	// may rr is TYPE_ANY, also ok
	rrKey := rr.GetRrKey(rrFullDomain, rrType, rrClass)
	connAndRrKey = connKey + "_" + rrKey
	belogs.Debug("getConnAndRrKeys(): connKey:", connKey, "  rrKey:", rrKey, " connAndRrKey:", connAndRrKey)

	rrAnyKey := rr.GetRrKey(rrFullDomain, dnsutil.DNS_TYPE_STR_ANY, rrClass)
	connAndRrAnyKey = connKey + "_" + rrAnyKey
	belogs.Debug("getConnAndRrKeys(): connKey:", connKey, "  rrAnyKey:", rrAnyKey, " connAndRrAnyKey:", connAndRrAnyKey)

	rrDelKey := rr.GetRrKey(rrFullDomain, dnsutil.DNS_RR_DEL_KEY, rrClass)
	connAndRrDelKey = connKey + "_" + rrDelKey
	belogs.Debug("getConnAndRrKeys(): connKey:", connKey, "  rrDelKey:", rrDelKey, " connAndRrDelKey:", connAndRrDelKey)

	return connAndRrKey, connAndRrAnyKey, connAndRrDelKey
}
