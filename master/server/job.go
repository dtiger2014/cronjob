package server

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	common "cronjob/common"

	"go.etcd.io/etcd/clientv3"
)

type JobMgr struct {
	client *clientv3.Client
	kv     clientv3.KV
	lease  clientv3.Lease
}

var (
	GJobMgr *JobMgr
)

func InitJobMgr() error {
	var (
		config clientv3.Config
		client *clientv3.Client
		kv     clientv3.KV
		lease  clientv3.Lease
		err    error
	)

	// 初始化配置
	fmt.Println(GConfig.EtcdEndPoints)
	config = clientv3.Config{
		Endpoints:   GConfig.EtcdEndPoints,
		DialTimeout: time.Duration(GConfig.EtcdDialTimeout) * time.Millisecond,
	}

	// 连接etcd
	if client, err = clientv3.New(config); err != nil {
		return err
	}

	fmt.Println(client)

	kv = clientv3.NewKV(client)
	lease = clientv3.NewLease(client)

	// 设置GJobMgr
	GJobMgr = &JobMgr{
		client: client,
		kv:     kv,
		lease:  lease,
	}

	return nil
}

func (jobmgr *JobMgr) CheckJobExist(jobName string) bool {
	jobKey := common.JobSaveDir + jobName

	// save to etcd
	getResp, err := jobmgr.kv.Get(context.TODO(), jobKey)
	if err != nil {
		return false
	}
	if len(getResp.Kvs) > 0 {
		return false
	}
	return true
}

func (jobmgr *JobMgr) SaveJob(job *common.Job) (*common.Job, error) {

	jobKey := common.JobSaveDir + job.Name

	jobValue, err := json.Marshal(job)
	if err != nil {
		return nil, err
	}

	// save to etcd
	putResp, err := jobmgr.kv.Put(context.TODO(), jobKey, string(jobValue), clientv3.WithPrevKV())
	if err != nil {
		return nil, err
	}

	var oldJobObj common.Job

	if putResp.PrevKv != nil {
		// unmarshal old value
		if err = json.Unmarshal(putResp.PrevKv.Value, &oldJobObj); err != nil {
			return nil, err
		}
	}

	return &oldJobObj, nil
}

func (jobmgr *JobMgr) DeleteJob(name string) (*common.Job, error) {
	jobKey := common.JobSaveDir + name

	// 从 etcd 删除
	delResp, err := jobmgr.kv.Delete(context.TODO(), jobKey, clientv3.WithPrevKV())
	if err != nil {
		return nil, err
	}

	var oldJobObj common.Job
	if len(delResp.PrevKvs) != 0 {
		err := json.Unmarshal(delResp.PrevKvs[0].Value, &oldJobObj)
		if err != nil {
			return nil, err
		}
	}

	return &oldJobObj, nil
}

func (jobmgr *JobMgr) ListJobs() ([]*common.Job, error) {
	// key
	dirKey := common.JobSaveDir

	// 获取
	getResp, err := jobmgr.kv.Get(context.TODO(), dirKey, clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}

	jobList := make([]*common.Job, 0)

	for _, kvPair := range getResp.Kvs {
		job := &common.Job{}
		err := json.Unmarshal(kvPair.Value, job)
		if err != nil {
			continue
		}
		jobList = append(jobList, job)
	}
	return jobList, nil
}

func (jobmgr *JobMgr) KillJob(name string) error {

	killerKey := common.JobKillerDir + name

	leaseGrantResp, err := jobmgr.lease.Grant(context.TODO(), 1)
	if err != nil {
		return err
	}

	// leaseID
	leaseID := leaseGrantResp.ID

	// 设置 kill job flag
	_, err = jobmgr.kv.Put(context.TODO(), killerKey, "", clientv3.WithLease(leaseID))
	if err != nil {
		return err
	}
	return nil
}
