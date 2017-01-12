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
	Config      = new(Conf)
	initialized bool
)

func Init() error {
	if initialized {
		return nil
	}

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

	Config.Cmd = cleanKeyPrefix(Config.Cmd)
	Config.Proc = cleanKeyPrefix(Config.Proc)
	Config.Group = cleanKeyPrefix(Config.Group)

	Config.Root = path.Join(Config.Root, "..")

	initialized = true
	return nil
}

type Conf struct {
	Root string // 项目根目录

	Proc  string // proc 路径
	Cmd   string // cmd 路径
	Group string // 节点分组

	Ttl        int64 // 节点超时时间，单位秒
	ReqTimeout int   // 请求超时时间，单位秒

	Log  log.Config
	Etcd client.Config
	Web  webConfig
}

type webConfig struct {
	BindAddr string
	UIDir    string
}

// 返回前后包含斜杆的 /a/b/ 的前缀
func cleanKeyPrefix(p string) string {
	p = path.Clean(p)
	if p[0] != '/' {
		p = "/" + p
	}

	p += "/"

	return p
}
