package server

import (
	"context"
	"cronjob/common"
	"time"

	"github.com/coreos/etcd/clientv3"
)

// 任务管理器
type JobMgr struct {
	client  *clientv3.Client
	kv      clientv3.KV
	lease   clientv3.Lease
	watcher clientv3.Watcher
}

var (
	GJobMgr *JobMgr
)

// 监听任务变化
func (jobMgr *JobMgr) watchJobs() error {

	// get /cron/jobs/目录下的所有任务，并且获得集群的revision
	getResp, err := jobMgr.kv.Get(context.TODO(), common.JobSaveDir, clientv3.WithPrefix())
	if err != nil {
		return err
	}

	// 当前任务
	for _, kvpair := range getResp.Kvs {
		// json unmarshal
		if job, err := common.UnpackJob(kvpair.Value); err != nil {
			jobEvent := common.BuildJobEvent(common.JobEventSave, job)

			// 同步给scheduler（调度协程）
			GScheduler.PushJobEvent(jobEvent)
		}
	}

	// 从该revision向后监听变化事件
	go func() { // 监听协程
		// 从GET时刻的后续版本开始监听变化
		watchStartRevision := getResp.Header.Revision + 1
		// 监听/cron/jobs/目录的后续变化
		watchChan := jobMgr.watcher.Watch(context.TODO(), common.JobSaveDir,
			clientv3.WithRev(watchStartRevision), clientv3.WithPrefix())
		// 处理监听事件
		for watchResp := range watchChan {
			for _, watchEvent := range watchResp.Events {
				var jobEvent *common.JobEvent
				switch watchEvent.Type {
				case clientv3.EventTypePut: // 任务保存事件
					job, err := common.UnpackJob(watchEvent.Kv.Value)
					if err != nil {
						continue
					}
					jobEvent = common.BuildJobEvent(common.JobEventSave, job)
				case clientv3.EventTypeDelete: // 任务被删除了
					// Delete /cron/jobs/job10
					jobName := common.ExtractJobName(string(watchEvent.Kv.Key))

					job := &common.Job{Name: jobName}

					// 构建一个删除Event
					jobEvent = common.BuildJobEvent(common.JobEventDelete, job)
				}
				// 变化推给scheduler
				GScheduler.PushJobEvent(jobEvent)
			}
		}
	}()
	return nil
}

// 监听强杀任务通知
func (jobMgr *JobMgr) watchKiller() {
	// 监听 /cron/killer
	go func() { // 监听协程
		// 监听/cron/killer/目录的变化
		watchChan := jobMgr.watcher.Watch(context.TODO(), common.JobKillerDir, clientv3.WithPrefix())
		// 处理监听事件
		for watchResp := range watchChan {
			for _, watchEvent := range watchResp.Events {
				switch watchEvent.Type {
				case clientv3.EventTypePut: // 杀死任务事件
					jobName := common.ExtractKillerName(string(watchEvent.Kv.Key))
					job := &common.Job{Name: jobName}
					jobEvent := common.BuildJobEvent(common.JobEventKill, job)

					// 事件推给scheduler
					GScheduler.PushJobEvent(jobEvent)
				case clientv3.EventTypeDelete: // killer标记过期，被自动删除
				}
			}
		}
	}()
}

// 初始化
func InitJobMgr() error {
	// 初始化配置
	config := clientv3.Config{
		Endpoints:   GConfig.EtcdEndpoints,                                     // 集群地址
		DialTimeout: time.Duration(GConfig.EtcdDialTimeout) * time.Millisecond, // 连接超时
	}

	// 建立连接
	client, err := clientv3.New(config)
	if err != nil {
		return err
	}

	// 得到kv和lease的API子集
	kv := clientv3.NewKV(client)
	lease := clientv3.NewLease(client)
	watcher := clientv3.NewWatcher(client)

	GJobMgr = &JobMgr{
		client:  client,
		kv:      kv,
		lease:   lease,
		watcher: watcher,
	}

	// 启动任务监听
	GJobMgr.watchJobs()

	// 启动监听killer
	GJobMgr.watchKiller()

	return nil
}

// 创建任务执行锁
func (jobMgr *JobMgr) CreateJobLock(jobName string) *JobLock {
	jobLock := InitJobLock(jobName, jobMgr.kv, jobMgr.lease)
	return jobLock
}
