package models

import (
	"errors"
	"os/exec"

	"sunteng/commons/db/imgo"
	"sunteng/commons/log"
	"sunteng/cronsun/conf"
)

var (
	initialized bool

	needPassword = false
	SudoErr      = errors.New("sudo need password")
)

func InitPwd() {
	// 不支持 windows
	if _, err := exec.Command("sh", "-c", "echo |sudo -S echo &>/dev/null").Output(); err != nil {
		log.Warnf("当前用户 sudo 需要密码，所有指定用户执行的命令都将失败")
		needPassword = true
	}
}

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
