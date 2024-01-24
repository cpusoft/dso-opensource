package clientcache

import (
	"fmt"
	"testing"

	"dns-model/rr"
	_ "github.com/cpusoft/goutil/conf"
	"github.com/cpusoft/goutil/dnsutil"
	"github.com/cpusoft/goutil/jsonutil"
	"github.com/guregu/null"
)

func TestInitCache(t *testing.T) {
	err := InitCache()
	fmt.Println("InitCache():", err)

	err = ResetCache()
	fmt.Println("ResetTable():", err)

	rr := rr.NewRrModelByFullDomain("test1.example1.com", dnsutil.DNS_TYPE_STR_A, dnsutil.DNS_CLASS_STR_IN,
		null.IntFrom(1000), "1.1.1.1")
	err = UpdateRrModel(rr, false)
	fmt.Println("UpdateRrModel():", jsonutil.MarshalJson(rr), err)

	rrs, err := QueryRrModels(rr)
	fmt.Println("QueryRrModels():", jsonutil.MarshalJson(rrs), err)

	messageId, err := GetNewMessageId(dnsutil.DNS_OPCODE_QUERY)
	fmt.Println("GetNewMessageId():", messageId, err)
}
