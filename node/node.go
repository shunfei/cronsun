package node

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"golang.org/x/net/context"

	client "github.com/coreos/etcd/clientv3"

	"sunteng/commons/log"
	"sunteng/commons/util"
	"sunteng/cronsun/conf"
	"sunteng/cronsun/models"
)

// Node 执行 cron 命令服务的结构体
type Node struct {
	*models.Client
	*models.Node

	ttl int64

	lID client.LeaseID // lease id
	lch <-chan *client.LeaseKeepAliveResponse

	done chan struct{}
}

func NewNode(cfg *conf.Conf) (n *Node, err error) {
	ip, err := util.GetLocalIP()
	if err != nil {
		return
	}

	n = &Node{
		Client: models.DefalutClient,
		Node: &models.Node{
			ID:  ip.String(),
			PID: strconv.Itoa(os.Getpid()),
		},

		ttl:  cfg.Ttl,
		done: make(chan struct{}),
	}
	return
}

// 注册到 /cronsun/proc/xx
func (n *Node) Register() (err error) {
	pid, err := n.Node.Exist()
	if err != nil {
		return
	}

	if pid != -1 {
		return fmt.Errorf("node[%s] pid[%d] exist", n.Node.ID, pid)
	}

	resp, err := n.Client.Grant(context.TODO(), n.ttl)
	if err != nil {
		return
	}

	if _, err = n.Node.Put(client.WithLease(resp.ID)); err != nil {
		return
	}

	ch, err := n.Client.KeepAlive(context.TODO(), resp.ID)
	if err != nil {
		return
	}

	n.lID, n.lch = resp.ID, ch
	return
}

// 启动服务
func (n *Node) Run() {
	go n.keepAlive()
}

// 断网掉线重新注册
func (n *Node) keepAlive() {
	for {
		for _ = range n.lch {
		}

		select {
		case <-n.done:
			return
		default:
		}
		time.Sleep(time.Duration(n.ttl+1) * time.Second)

		log.Noticef("%s has dropped, try to reconnect...", n.String())
		if err := n.Register(); err != nil {
			log.Warn(err.Error())
		} else {
			log.Noticef("%s reconnected", n.String())
		}
	}
}

// 启动服务
func (n *Node) Stop(i interface{}) {
	close(n.done)
	n.Node.Del()
	n.Client.Close()
}
