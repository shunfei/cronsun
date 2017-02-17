package node

import (
	"sunteng/cronsun/models"
)

type Jobs map[string]*models.Job

func loadJobs(id string, g Groups) (j Jobs, err error) {
	jobs, err := models.GetJobs()
	if err != nil {
		return
	}

	j = make(Jobs, len(jobs))
	for _, job := range jobs {
		if sch, _ := job.Schedule(id, g, false); len(sch) > 0 {
			j[job.GetID()] = job
		}
	}
	return
}
