package models

import (
	"sunteng/commons/db/imgo"
)

var (
	mgoDB *imgo.Mdb
)

func GetDb() *imgo.Mdb {
	return mgoDB
}
