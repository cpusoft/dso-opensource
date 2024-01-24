package clientcache

/*
var localResourceRecordCache *LocalResourceRecordCache

func init() {
	localResourceRecordCache = NewLocalResourceRecordCache()
}

type LocalResourceRecordCache struct {
	//
	KnownRrModels []*dnsmodel.RrModel `json:"knownRrModels"`

	// key: rrKey
	SubscribeRrModels map[string]*SubscribeRrModel `json:"subscribeResourceRecords"`
}

func NewLocalResourceRecordCache() *LocalResourceRecordCache {
	c := &LocalResourceRecordCache{}
	c.KnownRrModels = make([]*dnsmodel.RrModel, 0)
	c.SubscribeRrModels = make(map[string]*SubscribeRrModel)
	return c
}

type SubscribeRrModel struct {
	MessageId uint16            `json:"mssageId"`
	RrModel   *dnsmodel.RrModel `json:"resourceRecord"`
	// default:"",   "ok",  "fail"
	Result string `json:"result"`
	RCode  uint8  `json:"rCode"`
}

func NewSubscribeRrModel(messageId uint16, rrModel *dnsmodel.RrModel) *SubscribeRrModel {
	c := &SubscribeRrModel{
		MessageId: messageId,
		RrModel:   rrModel,
	}
	return c
}
*/
