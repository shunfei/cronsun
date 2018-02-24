package cmd

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path"
	"strings"
	"time"

	"github.com/coreos/etcd/clientv3"

	"github.com/spf13/cobra"

	"github.com/shunfei/cronsun"
	"github.com/shunfei/cronsun/conf"
)

var backupOutput string

var BackupCmd = &cobra.Command{
	Use:   "backup",
	Short: "Backup job & group data",
	Run: func(cmd *cobra.Command, args []string) {
		var err error
		var ea = NewExitAction()

		backupOutput = strings.TrimSpace(backupOutput)
		if len(backupOutput) > 0 {
			err = os.MkdirAll(backupOutput, os.ModeDir)
			if err != nil {
				ea.Exit("failed to make directory %s, err: %s", backupOutput, err)
			}
		}

		name := path.Join(backupOutput, time.Now().Format("20060102_150405")+".zip")
		f, err := os.OpenFile(name, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0600)
		ea.ExitOnErr(err)
		ea.Defer = func() {
			f.Close()
			if err != nil {
				os.Remove(name)
			}
		}

		var waitForStore = [][]string{
			// [file name in  ZIP archive, key prefix in etcd]
			[]string{"job", conf.Config.Cmd},
			[]string{"node_group", conf.Config.Group},
		}
		zw := zip.NewWriter(f)

		for i := range waitForStore {
			zf, err := zw.Create(waitForStore[i][0])
			ea.ExitOnErr(err)
			ea.ExitOnErr(storeKvs(zf, waitForStore[i][1]))
		}

		ea.ExitOnErr(zw.Close())
	},
}

func init() {
	BackupCmd.Flags().StringVarP(&backupOutput, "output-dir", "o", "", "the directory for store backup file")
}

type ExitAction struct {
	Defer func()
}

func NewExitAction() *ExitAction {
	return &ExitAction{}
}

func (ea *ExitAction) ExitOnErr(err error) {
	if err != nil {
		ea.Exit(err.Error())
	}
}

func (ea *ExitAction) Exit(format string, v ...interface{}) {
	if ea.Defer != nil {
		ea.Defer()
	}

	fmt.Printf(format+"\n", v...)
	os.Exit(1)
}

var (
	byteLD = []byte{'['}
	byteRD = []byte{']'}
)

func storeKvs(w io.Writer, keyPrefix string) error {
	gresp, err := cronsun.DefalutClient.Get(keyPrefix, clientv3.WithPrefix())
	if err != nil {
		return fmt.Errorf("failed to fetch %s from etcd: %s", keyPrefix, err)
	}

	w.Write(byteLD)
	for i := range gresp.Kvs {
		if len(gresp.Kvs)-1 != i {
			_, err = w.Write(append(gresp.Kvs[i].Value, ','))
		} else {
			_, err = w.Write(gresp.Kvs[i].Value)
		}

		if err != nil {
			return err
		}
	}
	w.Write(byteRD)

	return nil
}
