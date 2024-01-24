package sys

type SysStyle struct {
	// "init" :  will drop and create all table;
	// "resetall" will remove all data ;
	SysStyle string `json:"sysStyle"`
}
