package node

import (
	"encoding/json"

	"github.com/shunfei/cronsun"
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
		n.Node.RmOldInfo()
	case cronsun.NodeCmdSync:
		n.Node.SyncToMgo()
	}

	log.Infof("%s execute csctl command[%s] success", n.String(), cmd.Cmd.String())
	return nil
}
