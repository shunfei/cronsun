package node

import (
	client "github.com/coreos/etcd/clientv3"
)

// Node 执行 cron 命令服务的结构体
type Node struct {
	*client.Client
}

func NewNode(cfg client.Config) *Node {
	return &Node{}
}

// 注册到 /cronsun/proc/xx
func (n *Node) Register() {

}

// 更新 /cronsun/proc/xx/time
// 用于检查 node 是否存活
func (n *Node) Heartbeat() {

}
