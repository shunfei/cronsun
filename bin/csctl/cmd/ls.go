package cmd

import (
	"fmt"

	"github.com/shunfei/cronsun"
	"github.com/spf13/cobra"
)

var all bool

func init() {
	LsCmd.Flags().BoolVarP(&all, "all", "a", false, "list all nodes include not alive")
}

var LsCmd = &cobra.Command{
	Use:   "ls",
	Short: "list the nodes",
	Run: func(cmd *cobra.Command, args []string) {
		ea := NewExitAction()
		ea.After = func() {
			fmt.Println()
			cmd.Help()
		}

		nodes, err := cronsun.GetNodes()
		if err != nil {
			ea.Exit(err.Error())
		}

		fmt.Print("ID")
		for i := 0; i < 5; i++ {
			fmt.Print("\t")
		}
		fmt.Print("ip\t\t\t")
		fmt.Print("pid\t\t")
		fmt.Print("hostname\t")
		fmt.Print("alived\t")
		fmt.Println()

		for _, item := range nodes {
			if !all && !item.Alived {
				continue
			}
			fmt.Print(item.ID + "\t")
			fmt.Print(item.IP + "\t\t")
			fmt.Print(item.PID + "\t\t")
			fmt.Print(item.Hostname + "\t\t")
			if item.Alived {
				fmt.Print("Yes" + "\t\t")
			} else {
				fmt.Print("No" + "\t\t")
			}

		}
	},
}
