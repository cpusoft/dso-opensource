package rr

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/cpusoft/goutil/convert"
	"github.com/cpusoft/goutil/dnsutil"
	"github.com/cpusoft/goutil/jsonutil"
)

func TestRr(t *testing.T) {
	dnsNameStr1 := `dwn1.roo.bo`
	domain1, _ := dnsutil.DomainStrToBytes(dnsNameStr1)
	dnsRData1 := []byte("1.1.1.1")
	rr1 := NewPacketModel(domain1, uint16(DNS_TYPE_A), uint16(DNS_CLASS_IN),
		uint32(100), uint16(len(dnsRData1)), dnsRData1)
	fmt.Println(jsonutil.MarshalJson(rr1))
	b1 := rr1.Bytes()
	fmt.Println(convert.PrintBytesOneLine(b1))

	dnsNameStr2 := `dwn2.roo.bo`
	domain2, _ := dnsutil.DomainStrToBytes(dnsNameStr2)
	dnsRData2 := []byte("2.2.2.2")
	rr2 := NewPacketModel(domain2, uint16(DNS_TYPE_A), uint16(DNS_CLASS_IN),
		uint32(200), uint16(len(dnsRData2)), dnsRData2)
	fmt.Println(jsonutil.MarshalJson(rr2))
	b2 := rr2.Bytes()
	fmt.Println(convert.PrintBytesOneLine(b2))

	dnsNameStr3 := `dwn3.roo.bo`
	domain3, _ := dnsutil.DomainStrToBytes(dnsNameStr3)
	dnsRData3 := []byte("3.3.3.")
	rr3 := NewPacketModel(domain3, uint16(DNS_TYPE_A), uint16(DNS_CLASS_IN),
		uint32(300), uint16(len(dnsRData3)), dnsRData3)
	fmt.Println(jsonutil.MarshalJson(rr3))
	b3 := rr3.Bytes()
	fmt.Println(convert.PrintBytesOneLine(b3))

	allRr := make([]*PacketModel, 0)
	allRr = append(allRr, rr1)
	allRr = append(allRr, rr2)
	allRr = append(allRr, rr3)
	fmt.Println(jsonutil.MarshalJson(allRr))

	allByte := make([]byte, 0)
	allByte = append(allByte, b1...)
	allByte = append(allByte, b2...)
	allByte = append(allByte, b3...)

	buf := bytes.NewReader(allByte)
	rrs, err := ParseToRrPacketModels(buf, -1)
	fmt.Println(jsonutil.MarshalJson(rrs), err)
}
