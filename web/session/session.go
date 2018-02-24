package session

import (
	"bytes"
	"encoding/gob"
	"errors"
	"net/http"

	client "github.com/coreos/etcd/clientv3"
	"github.com/shunfei/cronsun"
	"github.com/shunfei/cronsun/conf"
	"github.com/shunfei/cronsun/log"
	"github.com/shunfei/cronsun/utils"
)

func init() {
	gob.Register(cronsun.Administrator)
	gob.Register(cronsun.Developer)
}

var Manager SessionManager
var cookieCharacters = []byte("0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

type SessionManager interface {
	Get(w http.ResponseWriter, r *http.Request) (*Session, error)
	Store(*Session) error
	Destroy(w http.ResponseWriter, r *http.Request)
	CleanSeesionData(id string)
}

type storeData struct {
	leaseID client.LeaseID
	Email   string
	Data    map[interface{}]interface{}
}

type Session struct {
	m   SessionManager
	key string
	storeData
}

func (s *Session) ID() string {
	return s.key
}
func (s *Session) Store() error {
	err := s.m.Store(s)
	if err != nil {
		log.Errorf("Failed to store session[%s]: %s", s.key, err.Error())
	}
	return err
}

type EtcdStore struct {
	client *cronsun.Client
	conf   conf.SessionConfig
}

func NewEtcdStore(cli *cronsun.Client, conf conf.SessionConfig) *EtcdStore {
	return &EtcdStore{
		client: cli,
		conf:   conf,
	}
}

func (this *EtcdStore) Get(w http.ResponseWriter, r *http.Request) (sess *Session, err error) {
	c, err := r.Cookie(this.conf.CookieName)
	if err != nil && err != http.ErrNoCookie {
		log.Infof("get cookie err: %s", err.Error())
	} else {
		err = nil
	}

	sess = &Session{
		m: this,
		storeData: storeData{
			Data: make(map[interface{}]interface{}, 2),
		},
	}

	if c == nil {
		sess.key = utils.RandString(32, cookieCharacters...)
		c = &http.Cookie{
			Name:     this.conf.CookieName,
			Value:    sess.key,
			Path:     "/",
			HttpOnly: true,
			Secure:   false,
			MaxAge:   this.conf.Expiration,
		}
		http.SetCookie(w, c)
		r.AddCookie(c)

		return
	}

	sess.key = c.Value
	resp, err := this.client.Get(this.storeKey(c.Value))
	if err != nil {
		return
	}

	if len(resp.Kvs) == 0 {
		return sess, nil
	}

	var buffer = bytes.NewBuffer(resp.Kvs[0].Value)
	dec := gob.NewDecoder(buffer)
	err = dec.Decode(&sess.storeData)
	return
}

func (this *EtcdStore) Store(sess *Session) (err error) {
	var buffer = bytes.NewBuffer(nil)
	enc := gob.NewEncoder(buffer)
	err = enc.Encode(sess.storeData)
	if err != nil {
		return
	}

	if sess.leaseID == 0 {
		lresp, err := this.client.Grant(int64(this.conf.Expiration))
		if err != nil {
			return errors.New("etcd create new lease faild: " + err.Error()) // err
		}
		sess.leaseID = lresp.ID
	}

	_, err = this.client.Put(this.storeKey(sess.key), buffer.String(), client.WithLease(sess.leaseID))
	this.client.KeepAliveOnce(sess.leaseID)
	return
}

func (this *EtcdStore) Destroy(w http.ResponseWriter, r *http.Request) {
	c, err := r.Cookie(this.conf.CookieName)
	if err != nil || c == nil {
		return
	}
	this.CleanSeesionData(c.Value)
}

func (this *EtcdStore) CleanSeesionData(id string) {
	_, err := this.client.Delete(this.storeKey(id))
	if err != nil {
		log.Errorf("Failed to remove session [%s] from etcd: %s", this.storeKey(id), err.Error())
	}
}

func (this *EtcdStore) storeKey(key string) string {
	return this.conf.StorePrefixPath + key
}
