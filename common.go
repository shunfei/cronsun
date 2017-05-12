package cronsun

import (
	"github.com/shunfei/cronsun/conf"
	"github.com/shunfei/cronsun/db"
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
	if mgoDB, err = db.NewMdb(conf.Config.Mgo); err != nil {
		return
	}

	initialized = true
	return
}
