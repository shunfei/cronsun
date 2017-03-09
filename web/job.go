package web

import (
	"encoding/json"
	"net/http"
	"sort"
	"strings"

	"github.com/coreos/etcd/clientv3"
	"github.com/gorilla/mux"

	"sunteng/commons/log"
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
		outJSONWithCode(w, statusCode, err.Error())
		return
	}

	outJSON(w, job)
}

func (j *Job) DeleteJob(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	_, err := models.DeleteJob(vars["group"], vars["id"])
	if err != nil {
		outJSONWithCode(w, http.StatusInternalServerError, err.Error())
		return
	}

	outJSONWithCode(w, http.StatusNoContent, nil)
}

func (j *Job) ChangeJobStatus(w http.ResponseWriter, r *http.Request) {
	job := &models.Job{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&job)
	if err != nil {
		outJSONWithCode(w, http.StatusBadRequest, err.Error())
		return
	}
	r.Body.Close()

	vars := mux.Vars(r)
	originJob, rev, err := models.GetJobAndRev(vars["group"], vars["id"])
	if err != nil {
		outJSONWithCode(w, http.StatusInternalServerError, err.Error())
		return
	}

	originJob.Pause = job.Pause
	b, err := json.Marshal(originJob)
	if err != nil {
		outJSONWithCode(w, http.StatusInternalServerError, err.Error())
		return
	}

	_, err = models.DefalutClient.PutWithModRev(originJob.Key(), string(b), rev)
	if err != nil {
		outJSONWithCode(w, http.StatusInternalServerError, err.Error())
		return
	}

	outJSON(w, originJob)
}

func (j *Job) UpdateJob(w http.ResponseWriter, r *http.Request) {
	var job = &struct {
		*models.Job
		OldGroup string `json:"oldGroup"`
	}{}

	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&job)
	if err != nil {
		outJSONWithCode(w, http.StatusBadRequest, err.Error())
		return
	}
	r.Body.Close()

	if err = job.Check(); err != nil {
		outJSONWithCode(w, http.StatusBadRequest, err.Error())
		return
	}

	var deleteOldKey string
	var successCode = http.StatusOK
	if len(job.ID) == 0 {
		successCode = http.StatusCreated
		job.ID = models.NextID()
	} else {
		job.OldGroup = strings.TrimSpace(job.OldGroup)
		if job.OldGroup != job.Group {
			deleteOldKey = models.JobKey(job.OldGroup, job.ID)
		}
	}

	b, err := json.Marshal(job)
	if err != nil {
		outJSONWithCode(w, http.StatusInternalServerError, err.Error())
		return
	}

	_, err = models.DefalutClient.Put(job.Key(), string(b))
	if err != nil {
		outJSONWithCode(w, http.StatusInternalServerError, err.Error())
		return
	}

	// remove old key
	if len(deleteOldKey) > 0 {
		if _, err = models.DefalutClient.Delete(deleteOldKey); err != nil {
			log.Errorf("failed to remove old job key[%s], err: %s.", deleteOldKey, err.Error())
		}
	}

	outJSONWithCode(w, successCode, nil)
}

func (j *Job) GetGroups(w http.ResponseWriter, r *http.Request) {
	resp, err := models.DefalutClient.Get(conf.Config.Cmd, clientv3.WithPrefix(), clientv3.WithKeysOnly())
	if err != nil {
		outJSONWithCode(w, http.StatusInternalServerError, err.Error())
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

func (j *Job) GetList(w http.ResponseWriter, r *http.Request) {
	group := strings.TrimSpace(r.FormValue("group"))
	var prefix = conf.Config.Cmd
	if len(group) != 0 {
		prefix += group
	}

	type jobStatus struct {
		*models.Job
		LatestStatus *models.JobLatestLog `json:"latestStatus"`
	}

	resp, err := models.DefalutClient.Get(prefix, clientv3.WithPrefix(), clientv3.WithSort(clientv3.SortByKey, clientv3.SortAscend))
	if err != nil {
		outJSONWithCode(w, http.StatusInternalServerError, err.Error())
		return
	}

	var jobIds []string
	var jobList = make([]*jobStatus, 0, resp.Count)
	for i := range resp.Kvs {
		job := models.Job{}
		err = json.Unmarshal(resp.Kvs[i].Value, &job)
		if err != nil {
			outJSONWithCode(w, http.StatusInternalServerError, err.Error())
			return
		}
		jobList = append(jobList, &jobStatus{Job: &job})
		jobIds = append(jobIds, job.ID)
	}

	m, err := models.GetJobLatestLogListByJobIds(jobIds)
	if err != nil {
		log.Error("GetJobLatestLogListByJobIds error:", err.Error())
	} else {
		for i := range jobList {
			jobList[i].LatestStatus = m[jobList[i].ID]
		}
	}

	outJSON(w, jobList)
}
