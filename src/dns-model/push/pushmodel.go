package push

import (
	"dns-model/rr"
	"github.com/cpusoft/goutil/dnsutil"
	"github.com/guregu/null"
)

// same from dns-server
type PushRrModel struct {
	ConnKey            string   `json:"connKey"`
	RrFullDomain       string   `json:"rrDomain"`
	RrType             string   `json:"rrType"`
	RrClass            string   `json:"rrClass"`
	RrTtl              null.Int `json:"rrTtl"`
	RrData             string   `json:"rrData"`
	SubscribeMessageId uint16   `json:"subscribeMessageId"`
}

func NewPushRrModel(connKey, rrFullDomain, rrType, rrClass string,
	rrTtl null.Int, rrData string, subscribeMessageId uint16) *PushRrModel {
	c := &PushRrModel{
		ConnKey:            connKey,
		RrFullDomain:       rrFullDomain,
		RrType:             rrType,
		RrClass:            rrClass,
		RrTtl:              rrTtl,
		RrData:             rrData,
		SubscribeMessageId: subscribeMessageId,
	}
	return c
}

// one conn key ,may have RrModels
type PushResultRrModel struct {
	ConnKey  string        `json:"connKey"`
	RrModels []*rr.RrModel `json:"rrModels"`
}

func NewPushResultModel(connKey string) *PushResultRrModel {
	c := &PushResultRrModel{}
	c.ConnKey = connKey
	c.RrModels = make([]*rr.RrModel, 0)
	return c
}

// rrType is DEL , ttl -->DSO_DEL_SPECIFIED_RESOURCE_RECORD_TTL
func (c *PushResultRrModel) AddRrModel(isDel bool, rrModel *rr.RrModel) {
	found := false
	for i := range c.RrModels {
		if rr.EqualRrModel(rrModel, c.RrModels[i]) {
			found = true
			break
		}
	}
	if found {
		return
	}
	if isDel {
		rrModel.RrTtl = null.IntFrom(dnsutil.DSO_DEL_SPECIFIED_RESOURCE_RECORD_TTL)
	}
	c.RrModels = append(c.RrModels, rrModel)
}
func (c *PushResultRrModel) AddRrModels(isDel bool, rrModels []*rr.RrModel) {
	for i := range rrModels {
		c.AddRrModel(isDel, rrModels[i])
	}
}

type UnpushRrModel struct {
	ConnKey            string `json:"connKey,omitempty"`
	SubscribeMessageId uint16 `json:"subscribeMessageId"`
}

func NewUnpushRrModel(connKey string, subscribeMessageId uint16) *UnpushRrModel {
	c := &UnpushRrModel{}
	c.ConnKey = connKey
	c.SubscribeMessageId = subscribeMessageId
	return c
}
