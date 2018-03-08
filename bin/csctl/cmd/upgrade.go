package cmd

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/coreos/etcd/clientv3"
	"github.com/spf13/cobra"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"github.com/shunfei/cronsun"
	"github.com/shunfei/cronsun/conf"
)

var prever string

func init() {
	UpgradeCmd.Flags().StringVarP(&prever, "prever", "p", "", "previous version of cronsun you are used")
}

var UpgradeCmd = &cobra.Command{
	Use:   "upgrade",
	Short: "upgrade will upgrade data to the current version(" + cronsun.VersionNumber + ")",
	Run: func(cmd *cobra.Command, args []string) {
		var ea = NewExitAction()

		prever = strings.TrimLeft(strings.TrimSpace(prever), "v")
		if len(prever) < 5 {
			ea.Exit("invalid version number")
		}

		if prever < "0.3.0" {
			fmt.Println("upgrading data to version 0.3.0")
			nodesById := getIPMapper(ea)
			if to_0_3_0(ea, nodesById) {
				return
			}
		}
	},
}

func getIPMapper(ea *ExitAction) map[string]*cronsun.Node {
	nodes, err := cronsun.GetNodes()
	if err != nil {
		ea.Exit("failed to fetch nodes from MongoDB: %s", err.Error())
	}

	var ipMapper = make(map[string]*cronsun.Node, len(nodes))
	for _, n := range nodes {
		n.IP = strings.TrimSpace(n.IP)
		if n.IP == "" || n.ID == "" {
			continue
		}

		ipMapper[n.IP] = n
	}

	return ipMapper
}

// to_0_3_0 can be run many times
func to_0_3_0(ea *ExitAction, nodesById map[string]*cronsun.Node) (shouldStop bool) {
	var replaceIDs = func(list []string) {
		for i := range list {
			if node, ok := nodesById[list[i]]; ok {
				list[i] = node.ID
			}
		}
	}

	// update job data
	gresp, err := cronsun.DefalutClient.Get(conf.Config.Cmd, clientv3.WithPrefix())
	ea.ExitOnErr(err)

	total := len(gresp.Kvs)
	upgraded := 0
	for i := range gresp.Kvs {
		job := cronsun.Job{}
		err = json.Unmarshal(gresp.Kvs[i].Value, &job)
		if err != nil {
			fmt.Printf("[Error] failed to decode job(%s) data: %s\n", string(gresp.Kvs[i].Key), err.Error())
			continue
		}

		for _, rule := range job.Rules {
			replaceIDs(rule.ExcludeNodeIDs)
			replaceIDs(rule.NodeIDs)
		}

		d, err := json.Marshal(&job)
		if err != nil {
			fmt.Printf("[Error] failed to encode job(%s) data: %s\n", string(gresp.Kvs[i].Key), err.Error())
			continue
		}

		_, err = cronsun.DefalutClient.Put(job.Key(), string(d))
		if err != nil {
			fmt.Printf("[Warn] failed to restore job(%s) data: %s\n", string(gresp.Kvs[i].Key), err.Error())
			continue
		}
		upgraded++
	}
	if total != upgraded {
		shouldStop = true
	}
	fmt.Printf("%d of %d jobs has been upgraded.\n", upgraded, total)

	// migrate node group data
	nodeGroups, err := cronsun.GetNodeGroups()
	if err != nil {
		ea.Exit("[Error] failed to get node group datas: ", err.Error())
	}

	total = len(nodeGroups)
	upgraded = 0
	for i := range nodeGroups {
		replaceIDs(nodeGroups[i].NodeIDs)
		if _, err = nodeGroups[i].Put(0); err != nil {
			fmt.Printf("[Warn] failed to restore node group(id: %s, name: %s) data: %s\n", nodeGroups[i].ID, nodeGroups[i].Name, err.Error())
			continue
		}
		upgraded++
	}
	if total != upgraded {
		shouldStop = true
	}
	fmt.Printf("%d of %d node group has been upgraded.\n", upgraded, total)

	// upgrade logs
	cronsun.GetDb().WithC(cronsun.Coll_JobLog, func(c *mgo.Collection) error {
		for ip, node := range nodesById {
			_, err = c.UpdateAll(bson.M{"node": ip}, bson.M{"$set": bson.M{"node": node.ID, "hostname": node.Hostname}})
			if err != nil {
				if err != nil {
					fmt.Println("failed to upgrade job logs: ", err.Error())
				}
				break
			}
		}
		shouldStop = true
		return err
	})

	// upgrade logs
	cronsun.GetDb().WithC(cronsun.Coll_JobLatestLog, func(c *mgo.Collection) error {
		for ip, node := range nodesById {
			_, err = c.UpdateAll(bson.M{"node": ip}, bson.M{"$set": bson.M{"node": node.ID, "hostname": node.Hostname}})
			if err != nil {
				if err != nil {
					fmt.Println("failed to upgrade job latest logs: ", err.Error())
				}
				break
			}
		}
		shouldStop = true
		return err
	})

	return
}
