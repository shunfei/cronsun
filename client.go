package cronsun

import (
	"context"
	"fmt"
	"strings"
	"time"

	client "github.com/coreos/etcd/clientv3"

	"github.com/shunfei/cronsun/conf"
	"github.com/shunfei/cronsun/log"
)

var (
	DefalutClient *Client
)

type Client struct {
	*client.Client

	reqTimeout time.Duration
}

func NewClient(cfg *conf.Conf) (c *Client, err error) {
	log.Infof("NewClient cfg:%+v", cfg)
	cli, err := client.New(cfg.Etcd.Copy())
	if err != nil {
		return
	}

	c = &Client{
		Client: cli,

		reqTimeout: time.Duration(cfg.ReqTimeout) * time.Second,
	}
	return
}

func (c *Client) Put(key, val string, opts ...client.OpOption) (*client.PutResponse, error) {
	log.Infof("Put key:%s", key)
	ctx, cancel := NewEtcdTimeoutContext(c)
	defer cancel()
	return c.Client.Put(ctx, key, val, opts...)
}

func (c *Client) PutWithModRev(key, val string, rev int64) (*client.PutResponse, error) {
	log.Infof("PutWithModRev key:%s, rev:%d", key, rev)
	if rev == 0 {
		return c.Put(key, val)
	}

	ctx, cancel := NewEtcdTimeoutContext(c)
	tresp, err := DefalutClient.Txn(ctx).
		If(client.Compare(client.ModRevision(key), "=", rev)).
		Then(client.OpPut(key, val)).
		Commit()
	cancel()
	if err != nil {
		return nil, err
	}

	if !tresp.Succeeded {
		return nil, ErrValueMayChanged
	}

	resp := client.PutResponse(*tresp.Responses[0].GetResponsePut())
	return &resp, nil
}

func (c *Client) Get(key string, opts ...client.OpOption) (*client.GetResponse, error) {
	log.Infof("Get key:%s", key)
	ctx, cancel := NewEtcdTimeoutContext(c)
	defer cancel()
	return c.Client.Get(ctx, key, opts...)
}

func (c *Client) Delete(key string, opts ...client.OpOption) (*client.DeleteResponse, error) {
	log.Infof("Delete key:%s", key)
	ctx, cancel := NewEtcdTimeoutContext(c)
	defer cancel()
	return c.Client.Delete(ctx, key, opts...)
}

func (c *Client) Watch(key string, opts ...client.OpOption) client.WatchChan {
	log.Infof("Wath key:%s", key)
	return c.Client.Watch(context.Background(), key, opts...)
}

func (c *Client) Grant(ttl int64) (*client.LeaseGrantResponse, error) {
	log.Infof("Grant ttl:%d", ttl)
	ctx, cancel := NewEtcdTimeoutContext(c)
	defer cancel()
	return c.Client.Grant(ctx, ttl)
}

func (c *Client) Revoke(id client.LeaseID) (*client.LeaseRevokeResponse, error) {
	log.Infof("Revoke id:%d", id)
	ctx, cancel := context.WithTimeout(context.Background(), c.reqTimeout)
	defer cancel()
	return c.Client.Revoke(ctx, id)
}

func (c *Client) KeepAliveOnce(id client.LeaseID) (*client.LeaseKeepAliveResponse, error) {
	log.Infof("KeepAliveOnce id:%d", id)
	ctx, cancel := NewEtcdTimeoutContext(c)
	defer cancel()
	return c.Client.KeepAliveOnce(ctx, id)
}

func (c *Client) GetLock(key string, id client.LeaseID) (bool, error) {
	log.Infof("GetLock key:%s, id:%d", key, id)
	key = conf.Config.Lock + key
	ctx, cancel := NewEtcdTimeoutContext(c)
	resp, err := DefalutClient.Txn(ctx).
		If(client.Compare(client.CreateRevision(key), "=", 0)).
		Then(client.OpPut(key, "", client.WithLease(id))).
		Commit()
	cancel()

	if err != nil {
		return false, err
	}

	return resp.Succeeded, nil
}

func (c *Client) DelLock(key string) error {
	log.Infof("DelLock key:%s", key)
	_, err := c.Delete(conf.Config.Lock + key)
	return err
}

func IsValidAsKeyPath(s string) bool {
	return strings.IndexAny(s, "/\\") == -1
}

// etcdTimeoutContext return better error info
type etcdTimeoutContext struct {
	context.Context

	etcdEndpoints []string
}

func (c *etcdTimeoutContext) Err() error {
	err := c.Context.Err()
	if err == context.DeadlineExceeded {
		err = fmt.Errorf("%s: etcd(%v) maybe lost",
			err, c.etcdEndpoints)
	}
	return err
}

// NewEtcdTimeoutContext return a new etcdTimeoutContext
func NewEtcdTimeoutContext(c *Client) (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithTimeout(context.Background(), c.reqTimeout)
	etcdCtx := &etcdTimeoutContext{}
	etcdCtx.Context = ctx
	etcdCtx.etcdEndpoints = c.Endpoints()
	return etcdCtx, cancel
}
