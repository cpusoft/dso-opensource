package drive

import (
	"dns-model/rr"
	"github.com/guregu/null"
)

type ClientDnsRrModel struct {
	Origin   string        `json:"origin"` // no end '.'
	Ttl      null.Int      `json:"ttl"`    // may be nil
	RrModels []*rr.RrModel `json:"rrModels"`
}

func NewClientDnsRrModel(origin string, ttl null.Int) (c *ClientDnsRrModel) {
	return &ClientDnsRrModel{
		Origin:   origin,
		Ttl:      ttl,
		RrModels: make([]*rr.RrModel, 0),
	}
}

func (c *ClientDnsRrModel) AddRrModel(rrModel *rr.RrModel) {
	c.RrModels = append(c.RrModels, rrModel)
}

func CloneClientDnsRrModel(c *ClientDnsRrModel) *ClientDnsRrModel {
	c1 := &ClientDnsRrModel{
		Origin:   c.Origin,
		Ttl:      c.Ttl,
		RrModels: make([]*rr.RrModel, len(c.RrModels)),
	}
	copy(c1.RrModels, c.RrModels)
	return c1
}

type ClientKeepaliveModel struct {
	InactivityTimeout uint32 `json:"inactivityTimeout"`
	KeepaliveInterval uint32 `json:"keepaliveInterval"`
}

type PreceptRpki struct {
	PreceptId          string              `json:"preceptId"`
	PreceptRpkiDomains []PreceptRpkiDomain `json:"preceptRpkiDomains"`
}
type PreceptRpkiDomain struct {
	NotifyUrl string `json:"notifyUrl"`
	SessionId string `json:"sessionId"`
	MaxSerial uint64 `json:"maxSerial"`
	MinSerial uint64 `json:"minSerial"`

	PreceptRpkiSnapshot PreceptRpkiSnapshot `json:"snapshot"`
	PreceptRpkiDeltas   []PreceptRpkiDelta  `json:"deltas"`
}

type PreceptRpkiSnapshot struct {
	SnapshotUrl string `json:"snapshotUrl"`
}
type PreceptRpkiDelta struct {
	DeltaUrl string `json:"deltaUrl"  xorm:"deltaUrl varchar(255)"`
	Serial   uint64 `json:"serial"  xorm:"serial int"`
}

const (
	PRECEPT_RPKI_RRDP_TYPE_SNASHOT_PUBLISH = "rrdpTypeSnapshotPublish"
	PRECEPT_RPKI_RRDP_TYPE_DELTA_PUBLISH   = "rrdpTypeDeltaPublish"
	PRECEPT_RPKI_RRDP_TYPE_DELTA_WITHDRAW  = "rrdpTypeDeltaWithdraw"
)

type PreceptRpkiSnapshotPublishBase64 struct {
	RrdpType    string `json:"rrdpType"` // PRECEPT_RPKI_RRDP_TYPE_SNASHOT_PUBLISH
	NotifyUrl   string `json:"notifyUrl"`
	SessionId   string `json:"sessionId"`
	Serial      uint64 `json:"serial"`
	SnapshotUrl string `json:"snapshotUrl"`
	Url         string `json:"url"`
	Base64      string `json:"base64"`
}

type PreceptRpkiDeltaPublishBase64 struct {
	RrdpType  string `json:"rrdpType"` // PRECEPT_RPKI_RRDP_TYPE_DELTA_PUBLISH
	NotifyUrl string `json:"notifyUrl"`
	SessionId string `json:"sessionId"`
	Serial    uint64 `json:"serial"`
	DeltaUrl  string `json:"deltaUrl"`
	Url       string `json:"url"`
	Hash      string `json:"hash"`
	Base64    string `json:"base64"`
}

type PreceptRpkiDeltaWithdrawBase64 struct {
	RrdpType  string `json:"rrdpType"` // PRECEPT_RPKI_RRDP_TYPE_DELTA_WITHDRAW
	NotifyUrl string `json:"notifyUrl"`
	SessionId string `json:"sessionId"`
	Serial    uint64 `json:"serial"`
	DeltaUrl  string `json:"deltaUrl"`
	Url       string `json:"url"`
	Hash      string `json:"hash"`
}
