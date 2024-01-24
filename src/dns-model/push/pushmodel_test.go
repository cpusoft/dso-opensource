package push

import (
	"fmt"
	"testing"

	"github.com/cpusoft/goutil/jsonutil"
)

func TestPushResultRrModel(t *testing.T) {
	s := `[
		{
			"connKey": "127.0.0.1:8125-127.0.0.1:42952",
			"rrModels": [
				{
					"id": 0,
					"originId": 0,
					"origin": "example.com.",
					"rrName": "dns1",
					"rrFullDomain": "dns1.example.com",
					"rrType": "A",
					"rrClass": "IN",
					"rrTtl": null,
					"rrData": "10.0.1.1",
					"updateTime": "0001-01-01T00:00:00Z"
				}
			]
		}
	]
	`
	pushResultRrModels := make([]*PushResultRrModel, 0)
	jsonutil.UnmarshalJson(s, &pushResultRrModels)
	for i := range pushResultRrModels {
		fmt.Println(pushResultRrModels[i])
	}
}
