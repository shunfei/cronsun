package models

import (
	"encoding/json"

	client "github.com/coreos/etcd/clientv3"

	"sunteng/commons/log"
	"sunteng/cronsun/conf"
)

// 需要执行的 cron cmd 命令
// 注册到 /cronsun/cmd/<id>
type Job struct {
	ID      string     `json:"-"`
	Name    string     `json:"name"`
	Group   string     `json:"group"`
	Command string     `json:"cmd"`
	Rule    []*JobRule `json:"rule"`
	Pause   bool       `json:"Pause"` // 可手工控制的状态，运行中/暂停

	Schedules map[string][]string `json:"-"` // map[ip][]timer node 服务使用
}

type JobRule struct {
	Timer          string   `json:"timer"`
	GroupIDs       []string `json:"gids"`
	NodeIDs        []string `json:"nids"`
	ExcludeNodeIDs []string `json:"exclude_nids"`
}

func GetJobs() (jobs map[string]*Job, err error) {
	resp, err := DefalutClient.Get(conf.Config.Cmd, client.WithPrefix())
	if err != nil {
		return
	}

	count := len(resp.Kvs)
	if count == 0 {
		return
	}

	jobs = make(map[string]*Job, count)
	for _, j := range resp.Kvs {
		job := new(Job)
		if e := json.Unmarshal(j.Value, job); e != nil {
			log.Warnf("job[%s] umarshal err: %s", string(j.Key), e.Error())
			continue
		}
		jobs[job.ID] = job
	}
	return
}

func (j *Job) BuildSchedules(gs map[string]*Group) {
	j.Schedules = make(map[string][]string)
	for _, r := range j.Rule {
		sch := make(map[string]string)
		for _, gid := range r.GroupIDs {
			g, ok := gs[gid]
			if !ok {
				continue
			}
			for _, id := range g.NodeIDs {
				sch[id] = r.Timer
			}
		}

		for _, id := range r.NodeIDs {
			sch[id] = r.Timer
		}

		for _, id := range r.ExcludeNodeIDs {
			delete(sch, id)
		}

		for id, t := range sch {
			j.Schedules[id] = append(j.Schedules[id], t)
		}
	}
}

func (j *Job) Schedule(id string) ([]string, bool) {
	if len(j.Schedules) == 0 {
		return nil, false
	}

	s, ok := j.Schedules[id]
	return s, ok
}

func (j *Job) Run() {

}
