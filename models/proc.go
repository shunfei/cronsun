package models

import (
	"context"
	"time"

	client "github.com/coreos/etcd/clientv3"

	"sunteng/cronsun/conf"
)

// 当前执行中的任务信息
// key: /cronsun/proc/node/job id/pid
// value: 开始执行时间
// key 会自动过期，防止进程意外退出后没有清除相关 key，过期时间可配置
type Process struct {
	ID     string    `json:"id"`
	JobID  string    `json:"job_id"`
	NodeID string    `json:"node_id"`
	Time   time.Time `json:"name"` // 开始执行时间
}

func (p *Process) Key() string {
	return conf.Config.Proc + p.NodeID + "/" + p.JobID + "/" + p.ID
}

func (p *Process) Val() string {
	return p.Time.Format(time.RFC3339)
}

// 获取结点正在执行任务的数量
func (p *Process) Count() (int64, error) {
	resp, err := DefalutClient.Get(conf.Config.Proc + p.NodeID + "/" + p.JobID + "/")
	if err != nil {
		return 0, err
	}

	return resp.Count, nil
}

func (p *Process) Put() error {
	if conf.Config.ProcTtl == 0 {
		_, err := DefalutClient.Put(p.Key(), p.Val())
		return err
	}

	resp, err := DefalutClient.Grant(context.TODO(), conf.Config.ProcTtl)
	if err != nil {
		return err
	}

	_, err = DefalutClient.Put(p.Key(), p.Val(), client.WithLease(resp.ID))
	return err
}
