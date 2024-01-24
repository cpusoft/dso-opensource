package drive

import (
	"dns-model/rr"
	"github.com/guregu/null"
)

type ServerDnsRrModel struct {
	Origin   string        `json:"origin"` // no end '.'
	Ttl      null.Int      `json:"ttl"`    // may be nil
	RrModels []*rr.RrModel `json:"rrModels"`
}

type ChainRepos struct {
	Repos []string `json:"repos"`
}

type RrDependentModel struct {
	Id           uint64   `json:"id"`
	RrFullDomain string   `json:"rrFullDomain"`
	ChildIds     []uint64 `json:"childIds"`
}
