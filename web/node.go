package web

import (
	"encoding/json"
	"net/http"
	"strings"

	v3 "github.com/coreos/etcd/clientv3"
	"github.com/gorilla/mux"

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

	// @TODO rev
	var rev int64 = 0
	if _, err = g.Put(rev); err != nil {
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
	_, err := models.DeleteGroupById(vars["id"])
	if err != nil {
		outJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	outJSONWithCode(w, http.StatusNoContent, nil)
}

func (n *Node) GetActivityNodeList(w http.ResponseWriter, r *http.Request) {
	ids, err := models.GetActivityNodeList()
	if err != nil {
		outJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
	outJSONWithCode(w, http.StatusOK, ids)
}
