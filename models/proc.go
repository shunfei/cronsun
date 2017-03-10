package models

import (
	"context"
	"fmt"
	"sync"
	"time"

	client "github.com/coreos/etcd/clientv3"

	"strings"
	"sunteng/commons/log"
	"sunteng/cronsun/conf"
)

var (
	lID *leaseID
)

// 维持 lease id 服务
func StartProc() error {
	lID = &leaseID{
		ttl:  conf.Config.ProcTtl,
		lk:   new(sync.RWMutex),
		done: make(chan struct{}),
	}

	if lID.ttl == 0 {
		return nil
	}

	err := lID.set()
	go lID.keepAlive()
	return err
}

func Reload(i interface{}) {
	if lID.ttl == conf.Config.ProcTtl {
		return
	}

	close(lID.done)
	lID.done, lID.ttl = make(chan struct{}), conf.Config.ProcTtl
	if conf.Config.ProcTtl == 0 {
		return
	}

	if err := lID.set(); err != nil {
		log.Warnf("proc lease id set err: %s", err.Error())
	}
	go lID.keepAlive()
}

func Exit(i interface{}) {
	if lID.done != nil {
		close(lID.done)
	}
}

type leaseID struct {
	ttl int64
	ID  client.LeaseID
	lk  *sync.RWMutex

	done chan struct{}
}

func (l *leaseID) get() client.LeaseID {
	if l.ttl == 0 {
		return -1
	}

	l.lk.RLock()
	id := l.ID
	l.lk.RUnlock()
	return id
}

func (l *leaseID) set() error {
	id := client.LeaseID(-1)
	resp, err := DefalutClient.Grant(context.TODO(), l.ttl+2)
	if err == nil {
		id = resp.ID
	}

	l.lk.Lock()
	l.ID = id
	l.lk.Unlock()
	return err
}

func (l *leaseID) keepAlive() {
	duration := time.Duration(l.ttl) * time.Second
	timer := time.NewTimer(duration)
	for {
		select {
		case <-l.done:
			return
		case <-timer.C:
			if l.ttl == 0 {
				return
			}

			id := l.get()
			if id < 0 {
				if err := l.set(); err != nil {
					log.Warnf("proc lease id set err: %s, try to reset after %d seconds...", err.Error(), l.ttl)
				}
				timer.Reset(duration)
				continue
			}

			_, err := DefalutClient.KeepAliveOnce(context.TODO(), l.ID)
			if err == nil {
				timer.Reset(duration)
				continue
			}

			log.Warnf("proc lease id keepAlive err: %s, try to reset...", err.Error())
			if err = l.set(); err != nil {
				log.Warnf("proc lease id set err: %s, try to reset after %d seconds...", err.Error(), l.ttl)
			}
			timer.Reset(duration)
		}
	}
}

// 当前执行中的任务信息
// key: /cronsun/proc/node/group/jobId/pid
// value: 开始执行时间
// key 会自动过期，防止进程意外退出后没有清除相关 key，过期时间可配置
type Process struct {
	ID     string    `json:"id"` // pid
	JobID  string    `json:"jobId"`
	Group  string    `json:"group"`
	NodeID string    `json:"nodeId"`
	Time   time.Time `json:"time"` // 开始执行时间

	running bool
}

func GetProcFromKey(key string) (proc *Process, err error) {
	ss := strings.Split(key, "/")
	var sslen = len(ss)
	if sslen < 5 {
		err = fmt.Errorf("invalid proc key [%s]", err.Error())
		return
	}

	proc = &Process{
		ID:     ss[sslen-1],
		JobID:  ss[sslen-2],
		Group:  ss[sslen-3],
		NodeID: ss[sslen-4],
	}
	return
}

func (p *Process) Key() string {
	return conf.Config.Proc + p.NodeID + "/" + p.Group + "/" + p.JobID + "/" + p.ID
}

func (p *Process) Val() string {
	return p.Time.Format(time.RFC3339)
}

// 获取结点正在执行任务的数量
func (j *Job) CountRunning() (int64, error) {
	resp, err := DefalutClient.Get(conf.Config.Proc+j.runOn+"/"+j.Group+"/"+j.ID, client.WithPrefix(), client.WithCountOnly())
	if err != nil {
		return 0, err
	}

	return resp.Count, nil
}

func (p *Process) put() error {
	id := lID.get()
	if id < 0 {
		_, err := DefalutClient.Put(p.Key(), p.Val())
		return err
	}

	_, err := DefalutClient.Put(p.Key(), p.Val(), client.WithLease(id))
	return err
}

func (p *Process) del() error {
	_, err := DefalutClient.Delete(p.Key())
	return err
}

func (p *Process) Start() {
	if err := p.put(); err != nil {
		log.Warnf("proc put err: %s", err.Error())
		return
	}

	p.running = true
}

func (p *Process) Stop() error {
	if !p.running {
		return nil
	}

	return p.del()
}
