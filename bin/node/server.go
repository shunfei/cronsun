// node 服务
// 用于在所需要执行 cron 任务的机器启动服务，替代 cron 执行所需的任务
package main

import (
	"flag"
	"runtime"

	"sunteng/commons/event"
	"sunteng/commons/log"

	"sunteng/cronsun/conf"
	"sunteng/cronsun/models"
	"sunteng/cronsun/node"
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

	go n.Run()

	log.Noticef("cronsun %s service started, Ctrl+C or send kill sign to exit", n.String())
	// 注册退出事件
	event.On(event.EXIT, n.Stop)
	// 监听退出信号
	event.Wait()
	// 处理退出事件
	event.Emit(event.EXIT, nil)
	log.Notice("exit success")
}
