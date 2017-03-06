package web

import (
	"encoding/json"
	"net/http"
	"strings"

	v3 "github.com/coreos/etcd/clientv3"
	"github.com/gorilla/mux"

	"fmt"
	"sunteng/commons/log"
	"sunteng/cronsun/conf"
	"sunteng/cronsun/models"
)

type Node struct{}

var ngKeyDeepLen = len(conf.Config.Group)

func (n *Node) UpdateGroup(w http.ResponseWriter, r *http.Request) {
	g := models.Group{}
	de := json.NewDecoder(r.Body)
	var err error
	if err = de.Decode(&g); err != nil {
		outJSONError(w, http.StatusBadRequest, err.Error())
		return
	}
	defer r.Body.Close()

	var successCode = http.StatusOK
	g.ID = strings.TrimSpace(g.ID)
	if len(g.ID) == 0 {
		successCode = http.StatusCreated
		g.ID = models.NextID()
	}

	if err = g.Check(); err != nil {
		outJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	// @TODO modRev
	var modRev int64 = 0
	if _, err = g.Put(modRev); err != nil {
		outJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	outJSONWithCode(w, successCode, nil)
}

func (n *Node) GetGroups(w http.ResponseWriter, r *http.Request) {
	resp, err := models.DefalutClient.Get(conf.Config.Group, v3.WithPrefix(), v3.WithSort(v3.SortByKey, v3.SortAscend))
	if err != nil {
		outJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	var list = make([]*models.Group, 0, resp.Count)
	for i := range resp.Kvs {
		g := models.Group{}
		err = json.Unmarshal(resp.Kvs[i].Value, &g)
		if err != nil {
			log.Errorf("node.GetGroups(key: %s) error: %s", string(resp.Kvs[i].Key), err.Error())
			outJSONError(w, http.StatusInternalServerError, err.Error())
			return
		}
		list = append(list, &g)
	}

	outJSON(w, list)
}

func (n *Node) GetGroupByGroupId(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	g, err := models.GetGroupById(vars["id"])
	if err != nil {
		outJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if g == nil {
		outJSONWithCode(w, http.StatusNotFound, nil)
		return
	}
	outJSON(w, g)
}

func (n *Node) DeleteGroup(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	groupId := strings.TrimSpace(vars["id"])
	if len(groupId) == 0 {
		outJSONError(w, http.StatusBadRequest, "empty node ground id.")
		return
	}

	_, err := models.DeleteGroupById(groupId)
	if err != nil {
		outJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	gresp, err := models.DefalutClient.Get(conf.Config.Cmd, v3.WithPrefix())
	if err != nil {
		errstr := fmt.Sprintf("failed to fetch jobs from etcd after deleted node group[%s]: %s", groupId, err.Error())
		log.Error(errstr)
		outJSONError(w, http.StatusInternalServerError, errstr)
		return
	}

	// update rule's node group
	for i := range gresp.Kvs {
		job := models.Job{}
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
			_, err = models.DefalutClient.PutWithModRev(key, string(v), gresp.Kvs[i].ModRevision)
			if err != nil {
				log.Errorf("failed to update job[%s]: %s", key, err.Error())
				continue
			}
		}
	}

	outJSONWithCode(w, http.StatusNoContent, nil)
}

func (n *Node) GetNodes(w http.ResponseWriter, r *http.Request) {
	nodes, err := models.GetNodes()
	if err != nil {
		outJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	gresp, err := models.DefalutClient.Get(conf.Config.Node, v3.WithPrefix(), v3.WithKeysOnly())
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

	outJSONWithCode(w, http.StatusOK, nodes)
}
