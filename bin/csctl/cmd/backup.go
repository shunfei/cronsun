package cmd

import (
	"archive/zip"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"path"
	"runtime"
	"strings"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/spf13/cobra"

	"github.com/shunfei/cronsun"
	"github.com/shunfei/cronsun/conf"
)

var (
	backupDir  string
	backupFile string
)

var BackupCmd = &cobra.Command{
	Use:   "backup",
	Short: "backup job & group data",
	Run: func(cmd *cobra.Command, args []string) {
		var err error
		var ea = NewExitAction()

		backupDir = strings.TrimSpace(backupDir)
		if len(backupDir) > 0 {
			err = os.MkdirAll(backupDir, 0755)
			if err != nil {
				ea.Exit("failed to make directory %s, err: %s", backupDir, err)
			}
		}

		backupFile = strings.TrimSpace(backupFile)
		if len(backupFile) == 0 {
			backupFile = time.Now().Format("20060102_150405")
		}
		backupFile += ".zip"

		name := path.Join(backupDir, backupFile)
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
			storeKvs(zf, waitForStore[i][1])
		}

		ea.ExitOnErr(zw.Close())
	},
}

func init() {
	BackupCmd.Flags().StringVarP(&backupDir, "dir", "d", "", "the directory to store backup file")
	BackupCmd.Flags().StringVarP(&backupFile, "file", "f", "", "the backup file name")
}

type ExitAction struct {
	Defer func()
}

func NewExitAction() *ExitAction {
	return &ExitAction{}
}

func (ea *ExitAction) ExitOnErr(err error) {
	if err != nil {
		_, f, l, _ := runtime.Caller(1)
		ea.Exit("%s line %d: %s", f, l, err.Error())
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
	sizeBuf = make([]byte, 2+4) // key length(uint16) + value length(uint32)
)

func storeKvs(w io.Writer, keyPrefix string) error {
	gresp, err := cronsun.DefalutClient.Get(keyPrefix, clientv3.WithPrefix())
	if err != nil {
		return fmt.Errorf("failed to fetch %s from etcd: %s", keyPrefix, err)
	}

	var prefixLen = len(keyPrefix)

	for i := range gresp.Kvs {
		key := gresp.Kvs[i].Key[prefixLen:]
		binary.LittleEndian.PutUint16(sizeBuf[:2], uint16(len(key)))
		binary.LittleEndian.PutUint32(sizeBuf[2:], uint32(len(gresp.Kvs[i].Value)))

		// length of key
		if _, err = w.Write(sizeBuf[:2]); err != nil {
			return err
		}
		if _, err = w.Write(key); err != nil {
			return err
		}

		// lenght of value
		if _, err = w.Write(sizeBuf[2:]); err != nil {
			return err
		}
		if _, err = w.Write(gresp.Kvs[i].Value); err != nil {
			return err
		}
	}

	return nil
}
