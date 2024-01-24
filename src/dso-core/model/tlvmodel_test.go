package model

import (
	"fmt"
	"testing"

	"dns-model/common"
	"github.com/cpusoft/goutil/dnsutil"
	"github.com/cpusoft/goutil/jsonutil"
)

func TestDsoModel(t *testing.T) {
	fmt.Println("")
	s := `
	{
		"headerForDsoModel": {
			"messageId": 10,
			"qrOpCodeZRCode": 45056,
			"qr": 1,
			"opCode": 6,
			"z": 0,
			"rCode": 0
		},
		"countQANAModel": {
			"qdCount": 0,
			"anCount": 0,
			"nsCount": 0,
			"arCount": 0
		},
		"dsoDataModel": {
			"tlvModels": [
				{
					"dsoType": 1,
					"dsoLength": 8,
					"inactivityTimeout": 1500,
					"keepaliveInterval": 1500
				}
			]
		}
	}
	`
	type TlvRrModel struct {
		DsoType   uint16      `json:"dsoType"`
		DsoLength uint16      `json:"dsoLength"`
		Data      interface{} `json:"data"`
	}

	type DsoDataRrModel struct {
		TlvModels []TlvRrModel `json:"tlvModels"`
	}

	type DsoRrModel struct {
		HeaderForDsoModel common.HeaderForDsoModel `json:"headerForDsoModel"`
		CountQANAModel    common.CountQANAModel    `json:"countQANAModel"`
		DsoDataModels     DsoDataRrModel           `json:"dsoDataModel"`
	}

	var responseDsoModel DsoRrModel
	err := jsonutil.UnmarshalJson(s, &responseDsoModel)
	fmt.Println("responseDsoModel:", jsonutil.MarshalJson(responseDsoModel), err)
	dsoDataRr := responseDsoModel.DsoDataModels
	dsoData := jsonutil.MarshalJson(dsoDataRr)
	fmt.Println("dsoData:", dsoData)
	for i := range dsoDataRr.TlvModels {
		tlv := dsoDataRr.TlvModels[i]
		fmt.Println(tlv)
		switch tlv.DsoType {
		case dnsutil.DSO_TYPE_KEEPALIVE:
			var k KeepaliveTlvModel
			err := jsonutil.UnmarshalJson(dsoData, &k)
			fmt.Println("k:", k, err)
		case dnsutil.DSO_TYPE_RETRY_DELAY:
		case dnsutil.DSO_TYPE_ENCRYPTION_PADDING:
		case dnsutil.DSO_TYPE_SUBSCRIBE:
		case dnsutil.DSO_TYPE_PUSH:
		case dnsutil.DSO_TYPE_UNSUBSCRIBE:
		case dnsutil.DSO_TYPE_RECONFIRM:
		}
	}
}
