package main

import (
	"net"

	"github.com/cockroachdb/cmux"

	"sunteng/commons/event"
	"sunteng/commons/log"
	"sunteng/cronsun"
	"sunteng/cronsun/conf"
	"sunteng/cronsun/web"
)

func main() {
	l, err := net.Listen("tcp", conf.Config.Web.BindAddr)
	if err != nil {
		cronsun.Fatalln(err)
	}

	// Create a cmux.
	m := cmux.New(l)
	httpL := m.Match(cmux.HTTP1Fast())
	httpServer, err := web.InitRouters()
	if err != nil {
		cronsun.Fatalln(err)
	}

	go httpServer.Serve(httpL)

	log.Noticef("cronsun web server started on %s, Ctrl+C or send kill sign to exit", conf.Config.Web.BindAddr)
	// 注册退出事件
	// event.On(event.EXIT, n.Stop)
	// 监听退出信号
	event.Wait()
	event.Emit(event.EXIT, nil)
	log.Notice("exit success")
}
