package web

import (
	"time"

	v3 "github.com/coreos/etcd/clientv3"

	"github.com/shunfei/cronsun"
	"github.com/shunfei/cronsun/conf"
)

type Info struct{}

func (inf *Info) Overview(ctx *Context) {
	var overview = struct {
		TotalJobs        int64                   `json:"totalJobs"`
		JobExecuted      *cronsun.StatExecuted   `json:"jobExecuted"`
		JobExecutedDaily []*cronsun.StatExecuted `json:"jobExecutedDaily"`
	}{}

	const day = 24 * time.Hour
	days := 7

	overview.JobExecuted, _ = cronsun.JobLogStat()
	end := time.Now()
	begin := end.Add(time.Duration(1-days) * day)
	statList, _ := cronsun.JobLogDailyStat(begin, end)
	list := make([]*cronsun.StatExecuted, days)
	cur := begin

	for i := 0; i < days; i++ {
		date := cur.Format("2006-01-02")
		var se *cronsun.StatExecuted

		for j := range statList {
			if statList[j].Date == date {
				se = statList[j]
				statList = statList[1:]
				break
			}
		}

		if se != nil {
			list[i] = se
		} else {
			list[i] = &cronsun.StatExecuted{Date: date}
		}

		cur = cur.Add(day)
	}

	overview.JobExecutedDaily = list
	gresp, err := cronsun.DefalutClient.Get(conf.Config.Cmd, v3.WithPrefix(), v3.WithCountOnly())
	if err == nil {
		overview.TotalJobs = gresp.Count
	}

	outJSON(ctx.W, overview)
}
