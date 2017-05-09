package web

import (
	"net/http"

	"github.com/shunfei/cronsun/conf"
)

type Configuration struct{}

func (cnf *Configuration) Configuratios(w http.ResponseWriter, r *http.Request) {
	outJSON(w, struct {
		Security *conf.Security `json:"security"`
		Alarm    bool           `json:"alarm"`
	}{
		Security: conf.Config.Security,
		Alarm:    conf.Config.Mail.Enable,
	})
}
