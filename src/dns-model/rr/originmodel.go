package rr

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/cpusoft/goutil/osutil"
	"github.com/guregu/null"
)

// for zonefile   // one originmodel --> one zonefile  // not support $include ;
type OriginModel struct {
	Id uint64 `json:"id" xorm:"id int"`
	// no have "." in the end
	Origin string `json:"origin" xorm:"origin varchar"` // lower
	// null.NewInt(0, false) or null.NewInt(i64, true)
	Ttl        null.Int  `json:"ttl" xorm:"dnsName int"`
	UpdateTime time.Time `json:"updateTime" xorm:"updateTime datetime"`

	RrModels []*RrModel `json:"rrModels"`
}

func NewOriginModel(origin string, ttl int64) *OriginModel {
	c := &OriginModel{}
	c.Origin = FormatRrOrigin(origin)
	c.Ttl = null.IntFrom(ttl)
	c.RrModels = make([]*RrModel, 0)
	return c
}

func (originModel *OriginModel) String() string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("%-10s%-20s", "$ORIGIN", originModel.Origin) + "." + osutil.GetNewLineSep())
	if originModel.Ttl.IsZero() {
		b.WriteString(osutil.GetNewLineSep())
	} else {
		b.WriteString(fmt.Sprintf("%-10s%-20s", "$TTL",
			strconv.Itoa(int(originModel.Ttl.ValueOrZero()))) + osutil.GetNewLineSep())
	}
	for i := range originModel.RrModels {
		b.WriteString(originModel.RrModels[i].String())
	}
	return b.String()
}
