package web

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"path"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/gorilla/mux"
	"golang.org/x/net/context"

	"fmt"
	"sunteng/commons/log"
	"sunteng/cronsun"
	"sunteng/cronsun/conf"
)

var etcdClient *clientv3.Client

func InitRouters() (s *http.Server, err error) {
	etcdClient, err = cronsun.EtcdInstance()
	if err != nil {
		return nil, err
	}

	r := mux.NewRouter()
	subrouter := r.PathPrefix("/v1").Subrouter()

	h := BaseHandler{Handle: getJobGroups}
	subrouter.Handle("/job/groups", h).Methods("GET")

	h = BaseHandler{Handle: getJobsByGroupName}
	subrouter.Handle("/job/group/{name}", h).Methods("GET")

	h = BaseHandler{Handle: updateJob}
	subrouter.Handle("/job", h).Methods("PUT")

	h = BaseHandler{Handle: getNodeGroups}
	subrouter.Handle("/node/groups", h).Methods("GET")

	h = BaseHandler{Handle: getNodeGroupByName}
	subrouter.Handle("/node/group/{name}", h).Methods("GET")

	h = BaseHandler{Handle: nodeJoinGroup}
	subrouter.Handle("/node/group", h).Methods("PUT")

	h = BaseHandler{Handle: nodeLeaveGroup}
	subrouter.Handle("/node/group", h).Methods("DELETE")

	s = &http.Server{
		Handler: r,
	}
	return s, nil
}

var cmdKeyDeepLen = len(strings.Split(conf.Config.Cmd, "/"))

func getJobGroups(w http.ResponseWriter, r *http.Request) {
	resp, err := etcdClient.Get(context.TODO(), conf.Config.Cmd, clientv3.WithPrefix(), clientv3.WithKeysOnly())
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

func getJobsByGroupName(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	resp, err := etcdClient.Get(context.TODO(), path.Join(conf.Config.Cmd, vars["name"]), clientv3.WithPrefix(), clientv3.WithSort(clientv3.SortByKey, clientv3.SortAscend))
	if err != nil {
		outJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	var jobList = make([]*cronsun.Job, 0, resp.Count)
	for i := range resp.Kvs {
		job := &cronsun.Job{}
		err = json.Unmarshal(resp.Kvs[i].Value, &job)
		if err != nil {
			outJSONError(w, http.StatusInternalServerError, err.Error())
			return
		}
		jobList = append(jobList)
	}

	outJSON(w, jobList)
}

func updateJob(w http.ResponseWriter, r *http.Request) {
	job := &cronsun.Job{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&job)
	if err != nil {
		outJSONError(w, http.StatusBadRequest, err.Error())
		return
	}
	r.Body.Close()

	var creation bool
	if len(job.Id) == 0 {
		creation = true
		now := time.Now()
		h := sha1.Sum([]byte(strconv.FormatInt(now.Unix(), 10) + strconv.FormatInt(now.UnixNano(), 10)))
		job.Id = hex.EncodeToString(h[:])
	}

	jobb, err := json.Marshal(job)
	if err != nil {
		outJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	_, err = etcdClient.Put(context.TODO(), path.Join(conf.Config.Cmd, job.Group, job.Id), string(jobb))
	if err != nil {
		outJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	statusCode := http.StatusOK
	if creation {
		statusCode = http.StatusCreated
	}
	outJSONWithCode(w, statusCode, nil)
}

var ngKeyDeepLen = len(conf.Config.NodeGroup)

func getNodeGroups(w http.ResponseWriter, r *http.Request) {
	resp, err := etcdClient.Get(context.TODO(), conf.Config.NodeGroup, clientv3.WithPrefix(), clientv3.WithKeysOnly())
	if err != nil {
		outJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	var groupMap = make(map[string]bool, 8)
	for i := range resp.Kvs {
		ss := strings.Split(string(resp.Kvs[i].Key), "/")
		groupMap[ss[ngKeyDeepLen]] = true
	}

	var groupList = make([]string, 0, len(groupMap))
	for k := range groupMap {
		groupList = append(groupList, k)
	}

	sort.Strings(groupList)
	outJSON(w, groupList)
}

func getNodeGroupByName(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	resp, err := etcdClient.Get(context.TODO(), path.Join(conf.Config.NodeGroup, vars["name"]), clientv3.WithPrefix(), clientv3.WithSort(clientv3.SortByKey, clientv3.SortAscend))
	if err != nil {
		outJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	var nodeList = make([]*cronsun.Node, 0, resp.Count)
	for i := range resp.Kvs {
		node := &cronsun.Node{}
		err = json.Unmarshal(resp.Kvs[i].Value, &node)
		if err != nil {
			outJSONError(w, http.StatusInternalServerError, err.Error())
			return
		}
		nodeList = append(nodeList)
	}

	outJSON(w, nodeList)
}

func nodeJoinGroup(w http.ResponseWriter, r *http.Request) {
	ng := []struct {
		Nodes []string
		Group string
	}{}

	de := json.NewDecoder(r.Body)
	err := de.Decode(&ng)
	if err != nil {
		outJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	gresp, err := etcdClient.Get(context.TODO(), conf.Config.Proc, clientv3.WithPrefix(), clientv3.WithKeysOnly())
	if err != nil {
		log.Errorf("get nodes list failed: %s", err.Error())
		outJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
	var nodes map[string]bool
	for i := range gresp.Kvs {
		ip := strings.TrimLeft(string(gresp.Kvs[i].Key), conf.Config.Proc)
		nodes[ip] = true
	}

	var errMsg string
	var status int
NGLOOP:
	for i := range ng {
		g := strings.TrimSpace(ng[i].Group)
		if len(g) == 0 {
			errMsg = "group name is emtpy."
			status = http.StatusBadRequest
			break
		}

		for _, n := range ng[i].Nodes {
			n = strings.TrimSpace(n)
			if len(n) == 0 {
				errMsg = fmt.Sprintf("[%s] node ip is emtpy.", g)
				status = http.StatusBadRequest
				break NGLOOP
			}

			if _, ok := nodes[n]; !ok {
				errMsg = fmt.Sprintf("node[%s] not found.", n)
				status = http.StatusBadRequest
				break NGLOOP
			}

			_, err = etcdClient.Put(context.TODO(), path.Join(conf.Config.NodeGroup, g, n), "")
			if err != nil {
				errMsg = "join failed: " + err.Error()
				status = http.StatusInternalServerError
				break NGLOOP
			}
		}
	}

	if len(errMsg) > 0 {
		outJSONError(w, status, errMsg)
		return
	}

	outJSON(w, nil)
}

func nodeLeaveGroup(w http.ResponseWriter, r *http.Request) {}
