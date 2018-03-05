package cmd

import (
	"archive/zip"
	"encoding/binary"
	"fmt"
	"io"
	"strings"
	"sync"

	"github.com/spf13/cobra"

	"github.com/shunfei/cronsun"
	"github.com/shunfei/cronsun/conf"
)

var restoreFile string

func init() {
	RestoreCmd.Flags().StringVarP(&restoreFile, "file", "f", "", "the backup zip file")
}

var RestoreCmd = &cobra.Command{
	Use:   "restore",
	Short: "restore job & group data",
	Run: func(cmd *cobra.Command, args []string) {
		var err error
		var ea = NewExitAction()

		restoreFile = strings.TrimSpace(restoreFile)
		if len(restoreFile) == 0 {
			ea.Exit("backup file is required")
		}

		r, err := zip.OpenReader(restoreFile)
		ea.ExitOnErr(err)
		ea.Defer = func() {
			r.Close()
		}

		restoreChan, wg := startRestoreProcess()
		for _, f := range r.File {
			var keyPrefix string
			switch f.Name {
			case "job":
				keyPrefix = conf.Config.Cmd
			case "node_group":
				keyPrefix = conf.Config.Group
			}

			rc, err := f.Open()
			ea.ExitOnErr(err)

			ea.ExitOnErr(restoreKvs(rc, keyPrefix, restoreChan, wg))
			rc.Close()
		}

		wg.Wait()
		close(restoreChan)
	},
}

type kv struct {
	k, v string
}

var (
	keyLenBuf = make([]byte, 2)
	valLenBuf = make([]byte, 4)
	keyBuf    = make([]byte, 256)
	valBuf    = make([]byte, 1024)
)

func restoreKvs(r io.Reader, keyPrefix string, storeChan chan *kv, wg *sync.WaitGroup) error {
	for {
		// read length of key
		n, err := r.Read(keyLenBuf)
		if err == io.EOF && n != 0 {
			return fmt.Errorf("unexcepted data, the file may broken")
		} else if err == io.EOF && n == 0 {
			break
		} else if err != nil {
			return err
		}
		keylen := binary.LittleEndian.Uint16(keyLenBuf)

		// read key
		if n, err = r.Read(keyBuf[:keylen]); err != nil {
			return err
		}
		key := keyBuf[:keylen]

		// read length of value
		if n, err = r.Read(valLenBuf); err != nil {
			return err
		}
		vallen := binary.LittleEndian.Uint32(valLenBuf)

		// read value
		if len(valBuf) < int(vallen) {
			valBuf = make([]byte, vallen*2)
		}
		if n, err = r.Read(valBuf[:vallen]); err != nil && err != io.EOF {
			return err
		}
		value := valBuf[:vallen]

		wg.Add(1)
		storeChan <- &kv{
			k: keyPrefix + string(key),
			v: string(value),
		}
	}

	return nil
}

func startRestoreProcess() (chan *kv, *sync.WaitGroup) {
	c := make(chan *kv, 0)
	wg := &sync.WaitGroup{}

	const maxResries = 3
	go func() {
		for _kv := range c {
			for retries := 1; retries <= maxResries; retries++ {
				_, err := cronsun.DefalutClient.Put(_kv.k, _kv.v)
				if err != nil {
					if retries == maxResries {
						fmt.Println("[Error] restore err:", err)
						fmt.Println("\tKey: ", string(_kv.k))
						fmt.Println("\tValue: ", string(_kv.v))
					}
					continue
				}
			}

			wg.Done()
		}
	}()

	return c, wg
}
