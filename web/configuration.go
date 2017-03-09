package web

import (
	"net/http"

	"sunteng/cronsun/conf"
)

type Configuration struct {
	Security *securityCnf `json:"security"`
}

type securityCnf struct {
	Enable       bool     `json:"enable"`
	AllowUsers   []string `json:"allowUsers,omitempty"`
	AllowSuffixs []string `json:"allowSuffixs,omitempty"`
}

func NewConfiguration() *Configuration {
	cnf := &Configuration{
		Security: &securityCnf{
			Enable: conf.Config.Security.Open,
		},
	}

	if conf.Config.Security.Open {
		cnf.Security.AllowUsers = conf.Config.Security.Users
		cnf.Security.AllowSuffixs = conf.Config.Security.Ext
	}

	return cnf
}

func (cnf *Configuration) Configuratios(w http.ResponseWriter, r *http.Request) {
	outJSON(w, cnf)
}
