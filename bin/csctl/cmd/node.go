package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/shunfei/cronsun"
)

var (
	nodeCmd     string
	nodeInclude string
	nodeExclude string

	spliter = ","
)

func init() {
	NodeCmd.Flags().StringVar(&nodeCmd, "cmd", "", "the command send to node")
	NodeCmd.Flags().StringVar(&nodeInclude, "include", "", "the node ids that needs to execute the command, split by ',', e.g: '--include=aa,bb,cc', empty means all nodes")
	NodeCmd.Flags().StringVar(&nodeExclude, "exclude", "", "the node ids that doesn't need to execute the command, split by ',', e.g: '--exclude=aa,bb,cc', empty means none")
}

var NodeCmd = &cobra.Command{
	Use:   "node",
	Short: "Send some commands to nodes",
	Long: `Send a command to nodes and execute it.

Available Commands:
  rmold:	remove old version(< 0.3.0) node info from mongodb and etcd
  sync:		sync node info to mongodb
`,
	Run: func(cmd *cobra.Command, args []string) {
		ea := NewExitAction()
		ea.After = func() {
			fmt.Println()
			cmd.Help()
		}
		nc, err := cronsun.ToNodeCmd(nodeCmd)
		if err != nil {
			ea.Exit(err.Error() + ": " + nodeCmd)
		}

		var include, exclude []string
		if len(nodeInclude) > 0 {
			include = strings.Split(nodeInclude, spliter)
		}
		if len(nodeExclude) > 0 {
			exclude = strings.Split(nodeExclude, spliter)
		}

		err = cronsun.PutCsctl(&cronsun.CsctlCmd{
			Cmd:     nc,
			Include: include,
			Exclude: exclude,
		})
		if err != nil {
			ea.ExitOnErr(err)
		}

		fmt.Printf("command[%s] send success\n", nodeCmd)
	},
}
