package web

import (
	"net/http"
	"path"

	"github.com/gorilla/mux"

	"sunteng/cronsun/conf"
	"sunteng/cronsun/models"
)

func GetVersion(w http.ResponseWriter, r *http.Request) {
	outJSON(w, models.Version)
}

func InitRouters() (s *http.Server, err error) {
	jobHandler := &Job{}
	nodeHandler := &Node{}
	jobLogHandler := &JobLog{}
	infoHandler := &Info{}
	configHandler := &Configuration{}

	r := mux.NewRouter()
	subrouter := r.PathPrefix("/v1").Subrouter()
	subrouter.Handle("/version", BaseHandler{Handle: GetVersion}).Methods("GET")

	// get job list
	h := BaseHandler{Handle: jobHandler.GetList}
	subrouter.Handle("/jobs", h).Methods("GET")
	// get a job group list
	h = BaseHandler{Handle: jobHandler.GetGroups}
	subrouter.Handle("/job/groups", h).Methods("GET")
	// create/update a job
	h = BaseHandler{Handle: jobHandler.UpdateJob}
	subrouter.Handle("/job", h).Methods("PUT")
	// pause/start
	h = BaseHandler{Handle: jobHandler.ChangeJobStatus}
	subrouter.Handle("/job/{group}-{id}", h).Methods("POST")
	// get a job
	h = BaseHandler{Handle: jobHandler.GetJob}
	subrouter.Handle("/job/{group}-{id}", h).Methods("GET")
	// remove a job
	h = BaseHandler{Handle: jobHandler.DeleteJob}
	subrouter.Handle("/job/{group}-{id}", h).Methods("DELETE")

	h = BaseHandler{Handle: jobHandler.GetJobNodes}
	subrouter.Handle("/job/{group}-{id}/nodes", h).Methods("GET")

	h = BaseHandler{Handle: jobHandler.JobExecute}
	subrouter.Handle("/job/{group}-{id}/execute", h).Methods("PUT")

	// query executing job
	h = BaseHandler{Handle: jobHandler.GetExecutingJob}
	subrouter.Handle("/job/executing", h).Methods("GET")

	// get job log list
	h = BaseHandler{Handle: jobLogHandler.GetList}
	subrouter.Handle("/logs", h).Methods("GET")
	// get job log
	h = BaseHandler{Handle: jobLogHandler.GetDetail}
	subrouter.Handle("/log/{id}", h).Methods("GET")

	h = BaseHandler{Handle: nodeHandler.GetNodes}
	subrouter.Handle("/nodes", h).Methods("GET")
	// get node group list
	h = BaseHandler{Handle: nodeHandler.GetGroups}
	subrouter.Handle("/node/groups", h).Methods("GET")
	// get a node group by group id
	h = BaseHandler{Handle: nodeHandler.GetGroupByGroupId}
	subrouter.Handle("/node/group/{id}", h).Methods("GET")
	// create/update a node group
	h = BaseHandler{Handle: nodeHandler.UpdateGroup}
	subrouter.Handle("/node/group", h).Methods("PUT")
	// delete a node group
	h = BaseHandler{Handle: nodeHandler.DeleteGroup}
	subrouter.Handle("/node/group/{id}", h).Methods("DELETE")

	h = BaseHandler{Handle: infoHandler.Overview}
	subrouter.Handle("/info/overview", h).Methods("GET")

	h = BaseHandler{Handle: configHandler.Configuratios}
	subrouter.Handle("/configurations", h).Methods("GET")

	uidir := conf.Config.Web.UIDir
	if len(uidir) == 0 {
		uidir = path.Join("web", "ui", "dist")
	}
	r.PathPrefix("/ui/").Handler(http.StripPrefix("/ui/", http.FileServer(http.Dir(uidir))))

	s = &http.Server{
		Handler: r,
	}
	return s, nil
}
