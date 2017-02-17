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
	groups Group
	cmds   map[string]*models.Cmd
	// map[group id]map[job id]bool
	// 用于 group 发生变化的时候修改相应的 job
	link map[string]map[string]bool

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
	if n.groups, err = models.GetGroups(n.ID); err != nil {
		return
	}
	if n.jobs, err = loadJobs(n.ID, n.groups); err != nil {
		return
	}

	n.addCmds()
	return
}

func (n *Node) addCmds() {
	if len(n.jobs) == 0 {
		return
	}

	for _, job := range n.jobs {
		for _, cmd := range job.Cmds(n.ID, n.groups) {
			n.addCmd(cmd)
		}
	}
}

func (n *Node) addCmd(cmd *models.Cmd) {
	c, ok := n.cmds[cmd.GetID()]
	if ok {
		sch := c.Schedule
		*c = *cmd

		// 节点执行时间不变，不用更新 cron
		if c.Schedule == sch {
			return
		}
	} else {
		c = cmd
	}

	if err := n.Cron.AddJob(c.Schedule, c); err != nil {
		msg := fmt.Sprintf("job[%s] rule[%s] timer[%s] parse err: %s", c.Job.ID, c.JobRule.ID, c.Schedule, err.Error())
		log.Warn(msg)
		c.Fail(time.Now(), msg)
		return
	}

	if !ok {
		n.cmds[c.GetID()] = c
	}
	return
}

func (n *Node) addLink(gid, jid string) {
	if len(gid) == 0 {
		return
	}

	js, ok := n.link[gid]
	if !ok {
		js = make(map[string]bool, 4)
		n.link[gid] = js
	}

	js[jid] = true
}

func (n *Node) delLink(gid, jid string) {
	if len(gid) == 0 {
		return
	}

	js, ok := n.link[gid]
	if !ok {
		return
	}

	delete(js, jid)
}

func (n *Node) addJob(job *models.Job) bool {
	sch, gid := job.Schedule(n.ID, n.groups, false)
	if len(sch) == 0 {
		return false
	}

	j, ok := n.jobs[job.GetID()]
	if ok {
		if j != job {
			*j = *job
		}
	} else {
		j = job
		n.jobs[j.GetID()] = j
	}

	j.RunOn(n.Node.ID)
	if err := n.Cron.AddJob(sch, j); err != nil {
		msg := fmt.Sprintf("job[%s] timer[%s] parse err: %s", j.GetID(), sch, err.Error())
		log.Warn(msg)
		j.Fail(time.Now(), msg)
		delete(n.jobs, j.GetID())
		return false
	}

	n.addLink(gid, j.GetID())
	return true
}

func (n *Node) delJob(job *models.Job) {
	sch, gid := job.Schedule(n.ID, n.groups, false)
	if len(sch) == 0 {
		return
	}

	n.delLink(gid, job.GetID())
	delete(n.jobs, job.GetID())
	n.Cron.DelJob(job)
	log.Noticef("job[%s] has deleted", job)
}

func (n *Node) addGroup(g *models.Group) bool {
	if !g.Included(n.ID) {
		return false
	}

	if og, ok := n.groups[g.ID]; ok {
		*og = *g
		// TODO 处理相应的 jobs
		return true
	}

	n.groups[g.ID] = g
	return true
}

func (n *Node) delGroup(g *models.Group) {
	if !g.Included(n.ID) {
		return
	}

	delete(n.groups, g.ID)
	// TODO 处理相应的 jobs
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

				if n.addJob(job) {
					log.Noticef("job[%s] has added", job)
				}
			case ev.IsModify():
				job, err := models.GetJobFromKv(ev.Kv)
				if err != nil {
					log.Warnf(err.Error())
					continue
				}
				prevJob, err := models.GetJobFromKv(ev.PrevKv)
				if err != nil {
					log.Warnf(err.Error())
					continue
				}

				if n.addJob(job) {
					log.Noticef("job[%s] has added", job)
					continue
				}

				// 此结点暂停或不再执行此 job
				n.delJob(prevJob)
			case ev.Type == client.EventTypeDelete:
				prevJob, err := models.GetJobFromKv(ev.PrevKv)
				if err != nil {
					log.Warnf(err.Error())
					continue
				}

				n.delJob(prevJob)
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
				prevG, err := models.GetGroupFromKv(ev.PrevKv)
				if err != nil {
					log.Warnf(err.Error())
					continue
				}

				if n.addGroup(g) {
					continue
				}

				// 此 group 已移除当前结点
				n.delGroup(prevG)
			case ev.Type == client.EventTypeDelete:
				prevG, err := models.GetGroupFromKv(ev.PrevKv)
				if err != nil {
					log.Warnf(err.Error())
					continue
				}

				n.delGroup(prevG)
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
	// go n.watchJobs()
	// go n.watchGroups()
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
