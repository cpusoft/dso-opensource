package model

import (
//	"bytes"
//	"fmt"
//	"testing"

// "github.com/cpusoft/goutil/convert"
// "github.com/cpusoft/goutil/dnsutil"
)

/*
func TestNewPushTlvModel(t *testing.T) {
	p := NewPushTlvModel()
	dnsName := `www.baidu.com` //
	dnsType := uint16(1)       // uint16,
	dnsClass := uint16(1)      // uint16,
	dnsTtl := uint32(3600)     //
	dnsRData := "182.61.200.7"
	dnsRdLen := uint16(len(dnsRData))
	o1 := dnsutil.NewResourceRecordModel([]byte(dnsName), dnsType, dnsClass,
		dnsTtl, dnsRdLen, []byte(dnsRData))
	p.AddDsoDataModel(o1)

	dnsName = `www.sina.com.cn` //
	dnsType = uint16(1)         // uint16,
	dnsClass = uint16(1)        // uint16,
	dnsTtl = uint32(3600)       //
	dnsRData = "219.238.4.9"
	dnsRdLen = uint16(len(dnsRData))
	o2 := dnsutil.NewResourceRecordModel([]byte(dnsName), dnsType, dnsClass,
		dnsTtl, dnsRdLen, []byte(dnsRData))
	p.AddDsoDataModel(o2)
	fmt.Println(convert.PrintBytes(p.Bytes(), 8))

	   	`00 41 00 47 //0x47 -->71
	   77 77 77 2e  // www.
	   62 61 69 64  //baid
	   75 2e 63 6f  //u.co
	   6d 00 01 00  //m     type:1
	   01 00 00 0e  // class 1,  ttl:0x0E10
	   10 00 0c 31  // rdlen 0x0c
	   38 32 2e 36  // 182.61.200.7
	   31 2e 32 30
	   30 2e 37 77  // 77 w
	   77 77 2e 73  // ww.s
	   69 6e 61 2e  // ina.
	   63 6f 6d 2e  // com.
	   63 6e 00 01  // cn type:1
	   00 01 00 00  // class 1,
	   0e 10 00 0b  // ttl 0x0E10  rdlen:11
	   32 31 39 2e
	   32 33 38 2e
	   34 2e 39`

}

func TestParseToPushTlvModel(t *testing.T) {
	//0x00, 0x41, 0x00, 0x47,
	s := []byte{0x77, 0x77, 0x77, 0x2e,
		0x62, 0x61, 0x69, 0x64, 0x75, 0x2e, 0x63, 0x6f,
		0x6d, 0x00, 0x01, 0x00, 0x01, 0x00, 0x00, 0x0e,
		0x10, 0x00, 0x0c, 0x31, 0x38, 0x32, 0x2e, 0x36,
		0x31, 0x2e, 0x32, 0x30, 0x30, 0x2e, 0x37, 0x77,
		0x77, 0x77, 0x2e, 0x73, 0x69, 0x6e, 0x61, 0x2e,
		0x63, 0x6f, 0x6d, 0x2e, 0x63, 0x6e, 0x00, 0x01,
		0x00, 0x01, 0x00, 0x00, 0x0e, 0x10, 0x00, 0x0b,
		0x32, 0x31, 0x39, 0x2e, 0x32, 0x33, 0x38, 0x2e,
		0x34, 0x2e, 0x39}
	//ss := strings.TrimSpace(s)
	//b, e := hex.DecodeString(ss)
	fmt.Println(s)
	buf := bytes.NewReader(s)
	p, e := ParseBytesToPushTlvModel(uint16(len(s)), buf)
	fmt.Println(p, e)
}
*/
