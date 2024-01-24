package main

import (
	"github.com/cpusoft/goutil/belogs"
	_ "github.com/cpusoft/goutil/logs"
)

// go build -ldflags "-X main.CompileParamStr=****"
var CompileParamStr string

func main() {
	belogs.Info("main(): CompileParamStr:", CompileParamStr)
	if CompileParamStr == "" || CompileParamStr == "dso-server" {
		startDnsServer()
	} else if CompileParamStr == "dso-client" {
		startDnsClient()
	}
}
