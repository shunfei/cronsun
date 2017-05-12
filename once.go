package cronsun

import (
	client "github.com/coreos/etcd/clientv3"

	"github.com/shunfei/cronsun/conf"
)

// 马上执行 job 任务
// 注册到 /cronsun/once/group/<jobID>
// value
// 若执行单个结点，则值为 NodeID
// 若 job 所在的结点都需执行，则值为空 ""
func PutOnce(group, jobID, nodeID string) error {
	_, err := DefalutClient.Put(conf.Config.Once+group+"/"+jobID, nodeID)
	return err
}

func WatchOnce() client.WatchChan {
	return DefalutClient.Watch(conf.Config.Once, client.WithPrefix())
}
