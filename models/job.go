package models

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	client "github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/mvcc/mvccpb"

	"sunteng/commons/log"
	"sunteng/cronsun/conf"
	"time"
)

const (
	DefaultJobGroup = "default"
)

// 需要执行的 cron cmd 命令
// 注册到 /cronsun/cmd/groupName/<id>
type Job struct {
	ID      string     `json:"id"`
	Name    string     `json:"name"`
	Group   string     `json:"group"`
	Command string     `json:"cmd"`
	User    string     `json:"user"`
	Rules   []*JobRule `json:"rules"`
	Pause   bool       `json:"pause"` // 可手工控制的状态

	// node 服务使用
	// 每个任务在单个结点上只支持一个时间规则
	// 如果需要多个时间规则，需建新的任务
	schedule string
	gid      string
	build    bool

	// 执行任务的结点，用于记录 job log
	runOn string
}

type JobRule struct {
	Timer          string   `json:"timer"`
	GroupIDs       []string `json:"gids"`
	NodeIDs        []string `json:"nids"`
	ExcludeNodeIDs []string `json:"exclude_nids"`
}

func (j *JobRule) included(nid string, gs map[string]*Group) (string, bool) {
	for _, gid := range j.GroupIDs {
		if _, ok := gs[gid]; ok {
			return gid, true
		}
	}

	for i, count := 0, len(j.NodeIDs); i < count; i++ {
		if nid == j.NodeIDs[i] {
			return "", true
		}
	}

	return "", false
}

func GetJob(group, id string) (job *Job, err error) {
	job, _, err = GetJobAndRev(group, id)
	return
}

func GetJobAndRev(group, id string) (job *Job, rev int64, err error) {
	resp, err := DefalutClient.Get(JobKey(group, id))
	if err != nil {
		return
	}

	if resp.Count == 0 {
		err = ErrNotFound
		return
	}

	rev = resp.Kvs[0].ModRevision
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

func WatchJobs() client.WatchChan {
	return DefalutClient.Watch(conf.Config.Cmd, client.WithPrefix(), client.WithPrevKV())
}

func GetJobFromKv(kv *mvccpb.KeyValue) (job *Job, err error) {
	job = new(Job)
	if err = json.Unmarshal(kv.Value, job); err != nil {
		err = fmt.Errorf("job[%s] umarshal err: %s", string(kv.Key), err.Error())
	}
	return
}

// Schedule return schedule and group id
func (j *Job) Schedule(nid string, gs map[string]*Group, rebuild bool) (sch string, gid string) {
	if j.Pause {
		return
	}

	if j.build && !rebuild {
		return j.schedule, j.gid
	}

	j.buildSchedule(nid, gs)
	return j.schedule, j.gid
}

func (j *Job) buildSchedule(nid string, gs map[string]*Group) {
	j.build = true
	for _, r := range j.Rules {
		for _, id := range r.ExcludeNodeIDs {
			if nid == id {
				return
			}
		}

		if gid, ok := r.included(nid, gs); ok {
			j.schedule, j.gid = r.Timer, gid
			return
		}
	}
}

func (j *Job) GetID() string {
	return j.ID
}

func (j *Job) RunOn(n string) {
	j.runOn = n
}

func (j *Job) String() string {
	data, err := json.Marshal(j)
	if err != nil {
		return err.Error()
	}
	return string(data)
}

// Run 执行任务
func (j *Job) Run() {
	t := time.Now()
	var cmd *exec.Cmd
	if len(j.User) > 0 {
		if needPassword {
			j.Fail(t, SudoErr)
			return
		}
		cmd = exec.Command("sudo", "su", j.User, "-c", j.Command)
	} else {
		args := strings.Split(j.Command, " ")
		cmd = exec.Command(args[0], args[1:]...)
	}

	out, err := cmd.Output()
	if err != nil {
		j.Fail(t, err)
		return
	}

	j.Success(t, out)
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

// 执行结果写入 mongoDB
func (j *Job) Success(t time.Time, out []byte) {
	CreateJobLog(j, t, string(out), true)
}

func (j *Job) Fail(t time.Time, err error) {
	CreateJobLog(j, t, err.Error(), false)
}
