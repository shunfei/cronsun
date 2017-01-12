package models

import (
	"encoding/json"
	"errors"
	"strings"

	client "github.com/coreos/etcd/clientv3"

	"sunteng/commons/log"
	"sunteng/cronsun/conf"
)

const (
	DefaultJobGroup = "Default"
)

// 需要执行的 cron cmd 命令
// 注册到 /cronsun/cmd/groupName/<id>
type Job struct {
	ID      string     `json:"id"`
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

var (
	ErrEmptyJobName    = errors.New("Name of job is empty.")
	ErrEmptyJobCommand = errors.New("Command of job is empty.")
)

func (j *Job) Key() string {
	return conf.Config.Cmd + j.Group + "/" + j.ID
}

func (j *Job) Check() error {
	j.Name = strings.TrimSpace(j.Name)
	if len(j.Name) == 0 {
		return ErrEmptyJobName
	}

	j.Group = strings.TrimSpace(j.Group)
	if len(j.Group) == 0 {
		j.Group = DefaultJobGroup
	}

	// 不修改 Command 的内容，简单判断是否为空
	if len(strings.TrimSpace(j.Command)) == 0 {
		return ErrEmptyJobCommand
	}

	return nil
}
