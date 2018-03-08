package cronsun

import (
	"time"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"github.com/shunfei/cronsun/conf"
	"github.com/shunfei/cronsun/log"
)

const (
	Coll_JobLog       = "job_log"
	Coll_JobLatestLog = "job_latest_log"
	Coll_Stat         = "stat"
)

// 任务执行记录
type JobLog struct {
	Id        bson.ObjectId `bson:"_id,omitempty" json:"id"`
	JobId     string        `bson:"jobId" json:"jobId"`               // 任务 Id，索引
	JobGroup  string        `bson:"jobGroup" json:"jobGroup"`         // 任务分组，配合 Id 跳转用
	User      string        `bson:"user" json:"user"`                 // 执行此次任务的用户
	Name      string        `bson:"name" json:"name"`                 // 任务名称
	Node      string        `bson:"node" json:"node"`                 // 运行此次任务的节点 id，索引
	Hostname  string        `bson:"hostname" json:"hostname"`         // 运行此次任务的节点主机名称，索引
	Command   string        `bson:"command" json:"command,omitempty"` // 执行的命令，包括参数
	Output    string        `bson:"output" json:"output,omitempty"`   // 任务输出的所有内容
	Success   bool          `bson:"success" json:"success"`           // 是否执行成功
	BeginTime time.Time     `bson:"beginTime" json:"beginTime"`       // 任务开始执行时间，精确到毫秒，索引
	EndTime   time.Time     `bson:"endTime" json:"endTime"`           // 任务执行完毕时间，精确到毫秒
	Cleanup   time.Time     `bson:"cleanup,omitempty" json:"-"`       // 日志清除时间标志
}

type JobLatestLog struct {
	JobLog   `bson:",inline"`
	RefLogId string `bson:"refLogId,omitempty" json:"refLogId"`
}

func GetJobLogById(id bson.ObjectId) (l *JobLog, err error) {
	err = mgoDB.FindId(Coll_JobLog, id, &l)
	return
}

var selectForJobLogList = bson.M{"command": 0, "output": 0}

func GetJobLogList(query bson.M, page, size int, sort string) (list []*JobLog, total int, err error) {
	err = mgoDB.WithC(Coll_JobLog, func(c *mgo.Collection) error {
		total, err = c.Find(query).Count()
		if err != nil {
			return err
		}
		return c.Find(query).Select(selectForJobLogList).Sort(sort).Skip((page - 1) * size).Limit(size).All(&list)
	})
	return
}

func GetJobLatestLogList(query bson.M, page, size int, sort string) (list []*JobLatestLog, total int, err error) {
	err = mgoDB.WithC(Coll_JobLatestLog, func(c *mgo.Collection) error {
		total, err = c.Find(query).Count()
		if err != nil {
			return err
		}
		return c.Find(query).Select(selectForJobLogList).Sort(sort).Skip((page - 1) * size).Limit(size).All(&list)
	})
	return
}

func GetJobLatestLogListByJobIds(jobIds []string) (m map[string]*JobLatestLog, err error) {
	var list []*JobLatestLog

	err = mgoDB.WithC(Coll_JobLatestLog, func(c *mgo.Collection) error {
		return c.Find(bson.M{"jobId": bson.M{"$in": jobIds}}).Select(selectForJobLogList).Sort("beginTime").All(&list)
	})
	if err != nil {
		return
	}

	m = make(map[string]*JobLatestLog, len(list))
	for i := range list {
		m[list[i].JobId] = list[i]
	}
	return
}

func CreateJobLog(j *Job, t time.Time, rs string, success bool) {
	et := time.Now()
	j.Avg(t, et)

	jl := JobLog{
		Id:    bson.NewObjectId(),
		JobId: j.ID,

		JobGroup: j.Group,
		Name:     j.Name,
		User:     j.User,

		Node:     j.runOn,
		Hostname: j.hostname,

		Command: j.Command,
		Output:  rs,
		Success: success,

		BeginTime: t,
		EndTime:   et,
	}

	if conf.Config.Web.LogCleaner.EveryMinute > 0 {
		var expiration int
		if j.LogExpiration > 0 {
			expiration = j.LogExpiration
		} else {
			expiration = conf.Config.Web.LogCleaner.ExpirationDays
		}
		jl.Cleanup = jl.EndTime.Add(time.Duration(expiration) * time.Hour * 24)
	}

	if err := mgoDB.Insert(Coll_JobLog, jl); err != nil {
		log.Errorf(err.Error())
	}

	latestLog := &JobLatestLog{
		RefLogId: jl.Id.Hex(),
		JobLog:   jl,
	}
	latestLog.Id = ""
	if err := mgoDB.Upsert(Coll_JobLatestLog, bson.M{"node": jl.Node, "hostname": jl.Hostname, "jobId": jl.JobId, "jobGroup": jl.JobGroup}, latestLog); err != nil {
		log.Errorf(err.Error())
	}

	var inc = bson.M{"total": 1}
	if jl.Success {
		inc["successed"] = 1
	} else {
		inc["failed"] = 1
	}

	err := mgoDB.Upsert(Coll_Stat, bson.M{"name": "job-day", "date": time.Now().Format("2006-01-02")}, bson.M{"$inc": inc})
	if err != nil {
		log.Errorf("increase stat.job %s", err.Error())
	}
	err = mgoDB.Upsert(Coll_Stat, bson.M{"name": "job"}, bson.M{"$inc": inc})
	if err != nil {
		log.Errorf("increase stat.job %s", err.Error())
	}
}

type StatExecuted struct {
	Total     int64  `bson:"total" json:"total"`
	Successed int64  `bson:"successed" json:"successed"`
	Failed    int64  `bson:"failed" json:"failed"`
	Date      string `bson:"date" json:"date"`
}

func JobLogStat() (s *StatExecuted, err error) {
	err = mgoDB.FindOne(Coll_Stat, bson.M{"name": "job"}, &s)
	return
}

func JobLogDailyStat(begin, end time.Time) (ls []*StatExecuted, err error) {
	const oneDay = time.Hour * 24
	err = mgoDB.WithC(Coll_Stat, func(c *mgo.Collection) error {
		dateList := make([]string, 0, 8)

		cur := begin
		for {
			dateList = append(dateList, cur.Format("2006-01-02"))
			cur = cur.Add(oneDay)
			if cur.After(end) {
				break
			}
		}
		return c.Find(bson.M{"name": "job-day", "date": bson.M{"$in": dateList}}).Sort("date").All(&ls)
	})

	return
}
