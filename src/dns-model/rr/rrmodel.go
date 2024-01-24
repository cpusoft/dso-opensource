package rr

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/dnsutil"
	"github.com/cpusoft/goutil/osutil"
	"github.com/guregu/null"
)

// for mysql, zonefile
type RrModel struct {
	Id       uint64 `json:"id" xorm:"id int"`
	OriginId uint64 `json:"originId" xorm:"originId int"`

	// not have "." in the end
	Origin string `json:"origin" xorm:"origin varchar"` // lower
	// is host/subdomain, not have "." int the end
	// if no subdomain, is "", not "@"
	RrName string `json:"rrName" xorm:"rrName varchar"` // lower
	// == rrName+.+Origin
	RrFullDomain string `json:"rrFullDomain" xorm:"rrFullDomain varchar"` // lower: rrName+"."+Origin[-"."]

	RrType  string `json:"rrType" xorm:"rrType varchar"`   // upper
	RrClass string `json:"rrClass" xorm:"rrClass varchar"` // upper
	// null.NewInt(0, false) or null.NewInt(i64, true)
	RrTtl  null.Int `json:"rrTtl" xorm:"rrTtl int"`
	RrData string   `json:"rrData" xorm:"rrData varchar"`

	UpdateTime time.Time `json:"updateTime" xorm:"updateTime datetime"`
}

//
func NewRrModel(origin, rrName, rrType, rrClass string,
	rrTtl null.Int, rrData string) (rrModel *RrModel) {
	rrModel = &RrModel{
		Origin:  FormatRrOrigin(origin),
		RrName:  FormatRrName(rrName),
		RrType:  FormatRrClassOrRrType(rrType),
		RrClass: FormatRrClassOrRrType(rrClass),
		RrTtl:   rrTtl,
		RrData:  rrData,
	}
	if rrModel.RrName == "" {
		rrModel.RrFullDomain = rrModel.Origin
	} else if rrModel.RrName == "@" {
		rrModel.RrFullDomain = rrModel.Origin
	} else {
		rrModel.RrFullDomain = rrModel.RrName + "." + rrModel.Origin
	}
	return rrModel
}

func NewRrModelByFullDomain(rrFullDomain, rrType, rrClass string,
	rrTtl null.Int, rrData string) (rrModel *RrModel) {
	rrModel = &RrModel{
		RrFullDomain: FormatRrFullDomain(rrFullDomain),
		RrType:       FormatRrClassOrRrType(rrType),
		RrClass:      FormatRrClassOrRrType(rrClass),
		RrTtl:        rrTtl,
		RrData:       rrData,
	}
	return rrModel
}

func FormatRrOrigin(t string) string {
	// should not have "." as end
	return FormatRrName(t)
}
func FormatRrFullDomain(t string) string {
	// should not have "." as end
	return FormatRrName(t)
}

// //  remove origin, only hostname, and will remove the end "." // lower
func FormatRrName(t string) string {
	if len(t) == 0 {
		return t
	}
	return strings.TrimSuffix(strings.TrimSpace(strings.ToLower(t)), ".")
}

func FormatRrClassOrRrType(t string) string {
	return strings.TrimSpace(strings.ToUpper(t))
}

// not include Domain:
func (rrModel *RrModel) String() string {
	var b strings.Builder
	b.Grow(128)

	ttl := ""
	if !rrModel.RrTtl.IsZero() {
		ttl = strconv.Itoa(int(rrModel.RrTtl.ValueOrZero()))
	}
	var space string
	b.WriteString(fmt.Sprintf("%-20s%-6s%-4s%-6s%-4s", rrModel.RrName, ttl, rrModel.RrClass, rrModel.RrType, space))
	if rrModel.RrType == "SOA" {
		split := strings.Split(rrModel.RrData, " ")
		if len(split) == 7 {
			b.WriteString(split[0] + " " + split[1] + " ( " +
				split[2] + " " + split[3] + " " + split[4] +
				split[5] + " " + split[6] + " )")
		}
	} else {
		b.WriteString(rrModel.RrData)
	}
	b.WriteString(osutil.GetNewLineSep())
	return b.String()
}

// get rrkey:
func GetRrKey(rrFullDomain, rrType, rrClass string) string {
	rrKey := rrFullDomain + "#" + rrType + "#" + rrClass
	belogs.Debug("GetRrKey():rrKey:", rrKey)
	return rrKey
}

// get rrkey:
func GetRrModelKey(rrModel *RrModel) string {
	if rrModel == nil {
		return ""
	}
	rrKey := GetRrKey(rrModel.RrFullDomain, rrModel.RrType, rrModel.RrClass)
	belogs.Debug("GetRrModelKey():GetRrKey rrKey:", rrKey)
	return rrKey
}

// get rrAnyKey:
func GetRrModelAnyTypeKey(rrModel *RrModel) string {
	if rrModel == nil {
		return ""
	}
	rrAnyKey := GetRrKey(rrModel.RrFullDomain, dnsutil.DNS_TYPE_STR_ANY, rrModel.RrClass)
	belogs.Debug("GetRrModelAnyTypeKey():GetRrKey rrAnyKey:", rrAnyKey)
	return rrAnyKey
}

// get rrDelKey
func GetRrModelDelTypeKey(rrModel *RrModel) string {
	if rrModel == nil {
		return ""
	}
	rrDelKey := GetRrKey(rrModel.RrFullDomain, dnsutil.DNS_RR_DEL_KEY, rrModel.RrClass)
	belogs.Debug("GetRrModelDelTypeKey():GetRrKey rrDelKey:", rrDelKey)
	return rrDelKey
}

func IsDelTypeKey(rrKey string) bool {
	return strings.Contains(rrKey, "#"+dnsutil.DNS_RR_DEL_KEY+"#")
}

func IsDelRrModelForDso(rrTtl null.Int) bool {
	belogs.Debug("IsDelRrModelForDso():rrTtl:", rrTtl)
	if rrTtl.ValueOrZero() == dnsutil.DSO_DEL_SPECIFIED_RESOURCE_RECORD_TTL ||
		rrTtl.ValueOrZero() == dnsutil.DSO_DEL_COLLECTIVE_RESOURCE_RECORD_TTL {
		return true
	}
	return false
}
func EqualRrModel(leftRr, rightRr *RrModel) bool {
	if leftRr == nil || rightRr == nil {
		return false
	}
	// not compare ttl
	if leftRr.Origin == rightRr.Origin &&
		leftRr.RrName == rightRr.RrName &&
		leftRr.RrType == rightRr.RrType &&
		leftRr.RrData == rightRr.RrData {
		return true
	}
	return false
}
