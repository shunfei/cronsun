package web

import (
	"net/http"

	"sunteng/cronsun/conf"
)

type Configuration struct{}

func (cnf *Configuration) Configuratios(w http.ResponseWriter, r *http.Request) {
	outJSON(w, struct {
		Security *conf.Security `json:"security"`
	}{
		Security: conf.Config.Security,
	})
}
