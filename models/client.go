package models

import (
	"context"
	"time"

	client "github.com/coreos/etcd/clientv3"

	"sunteng/cronsun/conf"
)

var (
	DefalutClient *Client

	initialized bool
)

func Init() (err error) {
	if initialized {
		return
	}

	if err = initID(); err != nil {
		return
	}

	if err = conf.Init(); err != nil {
		return
	}

	if DefalutClient, err = NewClient(conf.Config); err != nil {
		return
	}

	initialized = true
	return
}

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

func (c *Client) Get(key string, opts ...client.OpOption) (*client.GetResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), c.reqTimeout)
	defer cancel()
	return c.Client.Get(ctx, key, opts...)
}

func (c *Client) Del(key string, opts ...client.OpOption) (*client.DeleteResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), c.reqTimeout)
	defer cancel()
	return c.Client.Delete(ctx, key, opts...)
}

func (c *Client) Watch(key string, opts ...client.OpOption) client.WatchChan {
	ctx, cancel := context.WithTimeout(context.Background(), c.reqTimeout)
	defer cancel()
	return c.Client.Watch(ctx, key, opts...)
}
