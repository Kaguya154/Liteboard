package api

import "github.com/Kaguya154/dbhelper/types"

var db types.Conn

func SetDB(conn types.Conn) {
	db = conn
}
