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
	"time"
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
	group := getStringVal("group", r)
	node := getStringVal("node", r)
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

	var nodeGroupMap map[string]*models.Group
	if len(node) > 0 {
		nodeGrouplist, err := models.GetNodeGroups()
		if err != nil {
			outJSONWithCode(w, http.StatusInternalServerError, err.Error())
			return
		}
		nodeGroupMap = map[string]*models.Group{}
		for i := range nodeGrouplist {
			nodeGroupMap[nodeGrouplist[i].ID] = nodeGrouplist[i]
		}
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

		if len(node) > 0 && !job.IsRunOn(node, nodeGroupMap) {
			continue
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

func (j *Job) GetJobNodes(w http.ResponseWriter, r *http.Request) {
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

	var nodes []string
	var exNodes []string
	groups, err := models.GetGroups("")
	if err != nil {
		outJSONWithCode(w, http.StatusInternalServerError, err.Error())
		return
	}

	for i := range job.Rules {
		inNodes := append(nodes, job.Rules[i].NodeIDs...)
		for _, gid := range job.Rules[i].GroupIDs {
			if g, ok := groups[gid]; ok {
				inNodes = append(inNodes, g.NodeIDs...)
			}
		}
		exNodes = append(exNodes, job.Rules[i].ExcludeNodeIDs...)
		inNodes = SubtractStringArray(inNodes, exNodes)
		nodes = append(nodes, inNodes...)
	}

	outJSON(w, UniqueStringArray(nodes))
}

func (j *Job) JobExecute(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	group := strings.TrimSpace(vars["group"])
	id := strings.TrimSpace(vars["id"])
	if len(group) == 0 || len(id) == 0 {
		outJSONWithCode(w, http.StatusBadRequest, "Invalid job id or group.")
		return
	}

	node := getStringVal("node", r)
	err := models.PutOnce(group, id, node)
	if err != nil {
		outJSONWithCode(w, http.StatusInternalServerError, err.Error())
		return
	}

	outJSONWithCode(w, http.StatusNoContent, nil)
}

func (j *Job) GetExecutingJob(w http.ResponseWriter, r *http.Request) {
	opt := &ProcFetchOptions{
		Groups:  getStringArrayFromQuery("groups", ",", r),
		NodeIds: getStringArrayFromQuery("nodes", ",", r),
		JobIds:  getStringArrayFromQuery("jobs", ",", r),
	}

	gresp, err := models.DefalutClient.Get(conf.Config.Proc, clientv3.WithPrefix())
	if err != nil {
		outJSONWithCode(w, http.StatusInternalServerError, err.Error())
		return
	}

	var list = make([]*models.Process, 0, 8)
	for i := range gresp.Kvs {
		proc, err := models.GetProcFromKey(string(gresp.Kvs[i].Key))
		if err != nil {
			log.Error("Failed to unmarshal Proc from key: ", err.Error())
			continue
		}

		if !opt.Match(proc) {
			continue
		}
		proc.Time, _ = time.Parse(time.RFC3339, string(gresp.Kvs[i].Value))
		list = append(list, proc)
	}

	sort.Sort(ByProcTime(list))
	outJSON(w, list)
}

type ProcFetchOptions struct {
	Groups  []string
	NodeIds []string
	JobIds  []string
}

func (opt *ProcFetchOptions) Match(proc *models.Process) bool {
	if len(opt.Groups) > 0 && !InStringArray(proc.Group, opt.Groups) {
		return false
	}

	if len(opt.JobIds) > 0 && !InStringArray(proc.JobID, opt.JobIds) {
		return false

	}

	if len(opt.NodeIds) > 0 && !InStringArray(proc.NodeID, opt.NodeIds) {
		return false
	}

	return true
}

type ByProcTime []*models.Process

func (a ByProcTime) Len() int           { return len(a) }
func (a ByProcTime) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByProcTime) Less(i, j int) bool { return a[i].Time.After(a[j].Time) }
