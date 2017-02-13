package models

import (
	"sunteng/commons/db/imgo"

	"sunteng/cronsun/conf"
)

var (
	initialized bool
)

func Init() (err error) {
	if initialized {
		return
	}

	// init id creator
	if err = initID(); err != nil {
		return
	}

	// init config
	if err = conf.Init(); err != nil {
		return
	}

	// init etcd client
	if DefalutClient, err = NewClient(conf.Config); err != nil {
		return
	}

	// init mongoDB
	mgoDB = imgo.NewMdbWithConf(conf.Config.Mgo)

	initialized = true
	return
}
