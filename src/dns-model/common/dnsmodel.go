package common

type DnsModel interface {
	GetHeaderModel() HeaderModel
	GetCountModel() CountModel
	GetDataModel() interface{}
	Bytes() []byte
	GetDnsModelType() string //"packet","rr"
}
