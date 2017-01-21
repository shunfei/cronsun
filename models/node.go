package models

import (
	"os"
	"strconv"
	"syscall"

	client "github.com/coreos/etcd/clientv3"

	"sunteng/cronsun/conf"
)

// 执行 cron cmd 的进程
// 注册到 /cronsun/proc/<id>
type Node struct {
	ID  string `json:"-"`   // ip
	PID string `json:"pid"` // 进程 pid
}

func (n *Node) String() string {
	return "node[" + n.ID + "] pid[" + n.PID + "]"
}

func (n *Node) Put(opts ...client.OpOption) (*client.PutResponse, error) {
	return DefalutClient.Put(conf.Config.Proc+n.ID, n.PID, opts...)
}

func (n *Node) Del() (*client.DeleteResponse, error) {
	return DefalutClient.Delete(conf.Config.Proc+n.ID, client.WithFromKey())
}

// 判断 node 是否已注册到 etcd
// 存在则返回进行 pid，不存在返回 -1
func (n *Node) Exist() (pid int, err error) {
	resp, err := DefalutClient.Get(conf.Config.Proc+n.ID, client.WithFromKey())
	if err != nil {
		return
	}

	if len(resp.Kvs) == 0 {
		return -1, nil
	}

	if pid, err = strconv.Atoi(string(resp.Kvs[0].Value)); err != nil {
		if _, err = DefalutClient.Delete(conf.Config.Proc+n.ID, client.WithFromKey()); err != nil {
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

func GetActivityNodeList() (nodes []string, err error) {
	resp, err := DefalutClient.Get(conf.Config.Proc, client.WithPrefix(), client.WithKeysOnly())
	if err != nil {
		return
	}

	procKeyLen := len(conf.Config.Proc)
	for _, n := range resp.Kvs {
		nodes = append(nodes, string(n.Key[procKeyLen:]))
	}

	return
}
