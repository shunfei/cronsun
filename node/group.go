package node

import (
	"github.com/shunfei/cronsun"
)

type Groups map[string]*cronsun.Group

type jobLink struct {
	gname string
	// rule id
	rules map[string]bool
}

// map[group id]map[job id]*jobLink
// 用于 group 发生变化的时候修改相应的 job
type link map[string]map[string]*jobLink

func newLink(size int) link {
	return make(link, size)
}

func (l link) add(gid, jid, rid, gname string) {
	js, ok := l[gid]
	if !ok {
		js = make(map[string]*jobLink, 4)
		l[gid] = js
	}

	j, ok := js[jid]
	if !ok {
		j = &jobLink{
			gname: gname,
			rules: make(map[string]bool),
		}
		js[jid] = j
	}

	j.rules[rid] = true
}

func (l link) addJob(job *cronsun.Job) {
	for _, r := range job.Rules {
		for _, gid := range r.GroupIDs {
			l.add(gid, job.ID, r.ID, job.Group)
		}
	}
}

func (l link) del(gid, jid, rid string) {
	js, ok := l[gid]
	if !ok {
		return
	}

	j, ok := js[jid]
	if !ok {
		return
	}

	delete(j.rules, rid)
	if len(j.rules) == 0 {
		delete(js, jid)
	}
}

func (l link) delJob(job *cronsun.Job) {
	for _, r := range job.Rules {
		for _, gid := range r.GroupIDs {
			l.delGroupJob(gid, job.ID)
		}
	}
}

func (l link) delGroupJob(gid, jid string) {
	js, ok := l[gid]
	if !ok {
		return
	}

	delete(js, jid)
}

func (l link) delGroup(gid string) {
	delete(l, gid)
}
