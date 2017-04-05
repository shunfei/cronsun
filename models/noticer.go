package models

import (
	"time"

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

func NewMail() (m *Mail, err error) {
	cf := conf.Config.Mail
	sc, err := cf.Dialer.Dial()
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

			sm := gomail.NewMessage()
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
