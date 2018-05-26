package cronsun

import (
	"encoding/json"
	"errors"

	client "github.com/coreos/etcd/clientv3"

	"github.com/shunfei/cronsun/conf"
)

const (
	NodeCmdUnknown NodeCmd = iota
	NodeCmdRmOld
	NodeCmdSync
	NodeCmdMax
)

var (
	InvalidNodeCmdErr = errors.New("invalid node command")

	NodeCmds = []string{
		"unknown",
		"rmold",
		"sync",
	}
)

type NodeCmd int

func (cmd NodeCmd) String() string {
	if NodeCmdMax <= cmd || cmd <= NodeCmdUnknown {
		return "unknown"
	}
	return NodeCmds[cmd]
}

func ToNodeCmd(cmd string) (NodeCmd, error) {
	for nc := NodeCmdUnknown + 1; nc < NodeCmdMax; nc++ {
		if cmd == NodeCmds[nc] {
			return nc, nil
		}
	}
	return NodeCmdUnknown, InvalidNodeCmdErr
}

type CsctlCmd struct {
	// the command send to node
	Cmd NodeCmd
	// the node ids that needs to execute the command, empty means all node
	Include []string
	// the node ids that doesn't need to execute the command, empty means none
	Exclude []string
}

// 执行 csctl 发送的命令
// 注册到 /cronsun/csctl/<cmd>
func PutCsctl(cmd *CsctlCmd) error {
	if NodeCmdMax <= cmd.Cmd || cmd.Cmd <= NodeCmdUnknown {
		return InvalidNodeCmdErr
	}

	params, err := json.Marshal(cmd)
	if err != nil {
		return err
	}
	_, err = DefalutClient.Put(conf.Config.Csctl+NodeCmds[cmd.Cmd], string(params))
	return err
}

func WatchCsctl() client.WatchChan {
	return DefalutClient.Watch(conf.Config.Csctl, client.WithPrefix())
}
