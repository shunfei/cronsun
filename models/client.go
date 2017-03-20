package models

import (
	"strings"
	"time"

	"golang.org/x/net/context"

	client "github.com/coreos/etcd/clientv3"

	"sunteng/cronsun/conf"
)

var (
	DefalutClient *Client
)

type Client struct {
	*client.Client

	reqTimeout time.Duration
}

func NewClient(cfg *conf.Conf) (c *Client, err error) {
	cli, err := client.New(cfg.Etcd)
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
	ctx, cancel := context.WithTimeout(context.Background(), c.reqTimeout)
	defer cancel()
	return c.Client.Put(ctx, key, val, opts...)
}

func (c *Client) PutWithModRev(key, val string, rev int64) (*client.PutResponse, error) {
	if rev == 0 {
		return c.Put(key, val)
	}

	ctx, cancel := context.WithTimeout(context.Background(), c.reqTimeout)
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
	ctx, cancel := context.WithTimeout(context.Background(), c.reqTimeout)
	defer cancel()
	return c.Client.Get(ctx, key, opts...)
}

func (c *Client) Delete(key string, opts ...client.OpOption) (*client.DeleteResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), c.reqTimeout)
	defer cancel()
	return c.Client.Delete(ctx, key, opts...)
}

func (c *Client) Watch(key string, opts ...client.OpOption) client.WatchChan {
	return c.Client.Watch(context.Background(), key, opts...)
}

func (c *Client) Grant(ttl int64) (*client.LeaseGrantResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), c.reqTimeout)
	defer cancel()
	return c.Client.Grant(ctx, ttl)
}

func (c *Client) KeepAliveOnce(id client.LeaseID) (*client.LeaseKeepAliveResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), c.reqTimeout)
	defer cancel()
	return c.Client.KeepAliveOnce(ctx, id)
}

func IsValidAsKeyPath(s string) bool {
	return strings.IndexByte(s, '/') == -1
}
