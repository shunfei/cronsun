package web

import "github.com/shunfei/cronsun/conf"

type Configuration struct{}

func (cnf *Configuration) Configuratios(ctx *Context) {
	outJSON(ctx.W, struct {
		Security *conf.Security `json:"security"`
		Alarm    bool           `json:"alarm"`
	}{
		Security: conf.Config.Security,
		Alarm:    conf.Config.Mail.Enable,
	})
}
