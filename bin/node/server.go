// node 服务
// 用于在所需要执行 cron 任务的机器启动服务，替代 cron 执行所需的任务
package main

import (
	"flag"
	"runtime"

	"sunteng/commons/event"
	"sunteng/commons/log"
	"sunteng/commons/util"

	"sunteng/cronsun/conf"
)

var (
	gomax = flag.Int("gomax",
		4, "GOMAXPROCS: the max number of operating system threads that can execute")
	localIp = "cronsun_node"
)

func main() {
	flag.Parse()
	//set cpu usage
	runtime.GOMAXPROCS(*gomax)

	if err := conf.Init(); err != nil {
		log.Error(err.Error())
		return
	}

	if ip, err := util.GetLocalIP(); err != nil {
		log.Errorf("local ip error, node init may be fail, error: %s", err.Error())
	} else {
		localIp = ip.String()
	}

	log.Noticef("cronsun node[%s] service started, Ctrl+C or send kill sign to exit", localIp)
	// 注册退出事件
	event.On(event.EXIT)
	// 监听退出信号
	event.Wait()
	// 处理退出事件
	event.Emit(event.EXIT, nil)
	log.Notice("exit success")
}
