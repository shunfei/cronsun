package web

import "github.com/shunfei/cronsun/conf"

type Configuration struct{}

func (cnf *Configuration) Configuratios(ctx *Context) {
	r := struct {
		Security          *conf.Security `json:"security"`
		Alarm             bool           `json:"alarm"`
		LogExpirationDays int            `json:"log_expiration_days"`
	}{
		Security: conf.Config.Security,
		Alarm:    conf.Config.Mail.Enable,
	}

	if conf.Config.Web.LogCleaner.EveryMinute > 0 {
		r.LogExpirationDays = conf.Config.Web.LogCleaner.ExpirationDays
	}

	outJSON(ctx.W, r)
}
