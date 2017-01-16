package models

import (
	"encoding/json"
	"strings"

	client "github.com/coreos/etcd/clientv3"

	"sunteng/commons/log"
	"sunteng/cronsun/conf"
)

// 结点类型分组
// 注册到 /cronsun/group/<id>
type Group struct {
	ID   string `json:"id"`
	Name string `json:"name"`

	NodeIDs []string `json:"nids"`
}

func GetGroupById(gid string) (g *Group, err error) {
	if len(gid) == 0 {
		return
	}
	resp, err := DefalutClient.Get(conf.Config.Group + gid)
	if err != nil || resp.Count == 0 {
		return
	}

	err = json.Unmarshal(resp.Kvs[0].Value, &g)
	return
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

func DeleteGroupById(id string) (*client.DeleteResponse, error) {
	return DefalutClient.Delete(GroupKey(id))
}

func GroupKey(id string) string {
	return conf.Config.Group + id
}

func (g *Group) Key() string {
	return GroupKey(g.ID)
}

func (g *Group) Put(rev int64) (*client.PutResponse, error) {
	b, err := json.Marshal(g)
	if err != nil {
		return nil, err
	}

	return DefalutClient.PutWithRev(g.Key(), string(b), rev)
}

func (g *Group) Check() error {
	g.Name = strings.TrimSpace(g.Name)
	if len(g.Name) == 0 {
		return ErrEmptyNodeGroupName
	}

	return nil
}
