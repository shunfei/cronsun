package web

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"sunteng/cronsun/models"
)

type JobLog struct{}

func (jl *JobLog) GetDetail(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := strings.TrimSpace(vars["id"])
	if len(id) == 0 {
		outJSONWithCode(w, http.StatusBadRequest, "empty log id.")
		return
	}

	logDetail, err := models.GetJobLogById(bson.ObjectIdHex(id))
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
	nodes := GetStringArrayFromQuery("nodes", ",", r)
	names := GetStringArrayFromQuery("names", ",", r)
	begin := getTime(r.FormValue("begin"))
	end := getTime(r.FormValue("end"))
	page := getPage(r.FormValue("page"))
	pageSize := getPageSize(r.FormValue("pageSize"))
	sort := "-beginTime"
	if r.FormValue("sort") == "1" {
		sort = "beginTime"
	}

	query := bson.M{}
	if len(nodes) > 0 {
		query["node"] = bson.M{"$in": nodes}
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

	var pager struct {
		Total int              `json:"total"`
		List  []*models.JobLog `json:"list"`
	}
	var err error
	pager.List, pager.Total, err = models.GetJobLogList(query, page, pageSize, sort)
	if err != nil {
		outJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	pager.Total /= pageSize
	outJSON(w, pager)
}

func GetStringArrayFromQuery(name, sep string, r *http.Request) (arr []string) {
	val := strings.TrimSpace(r.FormValue(name))
	if len(val) == 0 {
		return
	}

	return strings.Split(val, sep)
}

func getPage(page string) int {
	p, err := strconv.Atoi(page)
	if err != nil || p < 1 {
		p = 1
	}

	return p
}

func getPageSize(ps string) int {
	p, err := strconv.Atoi(ps)
	if err != nil || p < 1 {
		p = 50
	} else if p > 200 {
		p = 200
	}
	return p
}

func getTime(t string) time.Time {
	t = strings.TrimSpace(t)
	time, _ := time.Parse("2006-01-02", t)
	return time
}
