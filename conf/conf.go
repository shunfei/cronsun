package conf

import (
	"flag"
	"path"
	"time"

	client "github.com/coreos/etcd/clientv3"
	"github.com/fsnotify/fsnotify"
	"github.com/go-gomail/gomail"

	"github.com/shunfei/cronsun/db"
	"github.com/shunfei/cronsun/event"
	"github.com/shunfei/cronsun/log"
	"github.com/shunfei/cronsun/utils"
)

var (
	confFile = flag.String("conf",
		"conf/files/base.json", "config file path")

	Config      = new(Conf)
	initialized bool

	watcher  *fsnotify.Watcher
	exitChan = make(chan struct{})
)

func Init() error {
	if initialized {
		return nil
	}

	flag.Parse()
	if err := Config.parse(); err != nil {
		return err
	}
	if err := Config.watch(); err != nil {
		return err
	}
	initialized = true
	return nil
}

type Conf struct {
	Node    string // node 进程路径
	Proc    string // 当前执行任务路径
	Cmd     string // cmd 路径
	Once    string // 马上执行任务路径
	Lock    string // job lock 路径
	Group   string // 节点分组
	Noticer string // 通知

	Ttl        int64 // 节点超时时间，单位秒
	ReqTimeout int   // 请求超时时间，单位秒
	// 执行任务信息过期时间，单位秒
	// 0 为不过期
	ProcTtl int64
	// 记录任务执行中的信息的执行时间阀值，单位秒
	// 0 为不限制
	ProcReq int64
	// 单机任务锁过期时间，单位秒
	// 默认 300
	LockTtl int64

	PIDFile string

	Etcd client.Config
	Mgo  *db.Config
	Web  *webConfig
	Mail *MailConf

	Security *Security
}

type webConfig struct {
	BindAddr string
	Auth     struct {
		Enabled bool
	}
	Session    SessionConfig
	LogCleaner struct {
		EveryMinute    int
		ExpirationDays int
	}
}

type SessionConfig struct {
	Expiration      int
	CookieName      string
	StorePrefixPath string
}

type MailConf struct {
	Enable bool
	To     []string
	// 如果配置，则按 http api 方式发送，否则按 smtp 方式发送
	HttpAPI string
	// 如果此时间段内没有邮件发送，则关闭 SMTP 连接，单位/秒
	Keepalive int64
	*gomail.Dialer
}

type Security struct {
	// 是不开启安全选项
	// true 开启
	// 所执行的命令只能是机器上的脚本，仅支持配置项里的扩展名
	// 执行用户只能选择配置里的用户
	// false 关闭，命令和用户可以用动填写
	Open bool `json:"open"`
	// 配置执行用户
	Users []string `json:"users"`
	// 支持的执行的脚本扩展名
	Ext []string `json:"ext"`
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

func (c *Conf) parse() error {
	err := utils.LoadExtendConf(*confFile, c)
	if err != nil {
		return err
	}

	if c.Etcd.DialTimeout > 0 {
		c.Etcd.DialTimeout *= time.Second
	}
	if c.Ttl <= 0 {
		c.Ttl = 10
	}
	if c.LockTtl < 2 {
		c.LockTtl = 300
	}
	if c.Mail.Keepalive <= 0 {
		c.Mail.Keepalive = 30
	}
	if c.Mgo.Timeout <= 0 {
		c.Mgo.Timeout = 10 * time.Second
	} else {
		c.Mgo.Timeout *= time.Second
	}

	if c.Web != nil {
		if c.Web.LogCleaner.EveryMinute < 0 {
			c.Web.LogCleaner.EveryMinute = 30
		}
		if c.Web.LogCleaner.ExpirationDays <= 0 {
			c.Web.LogCleaner.ExpirationDays = 1
		}
	}

	c.Node = cleanKeyPrefix(c.Node)
	c.Proc = cleanKeyPrefix(c.Proc)
	c.Cmd = cleanKeyPrefix(c.Cmd)
	c.Once = cleanKeyPrefix(c.Once)
	c.Lock = cleanKeyPrefix(c.Lock)
	c.Group = cleanKeyPrefix(c.Group)
	c.Noticer = cleanKeyPrefix(c.Noticer)

	return nil
}

func (c *Conf) watch() error {
	var err error
	watcher, err = fsnotify.NewWatcher()
	if err != nil {
		return err
	}

	go func() {
		duration := 3 * time.Second
		timer, update := time.NewTimer(duration), false
		for {
			select {
			case <-exitChan:
				return
			case event := <-watcher.Events:
				// 保存文件时会产生多个事件
				if event.Op&(fsnotify.Write|fsnotify.Chmod) > 0 {
					update = true
				}
				timer.Reset(duration)
			case <-timer.C:
				if update {
					c.reload()
					event.Emit(event.WAIT, nil)
					update = false
				}
				timer.Reset(duration)
			case err := <-watcher.Errors:
				log.Warnf("config watcher err: %v", err)
			}
		}
	}()

	return watcher.Add(*confFile)
}

// 重新加载配置项
// 注：与系统资源相关的选项不生效，需重启程序
// Etcd
// Mgo
// Web
func (c *Conf) reload() {
	cf := new(Conf)
	if err := cf.parse(); err != nil {
		log.Warnf("config file reload err: %s", err.Error())
		return
	}

	// etcd key 选项需要重启
	cf.Node, cf.Proc, cf.Cmd, cf.Once, cf.Lock, cf.Group, cf.Noticer = c.Node, c.Proc, c.Cmd, c.Once, c.Lock, c.Group, c.Noticer

	*c = *cf
	log.Infof("config file[%s] reload success", *confFile)
	return
}

func Exit(i interface{}) {
	close(exitChan)
	if watcher != nil {
		watcher.Close()
	}
}
