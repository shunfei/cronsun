package models

import (
	"encoding/json"
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
	Pause   bool       `json:"pause"` // 可手工控制的状态

	// map[ip]timer node 服务使用
	// 每个任务在单个结点上只支持一个时间规则
	// 如果需要多个时间规则，需建新的任务
	Schedules map[string]string `json:"-"`
}

type JobRule struct {
	Timer          string   `json:"timer"`
	GroupIDs       []string `json:"gids"`
	NodeIDs        []string `json:"nids"`
	ExcludeNodeIDs []string `json:"exclude_nids"`
}

func GetJob(group, id string) (job *Job, err error) {
	resp, err := DefalutClient.Get(JobKey(group, id))
	if err != nil {
		return
	}

	if resp.Count == 0 {
		err = ErrNotFound
		return
	}

	err = json.Unmarshal(resp.Kvs[0].Value, &job)
	return
}

func DeleteJob(group, id string) (resp *client.DeleteResponse, err error) {
	return DefalutClient.Delete(JobKey(group, id))
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
	j.Schedules = make(map[string]string)
	for _, r := range j.Rule {
		for _, gid := range r.GroupIDs {
			g, ok := gs[gid]
			if !ok {
				continue
			}
			for _, id := range g.NodeIDs {
				if t, ok := j.Schedules[id]; ok {
					log.Warnf("job[%s] already exists timer[%s], timer[%s] will skip", j.ID, t, r.Timer)
					continue
				}
				j.Schedules[id] = r.Timer
			}
		}

		for _, id := range r.NodeIDs {
			if t, ok := j.Schedules[id]; ok {
				log.Warnf("job[%s] already exists timer[%s], timer[%s] will skip", j.ID, t, r.Timer)
				continue
			}
			j.Schedules[id] = r.Timer
		}

		for _, id := range r.ExcludeNodeIDs {
			delete(j.Schedules, id)
		}
	}
}

func (j *Job) Schedule(id string) (string, bool) {
	if len(j.Schedules) == 0 {
		return "", false
	}

	s, ok := j.Schedules[id]
	return s, ok
}

func (j *Job) GetID() string {
	return j.ID
}

func (j *Job) Run() {
}

func JobKey(group, id string) string {
	return conf.Config.Cmd + group + "/" + id
}

func (j *Job) Key() string {
	return JobKey(j.Group, j.ID)
}

func (j *Job) Check() error {
	j.ID = strings.TrimSpace(j.ID)
	if !IsValidAsKeyPath(j.ID) {
		return ErrIllegalJobId
	}

	j.Name = strings.TrimSpace(j.Name)
	if len(j.Name) == 0 {
		return ErrEmptyJobName
	}

	j.Group = strings.TrimSpace(j.Group)
	if len(j.Group) == 0 {
		j.Group = DefaultJobGroup
	}

	if !IsValidAsKeyPath(j.Group) {
		return ErrIllegalJobGroupName
	}

	// 不修改 Command 的内容，简单判断是否为空
	if len(strings.TrimSpace(j.Command)) == 0 {
		return ErrEmptyJobCommand
	}

	return nil
}
