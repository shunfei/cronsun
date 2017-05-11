package mid

import (
	"fmt"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

var (
	field = &Field{
		Id:         "seq",
		Collection: "_id",
	}
)

type Field struct {
	Id         string
	Collection string
}

// 如果不设置，则用默认设置
func SetFieldName(id, collection string) {
	field.Id = id
	field.Collection = collection
}

//使collection 为 name 的 id 自增 1 并返回当前 id 的值
func AutoInc(c *mgo.Collection, name string) (id int, err error) {
	return incr(c, name, 1)
}

//批量申请一段id
func ApplyBatchIds(c *mgo.Collection, name string, amount int) (id int, err error) {
	return incr(c, name, amount)
}

func incr(c *mgo.Collection, name string, step int) (id int, err error) {
	result := make(map[string]interface{})
	change := mgo.Change{
		Update:    bson.M{"$inc": bson.M{field.Id: step}},
		Upsert:    true,
		ReturnNew: true,
	}
	_, err = c.Find(bson.M{field.Collection: name}).Apply(change, result)
	if err != nil {
		return
	}
	id, ok := result[field.Id].(int)
	if ok {
		return
	}
	id64, ok := result[field.Id].(int64)
	if !ok {
		err = fmt.Errorf("%s is ont int or int64", field.Id)
		return
	}
	id = int(id64)
	return
}
