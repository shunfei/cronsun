package cronsun

import (
	"encoding/json"
	"fmt"
	"strings"

	client "github.com/coreos/etcd/clientv3"

	"github.com/shunfei/cronsun/conf"
	"github.com/shunfei/cronsun/log"
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

// GetGroups 获取包含 nid 的 group
// 如果 nid 为空，则获取所有的 group
func GetGroups(nid string) (groups map[string]*Group, err error) {
	resp, err := DefalutClient.Get(conf.Config.Group, client.WithPrefix())
	if err != nil {
		return
	}

	count := len(resp.Kvs)
	groups = make(map[string]*Group, count)
	if count == 0 {
		return
	}

	for _, g := range resp.Kvs {
		group := new(Group)
		if e := json.Unmarshal(g.Value, group); e != nil {
			log.Warnf("group[%s] umarshal err: %s", string(g.Key), e.Error())
			continue
		}
		if len(nid) == 0 || group.Included(nid) {
			groups[group.ID] = group
		}
	}
	return
}

func WatchGroups() client.WatchChan {
	return DefalutClient.Watch(conf.Config.Group, client.WithPrefix(), client.WithPrevKV())
}

func GetGroupFromKv(key, value []byte) (g *Group, err error) {
	g = new(Group)
	if err = json.Unmarshal(value, g); err != nil {
		err = fmt.Errorf("group[%s] umarshal err: %s", string(key), err.Error())
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

func (g *Group) Put(modRev int64) (*client.PutResponse, error) {
	b, err := json.Marshal(g)
	if err != nil {
		return nil, err
	}

	return DefalutClient.PutWithModRev(g.Key(), string(b), modRev)
}

func (g *Group) Check() error {
	g.ID = strings.TrimSpace(g.ID)
	if !IsValidAsKeyPath(g.ID) {
		return ErrIllegalNodeGroupId
	}

	g.Name = strings.TrimSpace(g.Name)
	if len(g.Name) == 0 {
		return ErrEmptyNodeGroupName
	}

	return nil
}

func (g *Group) Included(nid string) bool {
	for i, count := 0, len(g.NodeIDs); i < count; i++ {
		if nid == g.NodeIDs[i] {
			return true
		}
	}

	return false
}
