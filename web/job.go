package web

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/gorilla/mux"

	"github.com/shunfei/cronsun"
	"github.com/shunfei/cronsun/conf"
	"github.com/shunfei/cronsun/log"
)

type Job struct{}

func (j *Job) GetJob(ctx *Context) {
	vars := mux.Vars(ctx.R)
	job, err := cronsun.GetJob(vars["group"], vars["id"])
	var statusCode int
	if err != nil {
		if err == cronsun.ErrNotFound {
			statusCode = http.StatusNotFound
		} else {
			statusCode = http.StatusInternalServerError
		}
		outJSONWithCode(ctx.W, statusCode, err.Error())
		return
	}

	outJSON(ctx.W, job)
}

func (j *Job) DeleteJob(ctx *Context) {
	vars := mux.Vars(ctx.R)
	_, err := cronsun.DeleteJob(vars["group"], vars["id"])
	if err != nil {
		outJSONWithCode(ctx.W, http.StatusInternalServerError, err.Error())
		return
	}

	outJSONWithCode(ctx.W, http.StatusNoContent, nil)
}

func (j *Job) ChangeJobStatus(ctx *Context) {
	job := &cronsun.Job{}
	decoder := json.NewDecoder(ctx.R.Body)
	err := decoder.Decode(&job)
	if err != nil {
		outJSONWithCode(ctx.W, http.StatusBadRequest, err.Error())
		return
	}
	ctx.R.Body.Close()

	vars := mux.Vars(ctx.R)
	job, err = j.updateJobStatus(vars["group"], vars["id"], job.Pause)
	if err != nil {
		outJSONWithCode(ctx.W, http.StatusInternalServerError, err.Error())
		return
	}

	outJSON(ctx.W, job)
}

func (j *Job) updateJobStatus(group, id string, isPause bool) (*cronsun.Job, error) {
	originJob, rev, err := cronsun.GetJobAndRev(group, id)
	if err != nil {
		return nil, err
	}

	if originJob.Pause == isPause {
		return nil, err
	}

	originJob.Pause = isPause
	b, err := json.Marshal(originJob)
	if err != nil {
		return nil, err
	}

	_, err = cronsun.DefalutClient.PutWithModRev(originJob.Key(), string(b), rev)
	if err != nil {
		return nil, err
	}

	return originJob, nil
}

func (j *Job) BatchChangeJobStatus(ctx *Context) {
	var jobIds []string
	decoder := json.NewDecoder(ctx.R.Body)
	err := decoder.Decode(&jobIds)
	if err != nil {
		outJSONWithCode(ctx.W, http.StatusBadRequest, err.Error())
		return
	}
	ctx.R.Body.Close()

	vars := mux.Vars(ctx.R)
	op := vars["op"]
	var isPause bool
	switch op {
	case "pause":
		isPause = true
	case "start":
	default:
		outJSONWithCode(ctx.W, http.StatusBadRequest, "Unknow batch operation.")
		return
	}

	var updated int
	for i := range jobIds {
		id := strings.Split(jobIds[i], "/") // [Group, ID]
		if len(id) != 2 || id[0] == "" || id[1] == "" {
			continue
		}

		_, err = j.updateJobStatus(id[0], id[1], isPause)
		if err != nil {
			continue
		}
		updated++
	}

	outJSON(ctx.W, fmt.Sprintf("%d of %d updated.", updated, len(jobIds)))
}

func (j *Job) UpdateJob(ctx *Context) {
	var job = &struct {
		*cronsun.Job
		OldGroup string `json:"oldGroup"`
	}{}

	decoder := json.NewDecoder(ctx.R.Body)
	err := decoder.Decode(&job)
	if err != nil {
		outJSONWithCode(ctx.W, http.StatusBadRequest, err.Error())
		return
	}
	ctx.R.Body.Close()

	if err = job.Check(); err != nil {
		outJSONWithCode(ctx.W, http.StatusBadRequest, err.Error())
		return
	}

	var deleteOldKey string
	var successCode = http.StatusOK
	if len(job.ID) == 0 {
		successCode = http.StatusCreated
		job.ID = cronsun.NextID()
	} else {
		job.OldGroup = strings.TrimSpace(job.OldGroup)
		if job.OldGroup != job.Group {
			deleteOldKey = cronsun.JobKey(job.OldGroup, job.ID)
		}
	}

	b, err := json.Marshal(job)
	if err != nil {
		outJSONWithCode(ctx.W, http.StatusInternalServerError, err.Error())
		return
	}

	// remove old key
	// it should be before the put method
	if len(deleteOldKey) > 0 {
		if _, err = cronsun.DefalutClient.Delete(deleteOldKey); err != nil {
			log.Errorf("failed to remove old job key[%s], err: %s.", deleteOldKey, err.Error())
			outJSONWithCode(ctx.W, http.StatusInternalServerError, err.Error())
			return
		}
	}

	_, err = cronsun.DefalutClient.Put(job.Key(), string(b))
	if err != nil {
		outJSONWithCode(ctx.W, http.StatusInternalServerError, err.Error())
		return
	}

	outJSONWithCode(ctx.W, successCode, nil)
}

func (j *Job) GetGroups(ctx *Context) {
	resp, err := cronsun.DefalutClient.Get(conf.Config.Cmd, clientv3.WithPrefix(), clientv3.WithKeysOnly())
	if err != nil {
		outJSONWithCode(ctx.W, http.StatusInternalServerError, err.Error())
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
	outJSON(ctx.W, groupList)
}

func (j *Job) GetList(ctx *Context) {
	group := getStringVal("group", ctx.R)
	node := getStringVal("node", ctx.R)
	var prefix = conf.Config.Cmd
	if len(group) != 0 {
		prefix += group
	}

	type jobStatus struct {
		*cronsun.Job
		LatestStatus *cronsun.JobLatestLog `json:"latestStatus"`
	}

	resp, err := cronsun.DefalutClient.Get(prefix, clientv3.WithPrefix(), clientv3.WithSort(clientv3.SortByKey, clientv3.SortAscend))
	if err != nil {
		outJSONWithCode(ctx.W, http.StatusInternalServerError, err.Error())
		return
	}

	var nodeGroupMap map[string]*cronsun.Group
	if len(node) > 0 {
		nodeGrouplist, err := cronsun.GetNodeGroups()
		if err != nil {
			outJSONWithCode(ctx.W, http.StatusInternalServerError, err.Error())
			return
		}
		nodeGroupMap = map[string]*cronsun.Group{}
		for i := range nodeGrouplist {
			nodeGroupMap[nodeGrouplist[i].ID] = nodeGrouplist[i]
		}
	}

	var jobIds []string
	var jobList = make([]*jobStatus, 0, resp.Count)
	for i := range resp.Kvs {
		job := cronsun.Job{}
		err = json.Unmarshal(resp.Kvs[i].Value, &job)
		if err != nil {
			outJSONWithCode(ctx.W, http.StatusInternalServerError, err.Error())
			return
		}

		if len(node) > 0 && !job.IsRunOn(node, nodeGroupMap) {
			continue
		}
		jobList = append(jobList, &jobStatus{Job: &job})
		jobIds = append(jobIds, job.ID)
	}

	m, err := cronsun.GetJobLatestLogListByJobIds(jobIds)
	if err != nil {
		log.Errorf("GetJobLatestLogListByJobIds error: %s", err.Error())
	} else {
		for i := range jobList {
			jobList[i].LatestStatus = m[jobList[i].ID]
		}
	}

	outJSON(ctx.W, jobList)
}

func (j *Job) GetJobNodes(ctx *Context) {
	vars := mux.Vars(ctx.R)
	job, err := cronsun.GetJob(vars["group"], vars["id"])
	var statusCode int
	if err != nil {
		if err == cronsun.ErrNotFound {
			statusCode = http.StatusNotFound
		} else {
			statusCode = http.StatusInternalServerError
		}
		outJSONWithCode(ctx.W, statusCode, err.Error())
		return
	}

	var nodes []string
	var exNodes []string
	groups, err := cronsun.GetGroups("")
	if err != nil {
		outJSONWithCode(ctx.W, http.StatusInternalServerError, err.Error())
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

	outJSON(ctx.W, UniqueStringArray(nodes))
}

func (j *Job) JobExecute(ctx *Context) {
	vars := mux.Vars(ctx.R)
	group := strings.TrimSpace(vars["group"])
	id := strings.TrimSpace(vars["id"])
	if len(group) == 0 || len(id) == 0 {
		outJSONWithCode(ctx.W, http.StatusBadRequest, "Invalid job id or group.")
		return
	}

	node := getStringVal("node", ctx.R)
	err := cronsun.PutOnce(group, id, node)
	if err != nil {
		outJSONWithCode(ctx.W, http.StatusInternalServerError, err.Error())
		return
	}

	outJSONWithCode(ctx.W, http.StatusNoContent, nil)
}

func (j *Job) GetExecutingJob(ctx *Context) {
	opt := &ProcFetchOptions{
		Groups:  getStringArrayFromQuery("groups", ",", ctx.R),
		NodeIds: getStringArrayFromQuery("nodes", ",", ctx.R),
		JobIds:  getStringArrayFromQuery("jobs", ",", ctx.R),
	}

	gresp, err := cronsun.DefalutClient.Get(conf.Config.Proc, clientv3.WithPrefix())
	if err != nil {
		outJSONWithCode(ctx.W, http.StatusInternalServerError, err.Error())
		return
	}

	var list = make([]*cronsun.Process, 0, 8)
	for i := range gresp.Kvs {
		proc, err := cronsun.GetProcFromKey(string(gresp.Kvs[i].Key))
		if err != nil {
			log.Errorf("Failed to unmarshal Proc from key: %s", err.Error())
			continue
		}

		if !opt.Match(proc) {
			continue
		}
		proc.Time, _ = time.Parse(time.RFC3339, string(gresp.Kvs[i].Value))
		list = append(list, proc)
	}

	sort.Sort(ByProcTime(list))
	outJSON(ctx.W, list)
}

type ProcFetchOptions struct {
	Groups  []string
	NodeIds []string
	JobIds  []string
}

func (opt *ProcFetchOptions) Match(proc *cronsun.Process) bool {
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

type ByProcTime []*cronsun.Process

func (a ByProcTime) Len() int           { return len(a) }
func (a ByProcTime) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByProcTime) Less(i, j int) bool { return a[i].Time.After(a[j].Time) }
