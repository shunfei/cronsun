package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/shunfei/cronsun"
	subcmd "github.com/shunfei/cronsun/bin/csctl/cmd"
)

var confFile string

var rootCmd = &cobra.Command{
	Use:   "csctl",
	Short: "cronsun command tools for data manage",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if err := cronsun.Init(confFile, false); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	},
	Run: func(cmd *cobra.Command, args []string) {},
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&confFile, "conf", "c", "", "base.json file path.")
	rootCmd.AddCommand(subcmd.BackupCmd, subcmd.RestoreCmd, subcmd.UpgradeCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
