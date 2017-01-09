package cronsun

import (
	"sunteng/cronsun/conf"

	"github.com/coreos/etcd/clientv3"
)

var etcdClient *clientv3.Client

func EtcdInstance() (*clientv3.Client, error) {
	if etcdClient != nil {
		return etcdClient, nil
	}

	if err := conf.Init(); err != nil {
		return nil, err
	}

	etcdClient, err := clientv3.New(conf.Config.Etcd)
	return etcdClient, err
}
