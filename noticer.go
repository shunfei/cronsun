package cronsun

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	client "github.com/coreos/etcd/clientv3"
	"github.com/go-gomail/gomail"

	"github.com/shunfei/cronsun/conf"
	"github.com/shunfei/cronsun/log"
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
					log.Warnf("smtp send msg[%+v] err: %s", msg, err.Error())
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
				log.Warnf("smtp send msg[%+v] err: %s", msg, err.Error())
			}
		case <-m.timer.C:
			if m.open {
				if err = m.sc.Close(); err != nil {
					log.Warnf("close smtp server err: %s", err.Error())
				}
				m.open = false
			}
			m.timer.Reset(time.Duration(m.cf.Keepalive) * time.Second)
		}
	}
}

func (m *Mail) Send(msg *Message) {
	m.msgChan <- msg
}

type HttpAPI struct{}

func (h *HttpAPI) Serve() {}

func (h *HttpAPI) Send(msg *Message) {
	body, err := json.Marshal(msg)
	if err != nil {
		log.Warnf("http api send msg[%+v] err: %s", msg, err.Error())
		return
	}

	req, err := http.NewRequest("POST", conf.Config.Mail.HttpAPI, bytes.NewBuffer(body))
	if err != nil {
		return
	}

	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Warnf("http api send msg[%+v] err: %s", msg, err.Error())
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		return
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Warnf("http api send msg[%+v] err: %s", msg, err.Error())
		return
	}
	log.Warnf("http api send msg[%+v] err: %s", msg, string(data))
	return
}

func StartNoticer(n Noticer) {
	go n.Serve()
	go monitorNodes(n)

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

				if len(conf.Config.Mail.To) > 0 {
					msg.To = append(msg.To, conf.Config.Mail.To...)
				}
				n.Send(msg)
			}
		}
	}
}

func monitorNodes(n Noticer) {
	var (
		err error
		ok  bool
		id  string
	)
	rch := WatchNode()

	for wresp := range rch {
		for _, ev := range wresp.Events {
			switch {
			case ev.Type == client.EventTypeDelete:
				id = GetIDFromKey(string(ev.Kv.Key))
				ok, err = ISNodeAlive(id)
				if err != nil {
					log.Warnf("query node[%s] err: %s", id, err.Error())
					continue
				}

				if ok {
					n.Send(&Message{
						Subject: "Node[" + id + "] break away cluster, this happed at " + time.Now().Format(time.RFC3339),
						Body:    "Node breaked away cluster, this might happed when node crash or network problems.",
						To:      conf.Config.Mail.To,
					})
				}
			}
		}
	}
}
