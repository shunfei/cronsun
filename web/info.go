package web

import (
	"net/http"
	"time"

	v3 "github.com/coreos/etcd/clientv3"

	"github.com/shunfei/cronsun/conf"
	"github.com/shunfei/cronsun"
)

type Info struct{}

func (inf *Info) Overview(w http.ResponseWriter, r *http.Request) {
	var overview = struct {
		TotalJobs        int64                `json:"totalJobs"`
		JobExecuted      *cronsun.StatExecuted `json:"jobExecuted"`
		JobExecutedDaily *cronsun.StatExecuted `json:"jobExecutedDaily"`
	}{}

	overview.JobExecuted, _ = cronsun.JobLogStat()
	overview.JobExecutedDaily, _ = cronsun.JobLogDayStat(time.Now())

	gresp, err := cronsun.DefalutClient.Get(conf.Config.Cmd, v3.WithPrefix(), v3.WithCountOnly())
	if err == nil {
		overview.TotalJobs = gresp.Count
	}

	outJSON(w, overview)
}
