package server

import (
	"context"
	"time"

	"cronjob/common"

	"go.etcd.io/etcd/clientv3"
)

type WorkerMgr struct {
	client *clientv3.Client
	kv     clientv3.KV
	lease  clientv3.Lease
}

var (
	GWorkerMgr *WorkerMgr
)

func InitWorkerMgr() error {
	// 初始化配置
	config := clientv3.Config{
		Endpoints:   GConfig.EtcdEndPoints, // 集群地址
		DialTimeout: time.Duration(GConfig.EtcdDialTimeout) * time.Millisecond,
	}

	// 建立连接
	client, err := clientv3.New(config)
	if err != nil {
		return err
	}

	// kv, lease
	kv := clientv3.NewKV(client)
	lease := clientv3.NewLease(client)

	GWorkerMgr = &WorkerMgr{
		client: client,
		kv:     kv,
		lease:  lease,
	}
	return nil
}

func (workerMgr *WorkerMgr) ListWorkers() ([]string, error) {

	workerArr := make([]string, 0)

	// 获取目录下所有kv
	getResp, err := workerMgr.kv.Get(context.TODO(), common.JobWorkerDir, clientv3.WithPrefix())
	if err != nil {
		return workerArr, err
	}

	// 解析每个节点的IP
	for _, kv := range getResp.Kvs {
		workerIP := common.ExtractWorkerIP(string(kv.Key))
		workerArr = append(workerArr, workerIP)
	}
	return workerArr, nil
}
