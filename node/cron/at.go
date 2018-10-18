package cron

import (
	"sort"
	"time"
)

// TimeListSchedule will run at the specify giving time.
type TimeListSchedule struct {
	timeList []time.Time
}

// At returns a crontab Schedule that activates every specify time.
func At(tl []time.Time) *TimeListSchedule {
	sort.Slice(tl, func(i, j int) bool { return tl[i].Unix() < tl[j].Unix() })
	return &TimeListSchedule{
		timeList: tl,
	}
}

// Next returns the next time this should be run.
// This rounds so that the next activation time will be on the second.
func (schedule *TimeListSchedule) Next(t time.Time) time.Time {
	cur := 0
	for cur < len(schedule.timeList) {
		nextt := schedule.timeList[cur]
		cur++
		if nextt.UnixNano() <= t.UnixNano() {
			continue
		}
		return nextt
	}
	return time.Time{}
}
