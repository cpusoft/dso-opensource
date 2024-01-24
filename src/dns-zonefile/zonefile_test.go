package zonefile

import (
	"fmt"
	"testing"

	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/xormdb"
)

func TestImportZoneFile(t *testing.T) {
	// start mysql
	err := xormdb.InitMySql()
	if err != nil {
		belogs.Error("startWebServer(): start InitMySql failed:", err)
		fmt.Println("rpstir2 failed to start, ", err)
		return
	}
	defer xormdb.XormEngine.Close()

	f := "example.com.zone"
	err = importZoneFile(f)
	fmt.Println(f, err)
}
