package web

import (
	"net/http"
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
	list := []*models.JobLog{
		&models.JobLog{
			Id:        bson.NewObjectId(),
			Name:      "test1",
			JobId:     "test1",
			Node:      "192.168.1.2",
			ExitCode:  0,
			BeginTime: time.Now(),
			EndTime:   time.Now().Add(time.Duration(time.Minute)),
		},
		&models.JobLog{
			Id:        bson.NewObjectId(),
			Name:      "test2",
			JobId:     "test2",
			Node:      "192.168.1.2",
			ExitCode:  1,
			BeginTime: time.Now(),
			EndTime:   time.Now().Add(time.Duration(350 * time.Millisecond)),
		},
		&models.JobLog{
			Id:        bson.NewObjectId(),
			Name:      "test3",
			JobId:     "test3",
			Node:      "192.168.1.3",
			ExitCode:  0,
			BeginTime: time.Now(),
			EndTime:   time.Now().Add(time.Duration(time.Hour * 12)),
		},
	}
	outJSON(w, list)
}
