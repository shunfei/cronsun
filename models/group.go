package models

import (
	"encoding/json"

	client "github.com/coreos/etcd/clientv3"

	"sunteng/commons/log"
	"sunteng/cronsun/conf"
)

// 结点类型分组
// 注册到 /cronsun/group/<id>
type Group struct {
	ID   string `json:"-"`
	Name string `json:"name"`

	NodeIDs []string `json:"nids"`
}

func GetGroups() (groups map[string]*Group, err error) {
	resp, err := DefalutClient.Get(conf.Config.Group, client.WithPrefix())
	if err != nil {
		return
	}

	count := len(resp.Kvs)
	if count == 0 {
		return
	}

	groups = make(map[string]*Group, count)
	for _, g := range resp.Kvs {
		group := new(Group)
		if e := json.Unmarshal(g.Value, group); e != nil {
			log.Warnf("group[%s] umarshal err: %s", string(g.Key), e.Error())
			continue
		}
		groups[group.ID] = group
	}
	return
}
