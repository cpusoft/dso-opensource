package convert

import (
	"fmt"
	"strings"
	"testing"

	"dns-model/rr"
	"github.com/cpusoft/goutil/dnsutil"
	"github.com/cpusoft/goutil/jsonutil"
	"github.com/guregu/null"
)

func TestNewRrModel(t *testing.T) {
	fullDomain := []byte("example.com")
	zName := "example.com"
	rrName := strings.TrimSuffix(string(fullDomain), zName)
	rrData := ""
	fmt.Println("ConvertPacketToRr(): fullDomain:", string(fullDomain), "  zName:", zName,
		" rrName:", rrName)
	rrModel := rr.NewRrModel(rr.FormatRrOrigin(zName), rr.FormatRrName(rrName),
		dnsutil.DnsIntTypes[1], dnsutil.DnsIntClasses[1],
		null.IntFrom(int64(1)), rrData)
	fmt.Println("ConvertPacketToRr(): rrModel:", jsonutil.MarshalJson(rrModel))
}
