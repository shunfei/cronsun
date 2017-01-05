package node

import (
	"os"

	client "github.com/coreos/etcd/clientv3"

	"sunteng/commons/util"
)

// Node 执行 cron 命令服务的结构体
type Node struct {
	*client.Client

	IP  string
	PID int
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

		IP:  ip.String(),
		PID: os.Getpid(),
	}
	return
}

// 注册到 /cronsun/proc/xx
func (n *Node) Register() error {
	return nil
}

// 更新 /cronsun/proc/xx/time
// 用于检查 node 是否存活
func (n *Node) Heartbeat() {

}

// 启动服务
func (n *Node) Run() {

}

// 启动服务
func (n *Node) Stop(i interface{}) {
	n.Client.Close()
}
