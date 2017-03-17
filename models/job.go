package models

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"os/user"
	"runtime"
	"strconv"
	"strings"
	"sync/atomic"
	"syscall"
	"time"

	"golang.org/x/net/context"

	client "github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/mvcc/mvccpb"

	"sunteng/commons/log"
	"sunteng/cronsun/conf"
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
	Pause   bool       `json:"pause"`   // 可手工控制的状态
	Timeout int64      `json:"timeout"` // 任务执行时间超时设置，大于 0 时有效
	// 设置任务在单个节点上可以同时允许多少个
	// 针对两次任务执行间隔比任务执行时间要长的任务启用
	Parallels int64 `json:"parallels"`
	// 执行任务失败重试次数
	// 默认为 0，不重试
	Retry int `json:"retry"`
	// 执行任务失败重试时间间隔
	// 单位秒，如果不大于 0 则马上重试
	Interval int `json:"interval"`
	// 平均执行时间，单位 ms
	AvgTime int64 `json:"avg_time"`

	// 执行任务的结点，用于记录 job log
	runOn string
	// 用于存储分隔后的任务
	cmd []string
	// 控制同时执行任务数
	count int64
}

type JobRule struct {
	ID             string   `json:"id"`
	Timer          string   `json:"timer"`
	GroupIDs       []string `json:"gids"`
	NodeIDs        []string `json:"nids"`
	ExcludeNodeIDs []string `json:"exclude_nids"`
}

type Cmd struct {
	*Job
	*JobRule
}

func (c *Cmd) GetID() string {
	return c.Job.ID + c.JobRule.ID
}

func (c *Cmd) Run() {
	// 同时执行任务数限制
	if c.Job.limit() {
		return
	}
	defer c.Job.unlimit()

	if c.Job.Retry <= 0 {
		c.Job.Run()
		return
	}

	for i := 0; i < c.Job.Retry; i++ {
		if c.Job.Run() {
			return
		}

		if c.Job.Interval > 0 {
			time.Sleep(time.Duration(c.Job.Interval) * time.Second)
		}
	}
}

// 优先取结点里的值，更新 group 时可用 gid 判断是否对 job 进行处理
func (j *JobRule) included(nid string, gs map[string]*Group) bool {
	for i, count := 0, len(j.NodeIDs); i < count; i++ {
		if nid == j.NodeIDs[i] {
			return true
		}
	}

	for _, gid := range j.GroupIDs {
		if g, ok := gs[gid]; ok && g.Included(nid) {
			return true
		}
	}

	return false
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
	if err = json.Unmarshal(resp.Kvs[0].Value, &job); err != nil {
		return
	}

	job.splitCmd()
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
	jobs = make(map[string]*Job, count)
	if count == 0 {
		return
	}

	for _, j := range resp.Kvs {
		job := new(Job)
		if e := json.Unmarshal(j.Value, job); e != nil {
			log.Warnf("job[%s] umarshal err: %s", string(j.Key), e.Error())
			continue
		}

		if err := job.Valid(); err != nil {
			log.Warnf("job[%s] is invalid: %s", string(j.Key), err.Error())
			continue
		}

		jobs[job.ID] = job
	}
	return
}

func WatchJobs() client.WatchChan {
	return DefalutClient.Watch(conf.Config.Cmd, client.WithPrefix())
}

func GetJobFromKv(kv *mvccpb.KeyValue) (job *Job, err error) {
	job = new(Job)
	if err = json.Unmarshal(kv.Value, job); err != nil {
		err = fmt.Errorf("job[%s] umarshal err: %s", string(kv.Key), err.Error())
		return
	}

	err = job.Valid()
	return
}

func (j *Job) RunOn(n string) {
	j.runOn = n
}

func (j *Job) splitCmd() {
	j.cmd = strings.Split(j.Command, " ")
}

func (j *Job) String() string {
	data, err := json.Marshal(j)
	if err != nil {
		return err.Error()
	}
	return string(data)
}

func (j *Job) limit() bool {
	if j.Parallels == 0 {
		return false
	}

	count := atomic.LoadInt64(&j.count)
	if j.Parallels <= count {
		j.Fail(time.Now(), fmt.Sprintf("job[%s] running on[%s] running:[%d]", j.Key(), j.runOn, count))
		return true
	}

	atomic.AddInt64(&j.count, 1)
	return false
}

func (j *Job) unlimit() {
	if j.Parallels == 0 {
		return
	}
	atomic.AddInt64(&j.count, -1)
}

// Run 执行任务
func (j *Job) Run() bool {
	var (
		cmd         *exec.Cmd
		proc        *Process
		sysProcAttr *syscall.SysProcAttr
	)

	t := time.Now()
	// 用户权限控制
	if len(j.User) > 0 {
		u, err := user.Lookup(j.User)
		if err != nil {
			j.Fail(t, err.Error())
			return false
		}

		uid, err := strconv.Atoi(u.Uid)
		if err != nil {
			j.Fail(t, "not support run with user on windows")
			return false
		}
		gid, _ := strconv.Atoi(u.Gid)
		sysProcAttr = &syscall.SysProcAttr{
			Credential: &syscall.Credential{
				Uid: uint32(uid),
				Gid: uint32(gid),
			},
		}
	}

	// 超时控制
	if j.Timeout > 0 {
		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(j.Timeout)*time.Second)
		defer cancel()
		cmd = exec.CommandContext(ctx, j.cmd[0], j.cmd[1:]...)
	} else {
		cmd = exec.Command(j.cmd[0], j.cmd[1:]...)
	}
	cmd.SysProcAttr = sysProcAttr
	var b bytes.Buffer
	cmd.Stdout = &b
	cmd.Stderr = &b
	if err := cmd.Start(); err != nil {
		j.Fail(t, fmt.Sprintf("%s", err.Error()))
		return false
	}

	proc = &Process{
		ID:     strconv.Itoa(cmd.Process.Pid),
		JobID:  j.ID,
		Group:  j.Group,
		NodeID: j.runOn,
		Time:   t,
	}
	proc.Start()
	defer proc.Stop()

	if err := cmd.Wait(); err != nil {
		j.Fail(t, fmt.Sprintf("%s", err.Error()))
		return false
	}

	j.Success(t, b.String())
	return true
}

func (j *Job) RunWithRecovery() {
	defer func() {
		if r := recover(); r != nil {
			const size = 64 << 10
			buf := make([]byte, size)
			buf = buf[:runtime.Stack(buf, false)]
			log.Warnf("panic running job: %v\n%s", r, buf)
		}
	}()
	j.Run()
}

// 从 etcd 的 key 中取 id
func GetIDFromKey(key string) string {
	index := strings.LastIndex(key, "/")
	if index < 0 {
		return ""
	}

	return key[index+1:]
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

	j.User = strings.TrimSpace(j.User)

	for i := range j.Rules {
		id := strings.TrimSpace(j.Rules[i].ID)
		if id == "" || strings.HasPrefix(id, "NEW") {
			j.Rules[i].ID = NextID()
		}
	}

	// 不修改 Command 的内容，简单判断是否为空
	if len(strings.TrimSpace(j.Command)) == 0 {
		return ErrEmptyJobCommand
	}

	return j.Valid()
}

// 执行结果写入 mongoDB
func (j *Job) Success(t time.Time, out string) {
	CreateJobLog(j, t, out, true)
}

func (j *Job) Fail(t time.Time, msg string) {
	CreateJobLog(j, t, msg, false)
}

func (j *Job) Avg(t, et time.Time) {
	execTime := int64(et.Sub(t) / time.Millisecond)
	if j.AvgTime == 0 {
		j.AvgTime = execTime
		return
	}

	j.AvgTime = (j.AvgTime + execTime) / 2
}

func (j *Job) Cmds(nid string, gs map[string]*Group) (cmds map[string]*Cmd) {
	cmds = make(map[string]*Cmd)
	if j.Pause {
		return
	}

	for _, r := range j.Rules {
		for _, id := range r.ExcludeNodeIDs {
			if nid == id {
				continue
			}
		}

		if r.included(nid, gs) {
			cmd := &Cmd{
				Job:     j,
				JobRule: r,
			}
			cmds[cmd.GetID()] = cmd
		}
	}

	return
}

func (j Job) IsRunOn(nid string, gs map[string]*Group) bool {
	for _, r := range j.Rules {
		for _, id := range r.ExcludeNodeIDs {
			if nid == id {
				continue
			}
		}

		if r.included(nid, gs) {
			return true
		}
	}

	return false
}

// 安全选项验证
func (j *Job) Valid() error {
	if len(j.cmd) == 0 {
		j.splitCmd()
	}

	security := conf.Config.Security
	if !security.Open {
		return nil
	}

	if !j.validUser() {
		return ErrSecurityInvalidUser
	}

	if !j.validCmd() {
		return ErrSecurityInvalidCmd
	}

	return nil
}

func (j *Job) validUser() bool {
	if len(conf.Config.Security.Users) == 0 {
		return true
	}

	for _, u := range conf.Config.Security.Users {
		if j.User == u {
			return true
		}
	}
	return false
}

func (j *Job) validCmd() bool {
	if len(conf.Config.Security.Ext) == 0 {
		return true
	}
	for _, ext := range conf.Config.Security.Ext {
		if strings.HasSuffix(j.cmd[0], ext) {
			return true
		}
	}
	return false
}
