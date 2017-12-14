package web

import (
	"net/http"
	"path"
	"strings"

	"github.com/gorilla/mux"

	"github.com/shunfei/cronsun"
)

func GetVersion(ctx *Context) {
	outJSON(ctx.W, cronsun.Version)
}

func initRouters() (s *http.Server, err error) {
	jobHandler := &Job{}
	nodeHandler := &Node{}
	jobLogHandler := &JobLog{}
	infoHandler := &Info{}
	configHandler := &Configuration{}
	authHandler := &Authentication{}
	adminHandler := &Administrator{}

	r := mux.NewRouter()
	subrouter := r.PathPrefix("/v1").Subrouter()
	subrouter.Handle("/version", NewBaseHandler(GetVersion)).Methods("GET")

	h := NewBaseHandler(authHandler.GetAuthSession)
	subrouter.Handle("/session", h).Methods("GET")
	h = NewBaseHandler(authHandler.DeleteAuthSession)
	subrouter.Handle("/session", h).Methods("DELETE")

	h = NewBaseHandler(authHandler.SetPassword)
	subrouter.Handle("/user/setpwd", h).Methods("POST")

	h = NewAdminAuthHandler(adminHandler.GetAccount)
	subrouter.Handle("/admin/account/{email}", h).Methods("GET")
	h = NewAdminAuthHandler(adminHandler.GetAccountList)
	subrouter.Handle("/admin/accounts", h).Methods("GET")
	h = NewAdminAuthHandler(adminHandler.AddAccount)
	subrouter.Handle("/admin/account", h).Methods("PUT")
	h = NewAdminAuthHandler(adminHandler.UpdateAccount)
	subrouter.Handle("/admin/account", h).Methods("POSt")

	// get job list
	h = NewAuthHandler(jobHandler.GetList)
	subrouter.Handle("/jobs", h).Methods("GET")
	// get a job group list
	h = NewAuthHandler(jobHandler.GetGroups)
	subrouter.Handle("/job/groups", h).Methods("GET")
	// create/update a job
	h = NewAuthHandler(jobHandler.UpdateJob)
	subrouter.Handle("/job", h).Methods("PUT")
	// pause/start
	h = NewAuthHandler(jobHandler.ChangeJobStatus)
	subrouter.Handle("/job/{group}-{id}", h).Methods("POST")
	// batch pause/start
	h = NewAuthHandler(jobHandler.BatchChangeJobStatus)
	subrouter.Handle("/jobs/{op}", h).Methods("POST")
	// get a job
	h = NewAuthHandler(jobHandler.GetJob)
	subrouter.Handle("/job/{group}-{id}", h).Methods("GET")
	// remove a job
	h = NewAuthHandler(jobHandler.DeleteJob)
	subrouter.Handle("/job/{group}-{id}", h).Methods("DELETE")

	h = NewAuthHandler(jobHandler.GetJobNodes)
	subrouter.Handle("/job/{group}-{id}/nodes", h).Methods("GET")

	h = NewAuthHandler(jobHandler.JobExecute)
	subrouter.Handle("/job/{group}-{id}/execute", h).Methods("PUT")

	// query executing job
	h = NewAuthHandler(jobHandler.GetExecutingJob)
	subrouter.Handle("/job/executing", h).Methods("GET")

	// get job log list
	h = NewAuthHandler(jobLogHandler.GetList)
	subrouter.Handle("/logs", h).Methods("GET")
	// get job log
	h = NewAuthHandler(jobLogHandler.GetDetail)
	subrouter.Handle("/log/{id}", h).Methods("GET")

	h = NewAuthHandler(nodeHandler.GetNodes)
	subrouter.Handle("/nodes", h).Methods("GET")
	h = NewAuthHandler(nodeHandler.DeleteNode)
	subrouter.Handle("/node/{ip}", h).Methods("DELETE")
	// get node group list
	h = NewAuthHandler(nodeHandler.GetGroups)
	subrouter.Handle("/node/groups", h).Methods("GET")
	// get a node group by group id
	h = NewAuthHandler(nodeHandler.GetGroupByGroupId)
	subrouter.Handle("/node/group/{id}", h).Methods("GET")
	// create/update a node group
	h = NewAuthHandler(nodeHandler.UpdateGroup)
	subrouter.Handle("/node/group", h).Methods("PUT")
	// delete a node group
	h = NewAuthHandler(nodeHandler.DeleteGroup)
	subrouter.Handle("/node/group/{id}", h).Methods("DELETE")

	h = NewAuthHandler(infoHandler.Overview)
	subrouter.Handle("/info/overview", h).Methods("GET")

	h = NewAuthHandler(configHandler.Configuratios)
	subrouter.Handle("/configurations", h).Methods("GET")

	r.PathPrefix("/ui/").Handler(http.StripPrefix("/ui/", newEmbeddedFileServer("", "index.html")))
	r.NotFoundHandler = NewBaseHandler(notFoundHandler)

	s = &http.Server{
		Handler: r,
	}
	return s, nil
}

type embeddedFileServer struct {
	Prefix    string
	IndexFile string
}

func newEmbeddedFileServer(prefix, index string) *embeddedFileServer {
	index = strings.TrimLeft(index, "/")
	return &embeddedFileServer{Prefix: prefix, IndexFile: index}
}

func (s *embeddedFileServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fp := path.Clean(s.Prefix + r.URL.Path)
	if fp == "." {
		fp = ""
	} else {
		fp = strings.TrimLeft(fp, "/")
	}

	b, err := Asset(fp)
	if err == nil {
		w.Write(b)
		return
	}

	if len(fp) > 0 {
		fp += "/"
	}
	fp += s.IndexFile

	b, err = Asset(fp)
	if err == nil {
		w.Write(b)
		return
	}

	_notFoundHandler(w, r)
}

func notFoundHandler(c *Context) {
	_notFoundHandler(c.W, c.R)
}

func _notFoundHandler(w http.ResponseWriter, r *http.Request) {
	const html = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>404 page not found</title>
</head>
<body>
    The page you are looking for is not found. Redirect to <a href="/ui/">Dashboard</a> after <span id="s">5</span> seconds.
</body>
<script type="text/javascript">
var s = 5;
setInterval(function(){
    s--;
    document.getElementById('s').innerText = s;
    if (s === 0) location.href = '/ui/';
}, 1000);
</script>
</html>`

	if strings.HasPrefix(strings.TrimLeft(r.URL.Path, "/"), "v1") {
		outJSONWithCode(w, http.StatusNotFound, "Api not found.")
	} else {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(html))
	}
}
