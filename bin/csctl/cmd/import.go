package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"os/exec"
	"strings"
	"time"

	"github.com/shunfei/cronsun"
	cron2 "github.com/shunfei/cronsun/node/cron"
	"github.com/spf13/cobra"
)

type cron struct {
	timer string
	cmd   string
}

var (
	importNodes string
)

func init() {
	ImportCmd.Flags().StringVar(&importNodes, "nodes", "", `the node ids that needs to run these imported job, 
				split by ',', e.g: '--nodes=aa,bb,cc', empty means no node will run`)
}

var ImportCmd = &cobra.Command{
	Use:   "import",
	Short: `it will load the job from the crontab, but you must to confirm you can execute 'crontab -l'`,
	Run: func(cmd *cobra.Command, args []string) {
		var nodeInclude []string
		if len(importNodes) > 0 {
			nodeInclude = strings.Split(importNodes, spliter)
		}

		ea := NewExitAction()
		crons, err := loadCrons()
		if err != nil {
			ea.Exit("load crontab failed,err:%s", err.Error())
		}
		total := len(crons)
		var successCount int
		ea.After = func() {
			fmt.Printf("total:%d,success:%d,failed:%d\n", total, successCount, total-successCount)
			if err := cmd.Help(); err != nil {
				return
			}
		}
		rand.Seed(time.Now().Unix())
		for _, cron := range crons {
			job := cronsun.Job{}
			job.ID = cronsun.NextID()
			job.Command = cron.cmd
			jr := &cronsun.JobRule{
				Timer: "* " + cron.timer,
			}
			jr.NodeIDs = nodeInclude
			job.Name = fmt.Sprintf("crontab-%d", rand.Intn(1000))
			job.Group = "crontab"
			job.Rules = append(job.Rules, jr)
			// 默认先暂停
			job.Pause = true
			if err := job.Check(); err != nil {
				ea.Exit("job check error:%s", err.Error())
			}
			b, err := json.Marshal(job)
			if err != nil {
				ea.Exit("json marshal error:%s", err.Error())
			}

			_, err = cronsun.DefalutClient.Put(job.Key(), string(b))
			if err != nil {
				ea.Exit("etcd put error:%s", err.Error())
			}
			successCount++
			fmt.Printf("crontab-%s %s has import to the cronsun, the job id is:%s\n", cron.timer, cron.cmd, job.ID)
		}

		fmt.Printf("import fininsh,succes:%d\n", successCount)
	},
}

func checkCrons(crons []string) (invalid []string) {
	for _, item := range crons {
		item = strings.TrimSpace(item)
		if item != "" && !strings.HasPrefix(item, "#") {
			expr := strings.Fields(item)
			expr = expr[:5]
			_, err := cron2.ParseStandard(strings.Join(expr, " "))
			if err != nil {
				invalid = append(invalid, item)
			}
		}
	}
	return
}

func loadCrons() (crons []cron, err error) {
	var b bytes.Buffer
	cmd := exec.Command("crontab", "-l")
	cmd.Stdout = &b
	cmd.Stderr = &b
	err = cmd.Run()
	if err != nil {
        return
	}

	result := strings.Split(b.String(), "\n")
	invalid := checkCrons(result)
	if len(invalid) > 0 {
		title := fmt.Sprintf("There are %d invalid cron expression,please check them at first.\n", len(invalid))
		err = fmt.Errorf(title + strings.Join(invalid, "\n"))
		return
	}

	for _, item := range result {
		item = strings.TrimSpace(item)
		if item != "" && !strings.HasPrefix(item, "#") {
			spec := strings.Split(item, " ")
			timer := strings.Join(spec[:5], " ")
			cmd := strings.Join(spec[5:], " ")
			crons = append(crons, cron{timer, cmd})
		}
	}
	return
}
