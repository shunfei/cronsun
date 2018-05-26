package node

import (
	"encoding/json"

	"gopkg.in/mgo.v2/bson"

	"github.com/shunfei/cronsun"
	"github.com/shunfei/cronsun/conf"
	"github.com/shunfei/cronsun/log"
)

func (n *Node) executCsctlCmd(key, value []byte) error {
	cmd := &cronsun.CsctlCmd{}
	err := json.Unmarshal(value, cmd)
	if err != nil {
		log.Warnf("invalid csctl command[%s] value[%s], err: %s", string(key), string(value), err.Error())
		return err
	}

	if cronsun.NodeCmdMax <= cmd.Cmd || cmd.Cmd <= cronsun.NodeCmdUnknown {
		log.Warnf("invalid csctl command[%s] value[%s], err: %s", string(key), string(value))
		return cronsun.InvalidNodeCmdErr
	}

	switch cmd.Cmd {
	case cronsun.NodeCmdRmOld:
		n.rmOld()
	}

	log.Infof("%s execute csctl command[%s] success", n.String(), cmd.Cmd.String())
	return nil
}

func (n *Node) rmOld() {
	// remove old version(< 0.3.0) node info
	cronsun.RemoveNode(bson.M{"_id": n.IP})
	cronsun.DefalutClient.Delete(conf.Config.Node + n.IP)
}
