package cronsun

import (
	"bytes"
	"context"
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

	client "github.com/coreos/etcd/clientv3"

	"github.com/shunfei/cronsun/conf"
	"github.com/shunfei/cronsun/log"
	"github.com/shunfei/cronsun/node/cron"
	"github.com/shunfei/cronsun/utils"
)

const (
	DefaultJobGroup = "default"
)

const (
	KindCommon   = iota
	KindAlone    // 任何时间段只允许单机执行
	KindInterval // 一个任务执行间隔内允许执行一次
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
	// 任务类型
	// 0: 普通任务
	// 1: 单机任务
	// 如果为单机任务，node 加载任务的时候 Parallels 设置 1
	Kind int `json:"kind"`
	// 平均执行时间，单位 ms
	AvgTime int64 `json:"avg_time"`
	// 执行失败发送通知
	FailNotify bool `json:"fail_notify"`
	// 发送通知地址
	To []string `json:"to"`
	// 单独对任务指定日志清除时间
	LogExpiration int `json:"log_expiration"`

	// 执行任务的结点，用于记录 job log
	runOn    string
	hostname string
	// 用于存储分隔后的任务
	cmd []string
	// 控制同时执行任务数
	Count *int64 `json:"-"`
}

type JobRule struct {
	ID             string   `json:"id"`
	Timer          string   `json:"timer"`
	GroupIDs       []string `json:"gids"`
	NodeIDs        []string `json:"nids"`
	ExcludeNodeIDs []string `json:"exclude_nids"`

	Schedule cron.Schedule `json:"-"`
}

// 任务锁
type locker struct {
	kind  int
	ttl   int64
	lID   client.LeaseID
	timer *time.Timer
	done  chan struct{}
}

func (l *locker) keepAlive() {
	duration := time.Duration(l.ttl)*time.Second - 500*time.Millisecond
	l.timer = time.NewTimer(duration)
	for {
		select {
		case <-l.done:
			return
		case <-l.timer.C:
			_, err := DefalutClient.KeepAliveOnce(l.lID)
			if err != nil {
				log.Warnf("lock keep alive err: %s", err.Error())
				return
			}
			l.timer.Reset(duration)
		}
	}
}

func (l *locker) unlock() {
	if l.kind != KindAlone {
		return
	}

	close(l.done)
	l.timer.Stop()
	if _, err := DefalutClient.Revoke(l.lID); err != nil {
		log.Warnf("unlock revoke err: %s", err.Error())
	}
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

	if c.Job.Kind != KindCommon {
		lk := c.lock()
		if lk == nil {
			return
		}
		defer lk.unlock()
	}

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

func (j *Job) limit() bool {
	if j.Parallels == 0 {
		return false
	}

	count := atomic.AddInt64(j.Count, 1)
	if j.Parallels < count {
		atomic.AddInt64(j.Count, -1)
		j.Fail(time.Now(), fmt.Sprintf("job[%s] running on[%s] running:[%d]", j.Key(), j.runOn, count))
		return true
	}

	return false
}

func (j *Job) unlimit() {
	if j.Parallels == 0 {
		return
	}
	atomic.AddInt64(j.Count, -1)
}

func (j *Job) Init(nodeID, hostname string) {
	var c int64
	j.Count, j.runOn, j.hostname = &c, nodeID, hostname
}

func (c *Cmd) lockTtl() int64 {
	now := time.Now()
	prev := c.JobRule.Schedule.Next(now)
	ttl := int64(c.JobRule.Schedule.Next(prev).Sub(prev) / time.Second)
	if ttl == 0 {
		return 0
	}

	if c.Job.Kind == KindInterval {
		ttl -= 2
		if ttl > conf.Config.LockTtl {
			ttl = conf.Config.LockTtl
		}
		if ttl < 1 {
			ttl = 1
		}
		return ttl
	}

	cost := c.Job.AvgTime / 1e3
	if c.Job.AvgTime/1e3-cost*1e3 > 0 {
		cost += 1
	}
	// 如果执行间隔时间不大于执行时间，把过期时间设置为执行时间的下限-1
	// 以便下次执行的时候，能获取到 lock
	if ttl >= cost {
		ttl -= cost
	}

	if ttl > conf.Config.LockTtl {
		ttl = conf.Config.LockTtl
	}

	// 支持的最小时间间隔 2s
	if ttl < 2 {
		ttl = 2
	}

	return ttl
}

func (c *Cmd) newLock() *locker {
	return &locker{
		kind: c.Job.Kind,
		ttl:  c.lockTtl(),
		done: make(chan struct{}),
	}
}

func (c *Cmd) lock() *locker {
	lk := c.newLock()
	// 非法的 rule
	if lk.ttl == 0 {
		return nil
	}

	resp, err := DefalutClient.Grant(lk.ttl)
	if err != nil {
		log.Infof("job[%s] didn't get a lock, err: %s", c.Job.Key(), err.Error())
		return nil
	}

	ok, err := DefalutClient.GetLock(c.Job.ID, resp.ID)
	if err != nil {
		log.Infof("job[%s] didn't get a lock, err: %s", c.Job.Key(), err.Error())
		return nil
	}

	if !ok {
		return nil
	}

	lk.lID = resp.ID
	if lk.kind == KindAlone {
		go lk.keepAlive()
	}
	return lk
}

// 优先取结点里的值，更新 group 时可用 gid 判断是否对 job 进行处理
func (rule *JobRule) included(nid string, gs map[string]*Group) bool {
	for i, count := 0, len(rule.NodeIDs); i < count; i++ {
		if nid == rule.NodeIDs[i] {
			return true
		}
	}

	for _, gid := range rule.GroupIDs {
		if g, ok := gs[gid]; ok && g.Included(nid) {
			return true
		}
	}

	return false
}

// 验证 timer 字段
func (rule *JobRule) Valid() error {
	// 注意 interface nil 的比较
	if rule.Schedule != nil {
		return nil
	}

	if len(rule.Timer) == 0 {
		return ErrNilRule
	}

	sch, err := cron.Parse(rule.Timer)
	if err != nil {
		return fmt.Errorf("invalid JobRule[%s], parse err: %s", rule.Timer, err.Error())
	}

	rule.Schedule = sch
	return nil
}

// Note: this function did't check the job.
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

		job.alone()
		jobs[job.ID] = job
	}
	return
}

func WatchJobs() client.WatchChan {
	return DefalutClient.Watch(conf.Config.Cmd, client.WithPrefix())
}

func GetJobFromKv(key, value []byte) (job *Job, err error) {
	job = new(Job)
	if err = json.Unmarshal(value, job); err != nil {
		err = fmt.Errorf("job[%s] umarshal err: %s", string(key), err.Error())
		return
	}

	err = job.Valid()
	job.alone()
	return
}

func (j *Job) alone() {
	if j.Kind == KindAlone {
		j.Parallels = 1
	}
}

func (j *Job) splitCmd() {
	ps := strings.SplitN(j.Command, " ", 2)
	if len(ps) == 1 {
		j.cmd = ps
		return
	}

	j.cmd = make([]string, 0, 2)
	j.cmd = append(j.cmd, ps[0])
	j.cmd = append(j.cmd, utils.ParseCmdArguments(ps[1])...)
}

func (j *Job) String() string {
	data, err := json.Marshal(j)
	if err != nil {
		return err.Error()
	}
	return string(data)
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
		if uid != _Uid {
			gid, _ := strconv.Atoi(u.Gid)
			sysProcAttr = &syscall.SysProcAttr{
				Credential: &syscall.Credential{
					Uid: uint32(uid),
					Gid: uint32(gid),
				},
			}
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
		j.Fail(t, fmt.Sprintf("%s\n%s", b.String(), err.Error()))
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
		j.Fail(t, fmt.Sprintf("%s\n%s", b.String(), err.Error()))
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

	if j.LogExpiration < 0 {
		j.LogExpiration = 0
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
	j.Notify(t, msg)
	CreateJobLog(j, t, msg, false)
}

func (j *Job) Notify(t time.Time, msg string) {
	if !conf.Config.Mail.Enable || !j.FailNotify {
		return
	}

	ts := t.Format(time.RFC3339)
	body := "job: " + j.Key() + "\n" +
		"job name: " + j.Name + "\n" +
		"job cmd: " + j.Command + "\n" +
		"node: " + j.runOn + "\n" +
		"time: " + ts + "\n" +
		"err: " + msg

	m := Message{
		Subject: "node[" + j.runOn + "] job[" + j.ShortName() + "] time[" + ts + "] exec failed",
		Body:    body,
		To:      j.To,
	}

	data, err := json.Marshal(m)
	if err != nil {
		log.Warnf("job[%s] send notice fail, err: %s", j.Key(), err.Error())
		return
	}

	_, err = DefalutClient.Put(conf.Config.Noticer+"/"+j.runOn, string(data))
	if err != nil {
		log.Warnf("job[%s] send notice fail, err: %s", j.Key(), err.Error())
		return
	}
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

LOOP_TIMER_CMD:
	for _, r := range j.Rules {
		for _, id := range r.ExcludeNodeIDs {
			if nid == id {
				// 在当前定时器规则中，任务不会在该节点执行（节点被排除）
				// 但是任务可以在其它定时器中，在该节点被执行
				// 比如，一个定时器设置在凌晨 1 点执行，但是此时不想在这个节点执行，然后，
				// 同时又设置一个定时器在凌晨 2 点执行，这次这个任务由于某些原因，必须在当前节点执行
				// 下面的 LOOP_TIMER 标签，原因同上
				continue LOOP_TIMER_CMD
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
LOOP_TIMER:
	for _, r := range j.Rules {
		for _, id := range r.ExcludeNodeIDs {
			if nid == id {
				continue LOOP_TIMER
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

	if err := j.ValidRules(); err != nil {
		return err
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

func (j *Job) ValidRules() error {
	for _, r := range j.Rules {
		if err := r.Valid(); err != nil {
			return err
		}
	}
	return nil
}

func (j *Job) ShortName() string {
	if len(j.Name) <= 10 {
		return j.Name
	}

	names := []rune(j.Name)
	if len(names) <= 10 {
		return j.Name
	}

	return string(names[:10]) + "..."
}
