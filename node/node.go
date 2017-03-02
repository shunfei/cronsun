package node

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"golang.org/x/net/context"

	client "github.com/coreos/etcd/clientv3"

	"sunteng/commons/log"
	"sunteng/commons/util"
	"sunteng/cronsun/conf"
	"sunteng/cronsun/models"
	"sunteng/cronsun/node/cron"
)

// Node 执行 cron 命令服务的结构体
type Node struct {
	*models.Client
	*models.Node
	*cron.Cron

	jobs   Jobs
	groups Groups
	cmds   map[string]*models.Cmd

	link
	// 删除的 job id，用于 group 更新
	delIDs map[string]bool

	ttl int64

	lID client.LeaseID // lease id
	lch <-chan *client.LeaseKeepAliveResponse

	done chan struct{}
}

func NewNode(cfg *conf.Conf) (n *Node, err error) {
	ip, err := util.GetLocalIP()
	if err != nil {
		return
	}

	n = &Node{
		Client: models.DefalutClient,
		Node: &models.Node{
			ID:  ip.String(),
			PID: strconv.Itoa(os.Getpid()),
		},
		Cron: cron.New(),

		cmds: make(map[string]*models.Cmd),

		ttl:  cfg.Ttl,
		done: make(chan struct{}),
	}
	return
}

// 注册到 /cronsun/proc/xx
func (n *Node) Register() (err error) {
	pid, err := n.Node.Exist()
	if err != nil {
		return
	}

	if pid != -1 {
		return fmt.Errorf("node[%s] pid[%d] exist", n.Node.ID, pid)
	}

	resp, err := n.Client.Grant(context.TODO(), n.ttl)
	if err != nil {
		return
	}

	if _, err = n.Node.Put(client.WithLease(resp.ID)); err != nil {
		return
	}

	ch, err := n.Client.KeepAlive(context.TODO(), resp.ID)
	if err != nil {
		return
	}

	n.lID, n.lch = resp.ID, ch
	return
}

func (n *Node) loadJobs() (err error) {
	if n.groups, err = models.GetGroups(""); err != nil {
		return
	}

	jobs, err := models.GetJobs()
	if err != nil {
		return
	}

	n.jobs, n.link = make(Jobs, len(jobs)), newLink(len(n.groups))
	if len(jobs) == 0 {
		return
	}

	for _, job := range jobs {
		n.addJob(job, false)
	}

	return
}

func (n *Node) addJob(job *models.Job, notice bool) {
	n.link.addJob(job)

	cmds := job.Cmds(n.ID, n.groups)
	if len(cmds) == 0 {
		return
	}

	n.jobs[job.ID] = job
	for _, cmd := range cmds {
		n.addCmd(cmd, notice)
	}
	return
}

func (n *Node) delJob(id string) {
	job, ok := n.jobs[id]
	// 之前此任务没有在当前结点执行
	if !ok {
		n.delIDs[id] = true
		return
	}

	delete(n.jobs, id)
	n.link.delJob(job)

	cmds := job.Cmds(n.ID, n.groups)
	if len(cmds) == 0 {
		return
	}

	for _, cmd := range cmds {
		n.delCmd(cmd)
	}
	return
}

func (n *Node) modJob(job *models.Job) {
	oJob, ok := n.jobs[job.ID]
	// 之前此任务没有在当前结点执行，直接增加任务
	if !ok {
		n.addJob(job, true)
		return
	}

	n.link.delJob(oJob)
	prevCmds := oJob.Cmds(n.ID, n.groups)
	*oJob = *job
	cmds := oJob.Cmds(n.ID, n.groups)

	for id, cmd := range cmds {
		n.addCmd(cmd, true)
		delete(prevCmds, id)
	}

	for _, cmd := range prevCmds {
		n.delCmd(cmd)
	}

	n.link.addJob(oJob)
}

func (n *Node) addCmd(cmd *models.Cmd, notice bool) {
	c, ok := n.cmds[cmd.GetID()]
	if ok {
		sch := c.JobRule.Timer
		*c = *cmd

		// 节点执行时间不变，不用更新 cron
		if c.JobRule.Timer == sch {
			return
		}
	} else {
		c = cmd
	}

	if err := n.Cron.AddJob(c.JobRule.Timer, c); err != nil {
		msg := fmt.Sprintf("job[%s] rule[%s] timer[%s] parse err: %s", c.Job.ID, c.JobRule.ID, c.JobRule.Timer, err.Error())
		log.Warn(msg)
		c.Fail(time.Now(), msg)
		return
	}

	if !ok {
		n.cmds[c.GetID()] = c
	}

	if notice {
		log.Noticef("job[%s] rule[%s] timer[%s] has added", c.Job.ID, c.JobRule.ID, c.JobRule.Timer)
	}
	return
}

func (n *Node) delCmd(cmd *models.Cmd) {
	delete(n.cmds, cmd.GetID())
	n.Cron.DelJob(cmd)
	log.Noticef("job[%s] rule[%s] timer[%s] has deleted", cmd.Job.ID, cmd.JobRule.ID, cmd.JobRule.Timer)
}

func (n *Node) addGroup(g *models.Group) {
	n.groups[g.ID] = g
}

func (n *Node) delGroup(id string) {
	delete(n.groups, id)
	n.link.delGroup(id)

	job, ok := n.jobs[id]
	// 之前此任务没有在当前结点执行
	if !ok {
		return
	}

	cmds := job.Cmds(n.ID, n.groups)
	if len(cmds) == 0 {
		return
	}

	for _, cmd := range cmds {
		n.delCmd(cmd)
	}
	return
}

func (n *Node) modGroup(g *models.Group) {
	oGroup, ok := n.groups[g.ID]
	if !ok {
		n.addGroup(g)
		return
	}

	// 都包含/都不包含当前节点，对当前节点任务无影响
	if (oGroup.Included(n.ID) && g.Included(n.ID)) || (!oGroup.Included(n.ID) && !g.Included(n.ID)) {
		*oGroup = *g
		return
	}

	// 增加当前节点
	if !oGroup.Included(n.ID) && g.Included(n.ID) {
		n.groupAddNode(g)
		return
	}

	// 移除当前节点
	n.groupRmNode(g, oGroup)
	return
}

func (n *Node) groupAddNode(g *models.Group) {
	n.groups[g.ID] = g
	jls := n.link[g.ID]
	if len(jls) == 0 {
		return
	}

	var err error
	for jid, jl := range jls {
		job, ok := n.jobs[jid]
		if !ok {
			// job 已删除
			if n.delIDs[jid] {
				n.link.delGroupJob(g.ID, jid)
				continue
			}

			if job, err = models.GetJob(jl.gname, jid); err != nil {
				log.Warnf("get job[%s][%s] err: %s", jl.gname, jid, err.Error())
				n.link.delGroupJob(g.ID, jid)
				continue
			}
		}

		cmds := job.Cmds(n.ID, n.groups)
		for _, cmd := range cmds {
			n.addCmd(cmd, true)
		}
	}
	return
}

func (n *Node) groupRmNode(g, og *models.Group) {
	jls := n.link[g.ID]
	if len(jls) == 0 {
		n.groups[g.ID] = g
		return
	}

	for jid, _ := range jls {
		job, ok := n.jobs[jid]
		if !ok {
			// 数据出错
			log.Warnf("WTF! group[%s] job[%s]", g.ID, jid)
			n.link.delGroupJob(g.ID, jid)
			continue
		}

		n.groups[og.ID] = og
		prevCmds := job.Cmds(n.ID, n.groups)
		n.groups[g.ID] = g
		cmds := job.Cmds(n.ID, n.groups)

		for id, cmd := range cmds {
			n.addCmd(cmd, true)
			delete(prevCmds, id)
		}

		for _, cmd := range prevCmds {
			n.delCmd(cmd)
		}
	}

	n.groups[g.ID] = g
}

func (n *Node) watchJobs() {
	rch := models.WatchJobs()
	for wresp := range rch {
		for _, ev := range wresp.Events {
			switch {
			case ev.IsCreate():
				job, err := models.GetJobFromKv(ev.Kv)
				if err != nil {
					log.Warnf(err.Error())
					continue
				}

				n.addJob(job, true)
			case ev.IsModify():
				job, err := models.GetJobFromKv(ev.Kv)
				if err != nil {
					log.Warnf(err.Error())
					continue
				}

				n.modJob(job)
			case ev.Type == client.EventTypeDelete:
				n.delJob(models.GetIDFromKey(string(ev.Kv.Key)))
			default:
				log.Warnf("unknown event type[%v] from job[%s]", ev.Type, string(ev.Kv.Key))
			}
		}
	}
}

func (n *Node) watchGroups() {
	rch := models.WatchGroups()
	for wresp := range rch {
		for _, ev := range wresp.Events {
			switch {
			case ev.IsCreate():
				g, err := models.GetGroupFromKv(ev.Kv)
				if err != nil {
					log.Warnf(err.Error())
					continue
				}

				n.addGroup(g)
			case ev.IsModify():
				g, err := models.GetGroupFromKv(ev.Kv)
				if err != nil {
					log.Warnf(err.Error())
					continue
				}

				n.modGroup(g)
			case ev.Type == client.EventTypeDelete:
				n.delGroup(models.GetIDFromKey(string(ev.Kv.Key)))
			default:
				log.Warnf("unknown event type[%v] from group[%s]", ev.Type, string(ev.Kv.Key))
			}
		}
	}
}

// 启动服务
func (n *Node) Run() (err error) {
	go n.keepAlive()

	defer func() {
		if err != nil {
			n.Stop(nil)
		}
	}()

	if err = n.loadJobs(); err != nil {
		return
	}

	n.Cron.Start()
	go n.watchJobs()
	go n.watchGroups()
	n.Node.On()
	return
}

// 断网掉线重新注册
func (n *Node) keepAlive() {
	for {
		for _ = range n.lch {
		}

		select {
		case <-n.done:
			return
		default:
		}
		time.Sleep(time.Duration(n.ttl+1) * time.Second)

		log.Noticef("%s has dropped, try to reconnect...", n.String())
		if err := n.Register(); err != nil {
			log.Warn(err.Error())
		} else {
			log.Noticef("%s reconnected", n.String())
		}
	}
}

// 停止服务
func (n *Node) Stop(i interface{}) {
	n.Node.Down()
	close(n.done)
	n.Node.Del()
	n.Client.Close()
	n.Cron.Stop()
}
