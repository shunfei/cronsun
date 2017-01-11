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

	h := BaseHandler{Handle: jobHandler.GetGroups}
	subrouter.Handle("/job/groups", h).Methods("GET")

	h = BaseHandler{Handle: jobHandler.GetListByGroupName}
	subrouter.Handle("/job/group/{name}", h).Methods("GET")

	h = BaseHandler{Handle: jobHandler.Update}
	subrouter.Handle("/job", h).Methods("PUT")

	h = BaseHandler{Handle: nodeHandler.GetGroups}
	subrouter.Handle("/node/groups", h).Methods("GET")

	h = BaseHandler{Handle: nodeHandler.GetGroupByGroupName}
	subrouter.Handle("/node/group/{name}", h).Methods("GET")

	h = BaseHandler{Handle: nodeHandler.JoinGroup}
	subrouter.Handle("/node/group", h).Methods("PUT")

	h = BaseHandler{Handle: nodeHandler.LeaveGroup}
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
