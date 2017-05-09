package main

import (
	"net"
	"time"

	"github.com/cockroachdb/cmux"

	"sunteng/commons/event"
	"sunteng/commons/log"
	"github.com/shunfei/cronsun/conf"
	"github.com/shunfei/cronsun/models"
	"github.com/shunfei/cronsun/web"
)

func main() {
	if err := models.Init(); err != nil {
		log.Error(err.Error())
		return
	}

	l, err := net.Listen("tcp", conf.Config.Web.BindAddr)
	if err != nil {
		log.Error(err.Error())
		return
	}

	// Create a cmux.
	m := cmux.New(l)
	httpL := m.Match(cmux.HTTP1Fast())
	httpServer, err := web.InitRouters()
	if err != nil {
		log.Error(err.Error())
		return
	}

	if conf.Config.Mail.Enable {
		var noticer models.Noticer

		if len(conf.Config.Mail.HttpAPI) > 0 {
			noticer = &models.HttpAPI{}
		} else {
			mailer, err := models.NewMail(10 * time.Second)
			if err != nil {
				log.Error(err.Error())
				return
			}
			noticer = mailer
		}
		go models.StartNoticer(noticer)
	}

	go func() {
		err := httpServer.Serve(httpL)
		if err != nil {
			panic(err.Error())
		}
	}()

	go m.Serve()

	log.Noticef("cronsun web server started on %s, Ctrl+C or send kill sign to exit", conf.Config.Web.BindAddr)
	// 注册退出事件
	event.On(event.EXIT, conf.Exit)
	// 监听退出信号
	event.Wait()
	event.Emit(event.EXIT, nil)
	log.Notice("exit success")
}
