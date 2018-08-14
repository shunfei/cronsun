// node 服务
// 用于在所需要执行 cron 任务的机器启动服务，替代 cron 执行所需的任务
package main

import (
	"flag"
	slog "log"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/shunfei/cronsun"
	"github.com/shunfei/cronsun/conf"
	"github.com/shunfei/cronsun/event"
	"github.com/shunfei/cronsun/log"
	"github.com/shunfei/cronsun/node"
)

var (
	level    = flag.Int("l", 0, "log level, -1:debug, 0:info, 1:warn, 2:error")
	confFile = flag.String("conf", "conf/files/base.json", "config file path")
)

func main() {
	flag.Parse()

	lcf := zap.NewDevelopmentConfig()
	lcf.Level.SetLevel(zapcore.Level(*level))
	lcf.Development = false
	logger, err := lcf.Build(zap.AddCallerSkip(1))
	if err != nil {
		slog.Fatalln("new log err:", err.Error())
	}
	log.SetLogger(logger.Sugar())

	if err = cronsun.Init(*confFile, true); err != nil {
		log.Errorf(err.Error())
		return
	}

	n, err := node.NewNode(conf.Config)
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	if err = n.Register(); err != nil {
		log.Errorf(err.Error())
		return
	}

	if err = cronsun.StartProc(); err != nil {
		log.Warnf("[process key will not timeout]proc lease id set err: %s", err.Error())
	}

	if err = n.Run(); err != nil {
		log.Errorf(err.Error())
		return
	}

	log.Infof("cronsun %s service started, Ctrl+C or send kill sign to exit", n.String())
	// 注册退出事件
	event.On(event.EXIT, n.Stop, conf.Exit, cronsun.Exit)
	// 注册监听配置更新事件
	event.On(event.WAIT, cronsun.Reload)
	// 监听退出信号
	event.Wait()
	// 处理退出事件
	event.Emit(event.EXIT, nil)
	log.Infof("exit success")
}
