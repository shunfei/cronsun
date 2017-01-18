package web

import (
	"net/http"
	"path"

	"github.com/gorilla/mux"

	"sunteng/cronsun/conf"
)

func InitRouters() (s *http.Server, err error) {
	jobHandler := &Job{}
	nodeHandler := &Node{}

	r := mux.NewRouter()
	subrouter := r.PathPrefix("/v1").Subrouter()

	// get a job group list
	h := BaseHandler{Handle: jobHandler.GetGroups}
	subrouter.Handle("/job/groups", h).Methods("GET")
	// get a job group by group name
	h = BaseHandler{Handle: jobHandler.GetListByGroupName}
	subrouter.Handle("/job/group/{name}", h).Methods("GET")
	// create/update a job
	h = BaseHandler{Handle: jobHandler.UpdateJob}
	subrouter.Handle("/job", h).Methods("PUT")
	// get a job
	h = BaseHandler{Handle: jobHandler.GetJob}
	subrouter.Handle("/job/{group}-{id}", h).Methods("GET")
	// remove a job
	h = BaseHandler{Handle: jobHandler.DeleteJob}
	subrouter.Handle("/job/{group}-{id}", h).Methods("DELETE")

	h = BaseHandler{Handle: nodeHandler.GetActivityNodeList}
	subrouter.Handle("/node/activitys", h).Methods("GET")
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
	subrouter.Handle("/node/group", h).Methods("DELETE")

	uidir := conf.Config.Web.UIDir
	if len(uidir) == 0 {
		uidir = path.Join(conf.Config.Root, "web", "ui")
	}
	r.PathPrefix("/ui/").Handler(http.StripPrefix("/ui/", http.FileServer(http.Dir(uidir))))

	s = &http.Server{
		Handler: r,
	}
	return s, nil
}
