package node

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"golang.org/x/net/context"

	client "github.com/coreos/etcd/clientv3"

	"sunteng/commons/util"
	"sunteng/cronsun/conf"
)

const (
	ReqTimeout = 2 * time.Second

	Spliter = "/"
)

// Node 执行 cron 命令服务的结构体
type Node struct {
	*client.Client

	Key string
	PID string

	lID client.LeaseID // lease id
}

func NewNode(cfg client.Config) (n *Node, err error) {
	ip, err := util.GetLocalIP()
	if err != nil {
		return
	}

	cli, err := client.New(cfg)
	if err != nil {
		return
	}

	n = &Node{
		Client: cli,

		Key: conf.Config.Proc + Spliter + ip.String(),
		PID: strconv.Itoa(os.Getpid()),
	}
	return
}

// 注册到 /cronsun/proc/xx
func (n *Node) Register() (err error) {
	pid, err := n.Exist()
	if err != nil {
		return
	}

	if pid != -1 {
		return fmt.Errorf("node[%s] pid[%d] exist", n.Key, pid)
	}

	resp, err := n.Client.Grant(context.TODO(), conf.Config.Ttl)
	if err != nil {
		return
	}

	if _, err = n.Client.Put(context.TODO(), n.Key, n.PID, client.WithLease(resp.ID)); err != nil {
		return
	}
	if _, err = n.Client.KeepAlive(context.TODO(), resp.ID); err != nil {
		return
	}
	n.lID = resp.ID

	return
}

// 判断 node 是否已注册到 etcd
// 存在则返回进行 pid，不存在返回 -1
func (n *Node) Exist() (pid int, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), ReqTimeout)
	resp, err := n.Client.Get(ctx, n.Key, client.WithFromKey())
	defer cancel()
	if err != nil {
		return
	}

	if len(resp.Kvs) == 0 {
		return -1, nil
	}

	if pid, err = strconv.Atoi(string(resp.Kvs[0].Value)); err != nil {
		if _, err = n.Client.Delete(ctx, n.Key, client.WithFromKey()); err != nil {
			return
		}
	}

	p, err := os.FindProcess(pid)
	if err != nil {
		return
	}

	if p != nil {
		return
	}

	return -1, nil
}

// 启动服务
func (n *Node) Run() {

}

// 启动服务
func (n *Node) Stop(i interface{}) {
	n.Client.Revoke(context.TODO(), n.lID)
	n.Client.Close()
}
