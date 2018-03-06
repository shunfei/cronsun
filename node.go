package cronsun

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"syscall"
	"time"

	client "github.com/coreos/etcd/clientv3"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"github.com/shunfei/cronsun/conf"
	"github.com/shunfei/cronsun/log"
)

const (
	Coll_Node = "node"
)

// 执行 cron cmd 的进程
// 注册到 /cronsun/node/<id>
type Node struct {
	ID       string `bson:"_id" json:"id"`  // machine id
	PID      string `bson:"pid" json:"pid"` // 进程 pid
	IP       string `bson:"ip" json:"ip"`   // node ip
	Hostname string `bson:"hostname" json:"hostname"`

	Version  string    `bson:"version" json:"version"`
	UpTime   time.Time `bson:"up" json:"up"`     // 启动时间
	DownTime time.Time `bson:"down" json:"down"` // 上次关闭时间

	Alived    bool `bson:"alived" json:"alived"` // 是否可用
	Connected bool `bson:"-" json:"connected"`   // 当 Alived 为 true 时有效，表示心跳是否正常
}

func (n *Node) String() string {
	return "node[" + n.ID + "] pid[" + n.PID + "]"
}

func (n *Node) Put(opts ...client.OpOption) (*client.PutResponse, error) {
	return DefalutClient.Put(conf.Config.Node+n.ID, n.PID, opts...)
}

func (n *Node) Del() (*client.DeleteResponse, error) {
	return DefalutClient.Delete(conf.Config.Node + n.ID)
}

// 判断 node 是否已注册到 etcd
// 存在则返回进行 pid，不存在返回 -1
func (n *Node) Exist() (pid int, err error) {
	resp, err := DefalutClient.Get(conf.Config.Node + n.ID)
	if err != nil {
		return
	}

	if len(resp.Kvs) == 0 {
		return -1, nil
	}

	if pid, err = strconv.Atoi(string(resp.Kvs[0].Value)); err != nil {
		if _, err = DefalutClient.Delete(conf.Config.Node + n.ID); err != nil {
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
	return GetNodesBy(nil)
}

func GetNodesBy(query interface{}) (nodes []*Node, err error) {
	err = mgoDB.WithC(Coll_Node, func(c *mgo.Collection) error {
		return c.Find(query).All(&nodes)
	})

	return
}

func RemoveNode(query interface{}) error {
	return mgoDB.WithC(Coll_Node, func(c *mgo.Collection) error {
		return c.Remove(query)
	})

}

func ISNodeAlive(id string) (bool, error) {
	n := 0
	err := mgoDB.WithC(Coll_Node, func(c *mgo.Collection) error {
		var e error
		n, e = c.Find(bson.M{"_id": id, "alived": true}).Count()
		return e
	})

	return n > 0, err
}

func GetNodeGroups() (list []*Group, err error) {
	resp, err := DefalutClient.Get(conf.Config.Group, client.WithPrefix(), client.WithSort(client.SortByKey, client.SortAscend))
	if err != nil {
		return
	}

	list = make([]*Group, 0, resp.Count)
	for i := range resp.Kvs {
		g := Group{}
		err = json.Unmarshal(resp.Kvs[i].Value, &g)
		if err != nil {
			err = fmt.Errorf("node.GetGroups(key: %s) error: %s", string(resp.Kvs[i].Key), err.Error())
			return
		}
		list = append(list, &g)
	}

	return
}

func WatchNode() client.WatchChan {
	return DefalutClient.Watch(conf.Config.Node, client.WithPrefix())
}

// On 结点实例启动后，在 mongoDB 中记录存活信息
func (n *Node) On() {
	// remove old version(< 0.3.0) node info
	mgoDB.RemoveId(Coll_Node, n.IP)

	n.Alived, n.Version, n.UpTime = true, Version, time.Now()
	if err := mgoDB.Upsert(Coll_Node, bson.M{"_id": n.ID}, n); err != nil {
		log.Errorf(err.Error())
	}
}

// On 结点实例停用后，在 mongoDB 中去掉存活信息
func (n *Node) Down() {
	n.Alived, n.DownTime = false, time.Now()
	if err := mgoDB.Upsert(Coll_Node, bson.M{"_id": n.ID}, n); err != nil {
		log.Errorf(err.Error())
	}
}
