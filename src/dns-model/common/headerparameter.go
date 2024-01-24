package common

import (
	"errors"
	"fmt"

	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/dnsutil"
)

type QrOpCodeZRCode uint16

func ComposeQrOpCodeZRCode(qr uint8, opCode uint8, rCode uint8) (qrOpCodeZRCode QrOpCodeZRCode) {
	q := uint16(0)
	if qr == dnsutil.DNS_QR_RESPONSE {
		q = (uint16(dnsutil.DNS_QR_RESPONSE) << 15)
	}
	q |= uint16(opCode) << uint16(11)
	q |= uint16(rCode)
	qrOpCodeZRCode = QrOpCodeZRCode(q)
	belogs.Debug("ComposeQrOpCodeZRCode():qr:", qr, " opCode:", opCode, "  rCode:", rCode, "  qrOpCodeZRCode:", qrOpCodeZRCode)
	return qrOpCodeZRCode
}

func ParseQrOpCodeZRCode(qrOpCodeZRCode QrOpCodeZRCode) (qr, opCode, z, rCode uint8, err error) {
	qr = uint8((qrOpCodeZRCode >> 15) & 1)        // 1000 0000 0000 0000 --> 1
	opCode = uint8((qrOpCodeZRCode >> 11) & 0x0f) // 0111 1000 0000 0000 --> 1111
	z = uint8((qrOpCodeZRCode >> 4) & 0x7f)       // 0000 0111 1111 0000 --> 111 1111
	rCode = uint8(qrOpCodeZRCode & 0x0f)          // 0000 0000 0000 1111 --> 1111
	belogs.Debug("ParseQrOpCodeZRCode():qrOpCodeZRCode:", qrOpCodeZRCode, " 0x:", fmt.Sprintf("%0x", qrOpCodeZRCode), "  qr:", qr, " opCode:", opCode,
		"  z:", z, "  rCode:", rCode)
	err = checkQr(qr)
	if err != nil {
		belogs.Error("ParseQrOpCodeZRCode(): checkQr fail, qr:", qr)
		return 0, 0, 0, 0, err
	}
	err = checkZ(z)
	if err != nil {
		belogs.Error("ParseQrOpCodeZRCode(): checkZ fail, z:", z)
		return 0, 0, 0, 0, err
	}
	err = checkRCode(rCode)
	if err != nil {
		belogs.Error("ParseQrOpCodeZRCode(): checkRCode fail, rCode:", rCode)
		return 0, 0, 0, 0, err
	}
	return qr, opCode, z, rCode, nil
}

type QrOpCodeAaTcRdRaZRCode uint16

func ComposeQrOpCodeAaTcRdRaZRCode(qr, opCode, aa, tc, rd, ra, rCode uint8) (qrOpCodeAaTcRdRaZRCode QrOpCodeAaTcRdRaZRCode) {
	q := uint16(0)
	if qr == dnsutil.DNS_QR_RESPONSE {
		q = (uint16(dnsutil.DNS_QR_RESPONSE) << 15)
	}
	q |= uint16(opCode) << uint16(11)
	q |= uint16(aa) << uint16(10)
	q |= uint16(tc) << uint16(9)
	q |= uint16(rd) << uint16(8)
	q |= uint16(ra) << uint16(7)
	q |= uint16(rCode)
	qrOpCodeAaTcRdRaZRCode = QrOpCodeAaTcRdRaZRCode(q)
	belogs.Debug("ComposeQrOpCodeAaTcRdRaZRCode():qr:", qr, " opCode:", opCode,
		"  aa:", aa, " tc:", tc, "  rd:", rd, " ra:", ra, "  rCode:", rCode, "  qrOpCodeZRCode:", qrOpCodeAaTcRdRaZRCode)
	return qrOpCodeAaTcRdRaZRCode
}

func ParseQrOpCodeAaTcRdRaZRCode(qrOpCodeAaTcRdRaZRCode QrOpCodeAaTcRdRaZRCode) (qr, opCode, aa, tc, rd, ra, z, rCode uint8, err error) {
	qr = uint8((qrOpCodeAaTcRdRaZRCode >> 15) & 1)        // 1000 0000 0000 0000 --> 1
	opCode = uint8((qrOpCodeAaTcRdRaZRCode >> 11) & 0x0f) // 0111 1000 0000 0000 --> 1111
	aa = uint8((qrOpCodeAaTcRdRaZRCode >> 10) & 0x01)     // 0000 0100 0000 0000 --> 1
	tc = uint8((qrOpCodeAaTcRdRaZRCode >> 9) & 0x01)      // 0000 0010 0000 0000 --> 1
	rd = uint8((qrOpCodeAaTcRdRaZRCode >> 8) & 0x01)      // 0000 0001 0000 0000 --> 1
	ra = uint8((qrOpCodeAaTcRdRaZRCode >> 7) & 0x01)      // 0000 0000 1000 0000 --> 1
	z = uint8((qrOpCodeAaTcRdRaZRCode >> 4) & 0x07)       // 0000 0000 0111 0000 --> 111 0000
	rCode = uint8(qrOpCodeAaTcRdRaZRCode & 0x0f)          // 0000 0000 0000 1111 --> 1111
	belogs.Debug("ParseQrOpCodeZRCode():qrOpCodeZRCode:", qrOpCodeAaTcRdRaZRCode, " 0x:", fmt.Sprintf("%0x", qrOpCodeAaTcRdRaZRCode),
		"  qr:", qr, " opCode:", opCode, "  aa:", aa, " tc:", tc, " ra:", ra,
		"  z:", z, "  rCode:", rCode)
	err = checkQr(qr)
	if err != nil {
		belogs.Error("ParseQrOpCodeZRCode(): checkQr fail, qr:", qr)
		return 0, 0, 0, 0, 0, 0, 0, 0, err
	}
	err = checkZ(z)
	if err != nil {
		belogs.Error("ParseQrOpCodeZRCode(): checkZ fail, z:", z)
		return 0, 0, 0, 0, 0, 0, 0, 0, err
	}
	err = checkRCode(rCode)
	if err != nil {
		belogs.Error("ParseQrOpCodeZRCode(): checkRCode fail, rCode:", rCode)
		return 0, 0, 0, 0, 0, 0, 0, 0, err
	}
	return qr, opCode, aa, tc, rd, ra, z, rCode, nil
}
func checkQr(qr uint8) error {
	// check qr
	if qr != dnsutil.DNS_QR_REQUEST && qr != dnsutil.DNS_QR_RESPONSE {
		belogs.Error("checkQr():qr is not 0(DNS_QR_REQUEST) nor 1(DNS_QR_RESPONSE),  qr:", qr)
		return errors.New("qr is not 0(DNS_QR_REQUEST) nor 1(DNS_QR_RESPONSE)")
	}
	return nil
}

func checkZ(z uint8) error {
	// check z
	belogs.Debug("checkZ(): z:", z)
	if z != 0 {
		if z == 0x02 { // 00 10 : is 'ad bit' https://blog.csdn.net/qq_37907693/article/details/120694795
			belogs.Debug("checkZ(): is 0x02:, is ad bit,is ok")
			return nil
		} else if z == 0x01 { // is non-authenticated data: unacdeptable  http://c.biancheng.net/view/6457.html
			belogs.Debug("checkZ(): is 0x02:, is non-authenticated data,is ok")
			return nil
		}
		belogs.Error("checkZ():z is not 0, fail, z:", z)
		return errors.New("z is not 0")
	}
	return nil
}

func checkRCode(rCode uint8) error {
	// check rCode
	_, ok := dnsutil.DnsRCodes[rCode]
	if !ok {
		belogs.Error("ParseBytesToHeaderModel():rCode is not a legal value,fail, rCode:", rCode)
		return errors.New("RCode is not a legal value")
	}
	return nil
}
