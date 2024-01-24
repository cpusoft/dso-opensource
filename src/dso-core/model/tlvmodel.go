package model

type TlvModel interface {
	Bytes() []byte
	PrintBytes() string
	GetDsoType() uint16
}
