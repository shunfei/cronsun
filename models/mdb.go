package models

import (
	"github.com/shunfei/cronsun/models/db"
)

var (
	mgoDB *db.Mdb
)

func GetDb() *db.Mdb {
	return mgoDB
}
