package web

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	v3 "github.com/coreos/etcd/clientv3"
	"github.com/gorilla/mux"
	"gopkg.in/mgo.v2/bson"

	"github.com/shunfei/cronsun"
	"github.com/shunfei/cronsun/conf"
	"github.com/shunfei/cronsun/log"
)

type Node struct{}

var ngKeyDeepLen = len(conf.Config.Group)

func (n *Node) UpdateGroup(ctx *Context) {
	g := cronsun.Group{}
	de := json.NewDecoder(ctx.R.Body)
	var err error
	if err = de.Decode(&g); err != nil {
		outJSONWithCode(ctx.W, http.StatusBadRequest, err.Error())
		return
	}
	defer ctx.R.Body.Close()

	var successCode = http.StatusOK
	g.ID = strings.TrimSpace(g.ID)
	if len(g.ID) == 0 {
		successCode = http.StatusCreated
		g.ID = cronsun.NextID()
	}

	if err = g.Check(); err != nil {
		outJSONWithCode(ctx.W, http.StatusBadRequest, err.Error())
		return
	}

	// @TODO modRev
	var modRev int64 = 0
	if _, err = g.Put(modRev); err != nil {
		outJSONWithCode(ctx.W, http.StatusBadRequest, err.Error())
		return
	}

	outJSONWithCode(ctx.W, successCode, nil)
}

func (n *Node) GetGroups(ctx *Context) {
	list, err := cronsun.GetNodeGroups()
	if err != nil {
		outJSONWithCode(ctx.W, http.StatusInternalServerError, err.Error())
		return
	}

	outJSON(ctx.W, list)
}

func (n *Node) GetGroupByGroupId(ctx *Context) {
	vars := mux.Vars(ctx.R)
	g, err := cronsun.GetGroupById(vars["id"])
	if err != nil {
		outJSONWithCode(ctx.W, http.StatusInternalServerError, err.Error())
		return
	}

	if g == nil {
		outJSONWithCode(ctx.W, http.StatusNotFound, nil)
		return
	}
	outJSON(ctx.W, g)
}

func (n *Node) DeleteGroup(ctx *Context) {
	vars := mux.Vars(ctx.R)
	groupId := strings.TrimSpace(vars["id"])
	if len(groupId) == 0 {
		outJSONWithCode(ctx.W, http.StatusBadRequest, "empty node ground id.")
		return
	}

	_, err := cronsun.DeleteGroupById(groupId)
	if err != nil {
		outJSONWithCode(ctx.W, http.StatusInternalServerError, err.Error())
		return
	}

	gresp, err := cronsun.DefalutClient.Get(conf.Config.Cmd, v3.WithPrefix())
	if err != nil {
		errstr := fmt.Sprintf("failed to fetch jobs from etcd after deleted node group[%s]: %s", groupId, err.Error())
		log.Errorf(errstr)
		outJSONWithCode(ctx.W, http.StatusInternalServerError, errstr)
		return
	}

	// update rule's node group
	for i := range gresp.Kvs {
		job := cronsun.Job{}
		err = json.Unmarshal(gresp.Kvs[i].Value, &job)
		key := string(gresp.Kvs[i].Key)
		if err != nil {
			log.Errorf("failed to unmarshal job[%s]: %s", key, err.Error())
			continue
		}

		update := false
		for j := range job.Rules {
			var ngs []string
			for _, gid := range job.Rules[j].GroupIDs {
				if gid != groupId {
					ngs = append(ngs, gid)
				}
			}
			if len(ngs) != len(job.Rules[j].GroupIDs) {
				job.Rules[j].GroupIDs = ngs
				update = true
			}
		}

		if update {
			v, err := json.Marshal(&job)
			if err != nil {
				log.Errorf("failed to marshal job[%s]: %s", key, err.Error())
				continue
			}
			_, err = cronsun.DefalutClient.PutWithModRev(key, string(v), gresp.Kvs[i].ModRevision)
			if err != nil {
				log.Errorf("failed to update job[%s]: %s", key, err.Error())
				continue
			}
		}
	}

	outJSONWithCode(ctx.W, http.StatusNoContent, nil)
}

func (n *Node) GetNodes(ctx *Context) {
	nodes, err := cronsun.GetNodes()
	if err != nil {
		outJSONWithCode(ctx.W, http.StatusInternalServerError, err.Error())
		return
	}

	gresp, err := cronsun.DefalutClient.Get(conf.Config.Node, v3.WithPrefix(), v3.WithKeysOnly())
	if err == nil {
		connecedMap := make(map[string]bool, gresp.Count)
		for i := range gresp.Kvs {
			k := string(gresp.Kvs[i].Key)
			index := strings.LastIndexByte(k, '/')
			connecedMap[k[index+1:]] = true
		}

		for i := range nodes {
			nodes[i].Connected = connecedMap[nodes[i].ID]
		}
	} else {
		log.Errorf("failed to fetch key[%s] from etcd: %s", conf.Config.Node, err.Error())
	}

	outJSONWithCode(ctx.W, http.StatusOK, nodes)
}

// DeleteNode force remove node (by ip) which state in offline or damaged.
func (n *Node) DeleteNode(ctx *Context) {
	vars := mux.Vars(ctx.R)
	ip := strings.TrimSpace(vars["ip"])
	if len(ip) == 0 {
		outJSONWithCode(ctx.W, http.StatusBadRequest, "node ip is required.")
		return
	}

	resp, err := cronsun.DefalutClient.Get(conf.Config.Node + ip)
	if err != nil {
		outJSONWithCode(ctx.W, http.StatusInternalServerError, err.Error())
		return
	}

	if len(resp.Kvs) > 0 {
		outJSONWithCode(ctx.W, http.StatusBadRequest, "can not remove a running node.")
		return
	}

	err = cronsun.RemoveNode(bson.M{"_id": ip})
	if err != nil {
		outJSONWithCode(ctx.W, http.StatusInternalServerError, err.Error())
		return
	}

	// remove node from group
	var errmsg = "failed to remove node %s from groups, please remove it manually: %s"
	resp, err = cronsun.DefalutClient.Get(conf.Config.Group, v3.WithPrefix())
	if err != nil {
		outJSONWithCode(ctx.W, http.StatusInternalServerError, fmt.Sprintf(errmsg, ip, err.Error()))
		return
	}

	for i := range resp.Kvs {
		g := cronsun.Group{}
		err = json.Unmarshal(resp.Kvs[i].Value, &g)
		if err != nil {
			outJSONWithCode(ctx.W, http.StatusInternalServerError, fmt.Sprintf(errmsg, ip, err.Error()))
			return
		}

		var nids = make([]string, 0, len(g.NodeIDs))
		for _, nid := range g.NodeIDs {
			if nid != ip {
				nids = append(nids, nid)
			}
		}
		g.NodeIDs = nids

		if _, err = g.Put(resp.Kvs[i].ModRevision); err != nil {
			outJSONWithCode(ctx.W, http.StatusInternalServerError, fmt.Sprintf(errmsg, ip, err.Error()))
			return
		}
	}

	outJSONWithCode(ctx.W, http.StatusNoContent, nil)
}
