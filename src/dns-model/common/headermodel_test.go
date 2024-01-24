package common

import (
	"fmt"
	"testing"

	"github.com/cpusoft/goutil/convert"
	"github.com/cpusoft/goutil/dnsutil"
	"github.com/cpusoft/goutil/jsonutil"
)

func TestQrOpCodeZRCode(t *testing.T) {
	qr := uint8(dnsutil.DNS_QR_RESPONSE)
	opCode := uint8(dnsutil.DNS_OPCODE_DSO)
	rCode := uint8(dnsutil.DNS_RCODE_DSOTYPENI)
	qrOpCodeZRCode := ComposeQrOpCodeZRCode(qr, opCode, rCode)
	fmt.Println(qrOpCodeZRCode)

	qr1, opCode1, z, rCode1, _ := ParseQrOpCodeZRCode(qrOpCodeZRCode)
	fmt.Println(qr1, opCode1, z, rCode1)
}

func TestNewHeaderModel(t *testing.T) {
	h, err := NewHeaderModel(0, 0, DNS_HEADER_TYPE_DSO)
	fmt.Println(jsonutil.MarshalJson(h), err)
	//hdso, _ := h.(HeaderForDsoModel)
	dso := NewDsoModel(h)
	fmt.Println(jsonutil.MarshalJson(dso))
}

type DsoModel struct {
	HeaderForDsoModel HeaderForDsoModel `json:"headerForDsoModel"`
	CountQANAModel    *CountQANAModel   `json:"countQANAModel"`
}

func NewDsoModel(headerModel HeaderModel) *DsoModel {
	c := &DsoModel{}
	c.HeaderForDsoModel, _ = headerModel.(HeaderForDsoModel)

	return c
}

func TestHeaderAndCountModel(t *testing.T) {
	//receiveBytes := []byte{0x2f, 0x10, 0x81, 0x80, 0x00, 0x01, 0x00, 0x01, 0x00, 0x04, 0x00, 0x11}
	receiveBytes := []byte{0xda, 0x3c,
		0x01, 0x20,
		0x00, 0x01,
		0x00, 0x00,
		0x00, 0x00,
		0x00, 0x01}
	headerType, headerModel, err := ParseBytesToHeaderModel(receiveBytes[:DNS_HEADER_LENGTH])
	if err != nil {
		fmt.Println("TestHeaderAndCountModel(): ParseBytesToHeaderModel fail:", convert.PrintBytesOneLine(receiveBytes[:DNS_HEADER_LENGTH]))
		return
	}
	offsetFromStart := uint16(DNS_HEADER_LENGTH)
	fmt.Println("TestHeaderAndCountModel(): headerType:", headerType, "   headerModel:", jsonutil.MarshalJson(headerModel), "  offsetFromStart:", offsetFromStart)

	// count
	countModel, err := ParseBytesToCountModel(receiveBytes[offsetFromStart:offsetFromStart+DNS_COUNT_LENGTH], headerType)
	if err != nil {
		fmt.Println("TestHeaderAndCountModel(): ParseBytesToCountModel fail:", convert.PrintBytesOneLine(receiveBytes[:DNS_HEADER_LENGTH]))
		return
	}
	offsetFromStart += uint16(DNS_COUNT_LENGTH)
	fmt.Println("TestHeaderAndCountModel(): headerType:", headerType, "   countModel:", jsonutil.MarshalJson(countModel), "  offsetFromStart:", offsetFromStart)

}
