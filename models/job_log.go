package models

import (
	"time"

	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"sunteng/commons/log"
)

const (
	Coll_JobLog = "job_log"
)

// 任务执行记录
type JobLog struct {
	Id        bson.ObjectId `bson:"_id,omitempty" json:"id"`
	JobId     string        `bson:"jobId" json:"jobId"`               // 任务 Id，索引
	JobGroup  string        `bson:"jobGroup" json:"jobGroup"`         // 任务分组，配合 Id 跳转用
	Name      string        `bson:"name" json:"name"`                 // 任务名称
	Node      string        `bson:"node" json:"node"`                 // 运行此次任务的节点 ip，索引
	Command   string        `bson:"command" json:"command,omitempty"` // 执行的命令，包括参数
	Output    string        `bson:"output" json:"output,omitempty"`   // 任务输出的所有内容
	Success   bool          `bson:"success" json:"success"`           // 是否执行成功
	BeginTime time.Time     `bson:"beginTime" json:"beginTime"`       // 任务开始执行时间，精确到毫秒，索引
	EndTime   time.Time     `bson:"endTime" json:"endTime"`           // 任务执行完毕时间，精确到毫秒
}

func GetJobLogById(id bson.ObjectId) (l *JobLog, err error) {
	err = mgoDB.FindId(Coll_JobLog, id, &l)
	return
}

var projection = bson.M{"command": -1, "output": -1}

func GetJobLogList(query bson.M, page, size int, sort string) (list []*JobLog, total int, err error) {
	err = mgoDB.WithC(Coll_JobLog, func(c *mgo.Collection) error {
		total, err = c.Find(query).Count()
		if err != nil {
			return err
		}
		return c.Find(query).Select(projection).Sort(sort).Skip((page - 1) * size).Limit(size).All(&list)
	})
	return
}

func CreateJobLog(j *Job, t time.Time, rs string, success bool) {
	jl := JobLog{
		Id:    bson.NewObjectId(),
		JobId: j.GetID(),

		JobGroup: j.Group,
		Name:     j.Name,

		Node: j.runOn,

		Command: j.Command,
		Output:  rs,
		Success: success,

		BeginTime: t,
		EndTime:   time.Now(),
	}
	if err := mgoDB.Insert(Coll_JobLog, jl); err != nil {
		log.Error(err.Error())
	}
}
