package internal

import (
	"github.com/Kaguya154/dbhelper"
	"github.com/Kaguya154/dbhelper/drivers/sqlite"
)

func init() {
	err := dbhelper.RegisterDriver(sqlite.DriverName, sqlite.GetDriver())
	if err != nil {
		return
	}
}
