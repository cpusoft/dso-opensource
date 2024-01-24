package packet

// shaodebug
type PacketData interface {
	Bytes() []byte
	Length() uint16
	ToRrData() string
}
