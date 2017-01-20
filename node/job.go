package node

import (
	"sunteng/cronsun/models"
)

type Job map[string]*models.Job

func newJob(id string, g Group) (j Job, err error) {
	jobs, err := models.GetJobs()
	if err != nil {
		return
	}

	j = make(Job, len(jobs))
	for _, job := range jobs {
		if sch, _ := job.Schedule(id, g, false); len(sch) > 0 {
			j[job.GetID()] = job
		}
	}
	return
}
