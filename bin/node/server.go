// node 服务
// 用于在所需要执行 cron 任务的机器启动服务，替代 cron 执行所需的任务
package main

import (
	"flag"
	"runtime"

	"sunteng/commons/log"

	"github.com/shunfei/cronsun/conf"
	"github.com/shunfei/cronsun/event"
	"github.com/shunfei/cronsun/models"
	"github.com/shunfei/cronsun/node"
)

var (
	gomax = flag.Int("gomax",
		4, "GOMAXPROCS: the max number of operating system threads that can execute")
)

func main() {
	flag.Parse()
	//set cpu usage
	runtime.GOMAXPROCS(*gomax)

	if err := models.Init(); err != nil {
		log.Error(err.Error())
		return
	}

	n, err := node.NewNode(conf.Config)
	if err != nil {
		log.Error(err.Error())
		return
	}

	if err = n.Register(); err != nil {
		log.Error(err.Error())
		return
	}

	if err = models.StartProc(); err != nil {
		log.Warnf("[process key will not timeout]proc lease id set err: %s", err.Error())
	}

	if err = n.Run(); err != nil {
		log.Error(err.Error())
		return
	}

	log.Noticef("cronsun %s service started, Ctrl+C or send kill sign to exit", n.String())
	// 注册退出事件
	event.On(event.EXIT, n.Stop, conf.Exit, models.Exit)
	// 注册监听配置更新事件
	event.On(event.WAIT, models.Reload)
	// 监听退出信号
	event.Wait()
	// 处理退出事件
	event.Emit(event.EXIT, nil)
	log.Notice("exit success")
}
