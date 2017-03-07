package models

import (
	"context"
	"time"

	client "github.com/coreos/etcd/clientv3"

	"sunteng/commons/log"
	"sunteng/cronsun/conf"
)

var (
	leaseID client.LeaseID
)

func ProcessKeepAlive() error {
	if conf.Config.ProcTtl == 0 {
		return nil
	}

	resp, err := DefalutClient.Grant(context.TODO(), conf.Config.ProcTtl+5)
	if err != nil {
		return err
	}

	leaseID = resp.ID
	return nil
}

// 当前执行中的任务信息
// key: /cronsun/proc/node/job id/pid
// value: 开始执行时间
// key 会自动过期，防止进程意外退出后没有清除相关 key，过期时间可配置
type Process struct {
	ID     string    `json:"id"`
	JobID  string    `json:"job_id"`
	NodeID string    `json:"node_id"`
	Time   time.Time `json:"name"` // 开始执行时间

	running bool
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
	if leaseID == 0 {
		_, err := DefalutClient.Put(p.Key(), p.Val())
		return err
	}

	_, err := DefalutClient.Put(p.Key(), p.Val(), client.WithLease(leaseID))
	return err
}

func (p *Process) Del() error {
	_, err := DefalutClient.Delete(p.Key())
	return err
}

func (p *Process) Start() {
	if err := p.Put(); err != nil {
		log.Warnf("proc put err: %s", err.Error())
		return
	}

	p.running = true
}

func (p *Process) Stop() error {
	if !p.running {
		return nil
	}

	return p.Del()
}
