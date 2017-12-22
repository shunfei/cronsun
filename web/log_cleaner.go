package web

import (
	"time"

	"github.com/shunfei/cronsun"
	"github.com/shunfei/cronsun/log"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

func RunLogCleaner(cleanPeriod, expiration time.Duration) (close chan struct{}) {
	t := time.NewTicker(cleanPeriod)
	close = make(chan struct{})
	go func() {
		for {
			select {
			case <-t.C:
				cleanupLogs(expiration)
			case <-close:
				return
			}
		}
	}()

	return
}

func cleanupLogs(expiration time.Duration) {
	err := cronsun.GetDb().WithC(cronsun.Coll_JobLog, func(c *mgo.Collection) error {
		_, err := c.RemoveAll(bson.M{"$or": []bson.M{
			bson.M{"$and": []bson.M{
				bson.M{"cleanup": bson.M{"$exists": true}},
				bson.M{"cleanup": bson.M{"$lte": time.Now()}},
			}},
			bson.M{"$and": []bson.M{
				bson.M{"cleanup": bson.M{"$exists": false}},
				bson.M{"endTime": bson.M{"$lte": time.Now().Add(-expiration)}},
			}},
		}})

		return err
	})

	if err != nil {
		log.Errorf("[Cleaner] Failed to remove expired logs: %s", err.Error())
		return
	}

}
