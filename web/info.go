package web

import (
	"net/http"
	"time"

	v3 "github.com/coreos/etcd/clientv3"

	"github.com/shunfei/cronsun/conf"
	"github.com/shunfei/cronsun/models"
)

type Info struct{}

func (inf *Info) Overview(w http.ResponseWriter, r *http.Request) {
	var overview = struct {
		TotalJobs        int64                `json:"totalJobs"`
		JobExecuted      *models.StatExecuted `json:"jobExecuted"`
		JobExecutedDaily *models.StatExecuted `json:"jobExecutedDaily"`
	}{}

	overview.JobExecuted, _ = models.JobLogStat()
	overview.JobExecutedDaily, _ = models.JobLogDayStat(time.Now())

	gresp, err := models.DefalutClient.Get(conf.Config.Cmd, v3.WithPrefix(), v3.WithCountOnly())
	if err == nil {
		overview.TotalJobs = gresp.Count
	}

	outJSON(w, overview)
}
