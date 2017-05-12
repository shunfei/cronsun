package cronsun

import (
	"github.com/shunfei/cronsun/db"
)

var (
	mgoDB *db.Mdb
)

func GetDb() *db.Mdb {
	return mgoDB
}
