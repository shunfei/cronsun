package web

import (
	"net/http"

	"sunteng/cronsun/conf"
)

type Configuration struct {
	Security *conf.Security `json:"security"`
}

func NewConfiguration() *Configuration {
	cnf := &Configuration{
		Security: conf.Config.Security,
	}

	return cnf
}

func (cnf *Configuration) Configuratios(w http.ResponseWriter, r *http.Request) {
	outJSON(w, cnf)
}
