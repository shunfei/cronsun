package web

import (
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"math"
	"github.com/shunfei/cronsun"
)

type JobLog struct{}

func (jl *JobLog) GetDetail(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := strings.TrimSpace(vars["id"])
	if len(id) == 0 {
		outJSONWithCode(w, http.StatusBadRequest, "empty log id.")
		return
	}

	if !bson.IsObjectIdHex(id) {
		outJSONWithCode(w, http.StatusBadRequest, "invalid ObjectId.")
		return
	}

	logDetail, err := cronsun.GetJobLogById(bson.ObjectIdHex(id))
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err == mgo.ErrNotFound {
			statusCode = http.StatusNotFound
			err = nil
		}
		outJSONWithCode(w, statusCode, err)
		return
	}

	outJSON(w, logDetail)
}

func (jl *JobLog) GetList(w http.ResponseWriter, r *http.Request) {
	nodes := getStringArrayFromQuery("nodes", ",", r)
	names := getStringArrayFromQuery("names", ",", r)
	ids := getStringArrayFromQuery("ids", ",", r)
	begin := getTime(r.FormValue("begin"))
	end := getTime(r.FormValue("end"))
	page := getPage(r.FormValue("page"))
	failedOnly := r.FormValue("failedOnly") == "true"
	pageSize := getPageSize(r.FormValue("pageSize"))
	sort := "-beginTime"
	if r.FormValue("sort") == "1" {
		sort = "beginTime"
	}

	query := bson.M{}
	if len(nodes) > 0 {
		query["node"] = bson.M{"$in": nodes}
	}

	if len(ids) > 0 {
		query["jobId"] = bson.M{"$in": ids}
	}

	if len(names) > 0 {
		var search []bson.M
		for _, k := range names {
			k = strings.TrimSpace(k)
			if len(k) == 0 {
				continue
			}
			search = append(search, bson.M{"name": bson.M{"$regex": bson.RegEx{Pattern: k, Options: "i"}}})
		}
		query["$or"] = search
	}

	if !begin.IsZero() {
		query["beginTime"] = bson.M{"$gte": begin}
	}
	if !end.IsZero() {
		query["endTime"] = bson.M{"$lt": end.Add(time.Hour * 24)}
	}

	if failedOnly {
		query["success"] = false
	}

	var pager struct {
		Total int              `json:"total"`
		List  []*cronsun.JobLog `json:"list"`
	}
	var err error
	if r.FormValue("latest") == "true" {
		var latestLogList []*cronsun.JobLatestLog
		latestLogList, pager.Total, err = cronsun.GetJobLatestLogList(query, page, pageSize, sort)
		for i := range latestLogList {
			latestLogList[i].JobLog.Id = bson.ObjectIdHex(latestLogList[i].RefLogId)
			pager.List = append(pager.List, &latestLogList[i].JobLog)
		}
	} else {
		pager.List, pager.Total, err = cronsun.GetJobLogList(query, page, pageSize, sort)
	}
	if err != nil {
		outJSONWithCode(w, http.StatusInternalServerError, err.Error())
		return
	}

	pager.Total = int(math.Ceil(float64(pager.Total) / float64(pageSize)))
	outJSON(w, pager)
}
