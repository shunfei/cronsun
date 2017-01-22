package web

import (
	"encoding/json"
	"net/http"
	"sort"
	"strings"

	"github.com/coreos/etcd/clientv3"
	"github.com/gorilla/mux"

	"sunteng/cronsun/conf"
	"sunteng/cronsun/models"
)

type Job struct{}

func (j *Job) GetJob(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	job, err := models.GetJob(vars["group"], vars["id"])
	var statusCode int
	if err != nil {
		if err == models.ErrNotFound {
			statusCode = http.StatusNotFound
		} else {
			statusCode = http.StatusInternalServerError
		}
		outJSONError(w, statusCode, err.Error())
		return
	}

	outJSON(w, job)
}

func (j *Job) DeleteJob(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	_, err := models.DeleteJob(vars["group"], vars["id"])
	if err != nil {
		outJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	outJSONWithCode(w, http.StatusNoContent, nil)
}

func (j *Job) ChangeJobStatus(w http.ResponseWriter, r *http.Request) {
	job := &models.Job{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&job)
	if err != nil {
		outJSONError(w, http.StatusBadRequest, err.Error())
		return
	}
	r.Body.Close()

	vars := mux.Vars(r)
	originJob, rev, err := models.GetJobAndRev(vars["group"], vars["id"])
	if err != nil {
		outJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	originJob.Pause = job.Pause
	b, err := json.Marshal(originJob)
	if err != nil {
		outJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	_, err = models.DefalutClient.PutWithModRev(originJob.Key(), string(b), rev)
	if err != nil {
		outJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	outJSON(w, originJob)
}

func (j *Job) UpdateJob(w http.ResponseWriter, r *http.Request) {
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

	_, err = models.DefalutClient.Put(job.Key(), string(b))
	if err != nil {
		outJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	outJSONWithCode(w, successCode, nil)
}

func (j *Job) GetGroups(w http.ResponseWriter, r *http.Request) {
	resp, err := models.DefalutClient.Get(conf.Config.Cmd, clientv3.WithPrefix(), clientv3.WithKeysOnly())
	if err != nil {
		outJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	var cmdKeyLen = len(conf.Config.Cmd)
	var groupMap = make(map[string]bool, 8)

	for i := range resp.Kvs {
		ss := strings.Split(string(resp.Kvs[i].Key)[cmdKeyLen:], "/")
		groupMap[ss[0]] = true
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
	resp, err := models.DefalutClient.Get(conf.Config.Cmd+vars["name"], clientv3.WithPrefix(), clientv3.WithSort(clientv3.SortByKey, clientv3.SortAscend))
	if err != nil {
		outJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	var jobList = make([]*models.Job, 0, resp.Count)
	for i := range resp.Kvs {
		job := models.Job{}
		err = json.Unmarshal(resp.Kvs[i].Value, &job)
		if err != nil {
			outJSONError(w, http.StatusInternalServerError, err.Error())
			return
		}
		jobList = append(jobList, &job)
	}

	outJSON(w, jobList)
}
