package model

import (
//	"bytes"
//	"fmt"
//	"testing"

// "github.com/cpusoft/goutil/convert"
)

/*
func TestNewReconfirmTlvModel(t *testing.T) {

	dnsName := `www.baidu.com` //
	dnsType := uint16(1)       // uint16,
	dnsClass := uint16(1)      // uint16,
	dnsRData := "182.61.200.7"
	p := NewReconfirmTlvModel([]byte(dnsName), dnsType, dnsClass,
		[]byte(dnsRData))

	fmt.Println(convert.PrintBytes(p.Bytes(), 8))

		00 40 00 1d 77 77 77 2e
		62 61 69 64 75 2e 63 6f
		6d 00 01 00 01 31 38 32
		2e 36 31 2e 32 30 30 2e
		37

}
func TestParseToReconfirmTlvModel(t *testing.T) {
	//0x00, 0x40, 0x00, 0x1d,
	s := []byte{0x77, 0x77, 0x77, 0x2e,
		0x62, 0x61, 0x69, 0x64, 0x75, 0x2e, 0x63, 0x6f,
		0x6d, 0x00, 0x01, 0x00, 0x01, 0x31, 0x38, 0x32,
		0x2e, 0x36, 0x31, 0x2e, 0x32, 0x30, 0x30, 0x2e,
		0x37}
	fmt.Println(s)
	buf := bytes.NewReader(s)
	p, e := ParseBytesToReconfirmTlvModel(uint16(len(s)), buf)
	fmt.Println(p, e)
}

*/
