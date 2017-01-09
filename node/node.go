package node

import (
	"fmt"
	"os"
	"strconv"
	"syscall"
	"time"

	"golang.org/x/net/context"

	client "github.com/coreos/etcd/clientv3"

	"sunteng/commons/log"
	"sunteng/commons/util"
	"sunteng/cronsun/conf"
)

// Node 执行 cron 命令服务的结构体
type Node struct {
	*client.Client

	ttl        int64
	reqTimeout time.Duration
	prefix     string

	Key string
	PID string

	lID client.LeaseID // lease id
	lch <-chan *client.LeaseKeepAliveResponse

	done chan struct{}
}

func NewNode(cfg *conf.Conf) (n *Node, err error) {
	ip, err := util.GetLocalIP()
	if err != nil {
		return
	}

	cli, err := client.New(cfg.Etcd)
	if err != nil {
		return
	}

	n = &Node{
		Client: cli,

		ttl:        cfg.Ttl,
		reqTimeout: time.Duration(cfg.ReqTimeout) * time.Second,
		prefix:     cfg.Proc,

		Key: cfg.Proc + ip.String(),
		PID: strconv.Itoa(os.Getpid()),

		done: make(chan struct{}),
	}
	return
}

func (n *Node) String() string {
	return "node[" + n.Key[len(n.prefix):] + "] pid[" + n.PID + "]"
}

// 注册到 /cronsun/proc/xx
func (n *Node) Register() (err error) {
	pid, err := n.Exist()
	if err != nil {
		return
	}

	if pid != -1 {
		return fmt.Errorf("node[%s] pid[%d] exist", n.Key[len(n.prefix):], pid)
	}

	resp, err := n.Client.Grant(context.TODO(), n.ttl)
	if err != nil {
		return
	}

	if _, err = n.Client.Put(context.TODO(), n.Key, n.PID, client.WithLease(resp.ID)); err != nil {
		return
	}

	ch, err := n.Client.KeepAlive(context.TODO(), resp.ID)
	if err != nil {
		return
	}

	n.lID, n.lch = resp.ID, ch
	return
}

// 判断 node 是否已注册到 etcd
// 存在则返回进行 pid，不存在返回 -1
func (n *Node) Exist() (pid int, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), n.reqTimeout)
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
		return -1, nil
	}

	p, err := os.FindProcess(pid)
	if err != nil {
		return -1, nil
	}

	// TODO: 暂时不考虑 linux/unix 以外的系统
	if p != nil && p.Signal(syscall.Signal(0)) == nil {
		return
	}

	return -1, nil
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
	// 防止断网时卡住
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	n.Client.Delete(ctx, n.Key, client.WithFromKey())
	cancel()
	n.Client.Close()
}
