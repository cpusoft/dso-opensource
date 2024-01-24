package drive

import (
	"fmt"
	"testing"

	"dns-model/rr"
	"github.com/cpusoft/goutil/dnsutil"
	"github.com/cpusoft/goutil/jsonutil"
	"github.com/guregu/null"
)

func TestClientDnsRrModel(t *testing.T) {
	c := ClientDnsRrModel{}
	c.Origin = "example.com"
	c.Ttl = null.IntFrom(1000)
	c.RrModels = make([]*rr.RrModel, 0)

	r1 := rr.NewRrModel("example.com", "test1", dnsutil.DNS_TYPE_STR_A, dnsutil.DNS_CLASS_STR_IN,
		null.IntFrom(1000), "1.1.1.1")
	fmt.Println(r1)
	c.RrModels = append(c.RrModels, r1)
	r2 := rr.NewRrModel("example.com", "test1", dnsutil.DNS_TYPE_STR_TXT, dnsutil.DNS_CLASS_STR_IN,
		null.IntFrom(1000), "v=spf1 include:spf.mail.qq.com ip4:203.99.30.50 ~all")
	fmt.Println(r2)
	c.RrModels = append(c.RrModels, r2)
	str := jsonutil.MarshalJson(c)
	fmt.Println(str)
	/*
		{
			"origin": "example.com",
			"ttl": 1000,
			"rrModels": [
				{
					"id": 0,
					"originId": 0,
					"origin": "example.com.",
					"rrName": "test1",
					"rrFullDomain": "test1.example.com",
					"rrType": "A",
					"rrClass": "IN",
					"rrTtl": 1000,
					"rrData": "1.1.1.1",
					"updateTime": "0001-01-01T00:00:00Z"
				},
				{
					"id": 0,
					"originId": 0,
					"origin": "example.com.",
					"rrName": "test1",
					"rrFullDomain": "test1.example.com",
					"rrType": "TXT",
					"rrClass": "IN",
					"rrTtl": 1000,
					"rrData": "v=spf1 include:spf.mail.qq.com ip4:203.99.30.50 ~all",
					"updateTime": "0001-01-01T00:00:00Z"
				}
			]
		}


	*/

}
