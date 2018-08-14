package cronsun

import (
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	client "github.com/coreos/etcd/clientv3"

	"encoding/json"

	"github.com/shunfei/cronsun/conf"
	"github.com/shunfei/cronsun/log"
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
	resp, err := DefalutClient.Grant(l.ttl + 2)
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
			if id > 0 {
				_, err := DefalutClient.KeepAliveOnce(l.ID)
				if err == nil {
					timer.Reset(duration)
					continue
				}

				log.Warnf("proc lease id[%x] keepAlive err: %s, try to reset...", id, err.Error())
			}

			if err := l.set(); err != nil {
				log.Warnf("proc lease id set err: %s, try to reset after %d seconds...", err.Error(), l.ttl)
			} else {
				log.Infof("proc set lease id[%x] success", l.get())
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
	Time   time.Time `json:"time"`   // 开始执行时间
	Killed bool      `json:"killed"` // 是否强制杀死

	running int32
	hasPut  int32
	wg      sync.WaitGroup
	done    chan struct{}
}

func GetProcFromKey(key string) (proc *Process, err error) {
	ss := strings.Split(key, "/")
	var sslen = len(ss)
	if sslen < 5 {
		err = fmt.Errorf("invalid proc key [%s]", key)
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
	val := map[string]interface{}{
		"time":   p.Time.Format(time.RFC3339),
		"killed": p.Killed,
	}
	str, _ := json.Marshal(val)
	return string(str)
}

// 获取结点正在执行任务的数量
func (j *Job) CountRunning() (int64, error) {
	resp, err := DefalutClient.Get(conf.Config.Proc+j.runOn+"/"+j.Group+"/"+j.ID, client.WithPrefix(), client.WithCountOnly())
	if err != nil {
		return 0, err
	}

	return resp.Count, nil
}

// put 出错也进行 del 操作
// 有可能某种原因，put 命令已经发送到 etcd server
// 目前已知的 deadline 会出现此情况
func (p *Process) put() (err error) {
	if atomic.LoadInt32(&p.running) != 1 {
		return
	}

	if !atomic.CompareAndSwapInt32(&p.hasPut, 0, 1) {
		return
	}

	id := lID.get()
	if id < 0 {
		if _, err = DefalutClient.Put(p.Key(), p.Val()); err != nil {
			return
		}
	}

	_, err = DefalutClient.Put(p.Key(), p.Val(), client.WithLease(id))
	return
}

func (p *Process) del() error {
	if atomic.LoadInt32(&p.hasPut) != 1 {
		return nil
	}

	_, err := DefalutClient.Delete(p.Key())
	return err
}

func (p *Process) Start() {
	if p == nil {
		return
	}

	if !atomic.CompareAndSwapInt32(&p.running, 0, 1) {
		return
	}

	if conf.Config.ProcReq == 0 {
		if err := p.put(); err != nil {
			log.Warnf("proc put[%s] err: %s", p.Key(), err.Error())
		}
		return
	}

	p.done = make(chan struct{})
	p.wg.Add(1)
	go func() {
		select {
		case <-p.done:
		case <-time.After(time.Duration(conf.Config.ProcReq) * time.Second):
			if err := p.put(); err != nil {
				log.Warnf("proc put[%s] err: %s", p.Key(), err.Error())
			}
		}
		p.wg.Done()
	}()
}

func (p *Process) Stop() {
	if p == nil {
		return
	}

	if !atomic.CompareAndSwapInt32(&p.running, 1, 0) {
		return
	}

	if p.done != nil {
		close(p.done)
	}
	p.wg.Wait()

	if err := p.del(); err != nil {
		log.Warnf("proc del[%s] err: %s", p.Key(), err.Error())
	}
}

func WatchProcs(nid string) client.WatchChan {
	return DefalutClient.Watch(conf.Config.Proc+nid, client.WithPrefix())
}
