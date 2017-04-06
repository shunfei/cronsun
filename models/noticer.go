package models

import (
	"encoding/json"
	"fmt"
	"time"

	client "github.com/coreos/etcd/clientv3"
	"github.com/go-gomail/gomail"

	"sunteng/commons/log"
	"sunteng/cronsun/conf"
)

type Noticer interface {
	Serve()
	Send(*Message)
}

type Message struct {
	Subject string
	Body    string
	To      []string
}

type Mail struct {
	cf      *conf.MailConf
	open    bool
	sc      gomail.SendCloser
	timer   *time.Timer
	msgChan chan *Message
}

func NewMail(timeout time.Duration) (m *Mail, err error) {
	var (
		sc   gomail.SendCloser
		done = make(chan struct{})
		cf   = conf.Config.Mail
	)

	// qq 邮箱的 Auth 出错后， 501 命令超时 2min 才能退出
	go func() {
		sc, err = cf.Dialer.Dial()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(timeout):
		err = fmt.Errorf("connect to smtp timeout")
	}

	if err != nil {
		return
	}

	m = &Mail{
		cf:      cf,
		open:    true,
		sc:      sc,
		timer:   time.NewTimer(time.Duration(cf.Keepalive) * time.Second),
		msgChan: make(chan *Message, 8),
	}
	return
}

func (m *Mail) Serve() {
	var err error
	sm := gomail.NewMessage()
	for {
		select {
		case msg := <-m.msgChan:
			m.timer.Reset(time.Duration(m.cf.Keepalive) * time.Second)
			if !m.open {
				if m.sc, err = m.cf.Dialer.Dial(); err != nil {
					log.Warnf("send msg[%+v] err: %s", msg, err.Error())
					continue
				}
				m.open = true
			}

			sm.Reset()
			sm.SetHeader("From", m.cf.Username)
			sm.SetHeader("To", msg.To...)
			sm.SetHeader("Subject", msg.Subject)
			sm.SetBody("text/plain", msg.Body)
			if err := gomail.Send(m.sc, sm); err != nil {
				log.Warnf("send msg[%+v] err: %s", msg, err.Error())
			}
		case <-m.timer.C:
			if m.open {
				if err = m.sc.Close(); err != nil {
					log.Warnf("close smtp server err: %s", err.Error())
				} else {
					m.open = false
				}
			}
			m.timer.Reset(time.Duration(m.cf.Keepalive) * time.Second)
		}
	}
}

func (m *Mail) Send(msg *Message) {
	m.msgChan <- msg
}

func StartNoticer(n Noticer) {
	go n.Serve()
	rch := DefalutClient.Watch(conf.Config.Noticer, client.WithPrefix())
	var err error
	for wresp := range rch {
		for _, ev := range wresp.Events {
			switch {
			case ev.IsCreate(), ev.IsModify():
				msg := new(Message)
				if err = json.Unmarshal(ev.Kv.Value, msg); err != nil {
					log.Warnf("msg[%s] umarshal err: %s", string(ev.Kv.Value), err.Error())
					continue
				}
				n.Send(msg)
			}
		}
	}
}
