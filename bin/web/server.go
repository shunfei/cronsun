package main

import (
	"net"
	"time"

	"github.com/cockroachdb/cmux"

	"github.com/shunfei/cronsun"
	"github.com/shunfei/cronsun/conf"
	"github.com/shunfei/cronsun/event"
	"github.com/shunfei/cronsun/log"
	"github.com/shunfei/cronsun/web"
)

func main() {
	if err := cronsun.Init(); err != nil {
		log.Errorf(err.Error())
		return
	}

	l, err := net.Listen("tcp", conf.Config.Web.BindAddr)
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	// Create a cmux.
	m := cmux.New(l)
	httpL := m.Match(cmux.HTTP1Fast())
	httpServer, err := web.InitRouters()
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	if conf.Config.Mail.Enable {
		var noticer cronsun.Noticer

		if len(conf.Config.Mail.HttpAPI) > 0 {
			noticer = &cronsun.HttpAPI{}
		} else {
			mailer, err := cronsun.NewMail(10 * time.Second)
			if err != nil {
				log.Errorf(err.Error())
				return
			}
			noticer = mailer
		}
		go cronsun.StartNoticer(noticer)
	}

	go func() {
		err := httpServer.Serve(httpL)
		if err != nil {
			panic(err.Error())
		}
	}()

	go m.Serve()

	log.Infof("cronsun web server started on %s, Ctrl+C or send kill sign to exit", conf.Config.Web.BindAddr)
	// 注册退出事件
	event.On(event.EXIT, conf.Exit)
	// 监听退出信号
	event.Wait()
	event.Emit(event.EXIT, nil)
	log.Infof("exit success")
}
