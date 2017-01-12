package web

import (
	"encoding/json"
	"net/http"
	"path"
	"sort"
	"strings"

	"github.com/coreos/etcd/clientv3"
	"github.com/gorilla/mux"

	"sunteng/cronsun/conf"
	"sunteng/cronsun/models"
)

type Job struct{}

func (j *Job) Update(w http.ResponseWriter, r *http.Request) {
	job := &models.Job{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&job)
	if err != nil {
		outJSONError(w, http.StatusBadRequest, err.Error())
		return
	}
	r.Body.Close()

	if err = job.Check(); err != nil {
		outJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	var successCode = http.StatusOK
	if len(job.ID) == 0 {
		successCode = http.StatusCreated
		job.ID = models.NextID()
	}

	b, err := json.Marshal(job)
	if err != nil {
		outJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	_, err = models.DefalutClient.Put(path.Join(conf.Config.Cmd, job.Group, job.ID), string(b))
	if err != nil {
		outJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	outJSONWithCode(w, successCode, nil)
}

var cmdKeyDeepLen = len(strings.Split(conf.Config.Cmd, "/"))

func (j *Job) GetGroups(w http.ResponseWriter, r *http.Request) {
	resp, err := models.DefalutClient.Get(conf.Config.Cmd, clientv3.WithPrefix(), clientv3.WithKeysOnly())
	if err != nil {
		outJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	var groupMap = make(map[string]bool, 8)
	for i := range resp.Kvs {
		ss := strings.Split(string(resp.Kvs[i].Key), "/")
		groupMap[ss[cmdKeyDeepLen]] = true
	}

	var groupList = make([]string, 0, len(groupMap))
	for k := range groupMap {
		groupList = append(groupList, k)
	}

	sort.Strings(groupList)
	outJSON(w, groupList)

}

func (j *Job) GetListByGroupName(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	resp, err := models.DefalutClient.Get(path.Join(conf.Config.Cmd, vars["name"]), clientv3.WithPrefix(), clientv3.WithSort(clientv3.SortByKey, clientv3.SortAscend))
	if err != nil {
		outJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	var jobList = make([]*models.Job, 0, resp.Count)
	for i := range resp.Kvs {
		job := &models.Job{}
		err = json.Unmarshal(resp.Kvs[i].Value, &job)
		if err != nil {
			outJSONError(w, http.StatusInternalServerError, err.Error())
			return
		}
		jobList = append(jobList)
	}

	outJSON(w, jobList)
}
