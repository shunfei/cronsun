package conf

import (
	"path"
	"time"

	client "github.com/coreos/etcd/clientv3"

	"sunteng/commons/confutil"
	"sunteng/commons/log"
	"sunteng/commons/util"
)

var (
	Config = new(Conf)
)

func Init() error {
	Config.Root = util.CurDir()

	confFile := path.Join(Config.Root, "files", "base.json")
	err := confutil.LoadExtendConf(confFile, Config)
	if err != nil {
		return err
	}

	if Config.Etcd.DialTimeout > 0 {
		Config.Etcd.DialTimeout *= time.Second
	}
	log.InitConf(&Config.Log)

	return nil
}

type Conf struct {
	Root string // 项目根目录

	Proc string // proc 路径
	Cmd  string // cmd 路径
	Sep  string // etcd key 的连接符

	Ttl int64 // 节点超时时间，单位秒

	Log  log.Config
	Etcd client.Config
}
