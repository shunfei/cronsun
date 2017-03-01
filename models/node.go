package models

import (
	"os"
	"strconv"
	"syscall"

	client "github.com/coreos/etcd/clientv3"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"sunteng/commons/log"
	"sunteng/cronsun/conf"
)

const (
	Coll_Node = "node"
)

// 执行 cron cmd 的进程
// 注册到 /cronsun/proc/<id>
type Node struct {
	ID  string `bson:"_id" json:"id"`  // ip
	PID string `bson:"pid" json:"pid"` // 进程 pid

	Alived    bool `bson:"alived" json:"alived"` // 是否可用
	Connected bool `bson:"-" json:"connected"`   // 当 Alived 为 true 时有效，表示心跳是否正常
}

func (n *Node) String() string {
	return "node[" + n.ID + "] pid[" + n.PID + "]"
}

func (n *Node) Put(opts ...client.OpOption) (*client.PutResponse, error) {
	return DefalutClient.Put(conf.Config.Proc+n.ID, n.PID, opts...)
}

func (n *Node) Del() (*client.DeleteResponse, error) {
	return DefalutClient.Delete(conf.Config.Proc + n.ID)
}

// 判断 node 是否已注册到 etcd
// 存在则返回进行 pid，不存在返回 -1
func (n *Node) Exist() (pid int, err error) {
	resp, err := DefalutClient.Get(conf.Config.Proc + n.ID)
	if err != nil {
		return
	}

	if len(resp.Kvs) == 0 {
		return -1, nil
	}

	if pid, err = strconv.Atoi(string(resp.Kvs[0].Value)); err != nil {
		if _, err = DefalutClient.Delete(conf.Config.Proc + n.ID); err != nil {
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

func GetNodes() (nodes []*Node, err error) {
	err = mgoDB.WithC(Coll_Node, func(c *mgo.Collection) error {
		return c.Find(nil).All(&nodes)
	})

	return
}

// On 结点实例启动后，在 mongoDB 中记录存活信息
func (n *Node) On() {
	n.Alived = true
	if err := mgoDB.Upsert(Coll_Node, bson.M{"_id": n.ID}, n); err != nil {
		log.Error(err.Error())
	}
}

// On 结点实例停用后，在 mongoDB 中去掉存活信息
func (n *Node) Down() {
	n.Alived = false
	if err := mgoDB.Upsert(Coll_Node, bson.M{"_id": n.ID}, n); err != nil {
		log.Error(err.Error())
	}
}
