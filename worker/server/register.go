package server

import (
	"context"
	"cronjob/common"
	"errors"
	"net"
	"time"

	"go.etcd.io/etcd/clientv3"
)

/* worker节点注册到etcd：/cron/wrokers/IP地址  */
type Register struct {
	client  *clientv3.Client
	kv      clientv3.KV
	lease   clientv3.Lease
	localIP string // 本机IP
}

var (
	GRegister *Register
)

// 获取本机网卡IP
func getLocalIP() (string, error) {
	// 获取所有网卡
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", err
	}

	// 取第一个非lo的网卡IP
	for _, addr := range addrs {
		if ipNet, isIpNet := addr.(*net.IPNet); isIpNet && !ipNet.IP.IsLoopback() {
			// 跳过IPV6
			if ipNet.IP.To4() != nil {
				ipv4 := ipNet.IP.String()
				return ipv4, nil
			}
		}
	}

	err = errors.New("Not found local IP")
	return "", err
}

// 注册到/cron/workers/IP, 并自动续约
func (register *Register) keepOnline() {
	var (
		regKey         string
		leaseGrantResp *clientv3.LeaseGrantResponse
		err            error
		keepAliveChan  <-chan *clientv3.LeaseKeepAliveResponse
		keepAliveResp  *clientv3.LeaseKeepAliveResponse
		cancelCtx      context.Context
		cancelFunc     context.CancelFunc
	)
	for {
		// 注册路径
		regKey = common.JobWorkerDir + register.localIP

		cancelFunc = nil

		// 创建租约
		leaseGrantResp, err = register.lease.Grant(context.TODO(), 10)
		if err != nil {
			goto RETRY
		}

		// 自动续约
		keepAliveChan, err = register.lease.KeepAlive(context.TODO(), leaseGrantResp.ID)
		if err != nil {
			goto RETRY
		}

		cancelCtx, cancelFunc = context.WithCancel(context.TODO())

		// 注册到etcd
		_, err = register.kv.Put(cancelCtx, regKey, "", clientv3.WithLease(leaseGrantResp.ID))
		if err != nil {
			goto RETRY
		}

		// 处理续约应答
		for {
			select {
			case keepAliveResp = <-keepAliveChan:
				if keepAliveResp == nil { // 续约失败
					goto RETRY
				}
			}
		}

	RETRY:
		time.Sleep(1 * time.Second)
		if cancelFunc != nil {
			cancelFunc()
		}
	}
}

// 初始化
func InitRegister() error {
	config := clientv3.Config{
		Endpoints:   GConfig.EtcdEndpoints,                                     // 集群地址
		DialTimeout: time.Duration(GConfig.EtcdDialTimeout) * time.Millisecond, // 连接超时
	}

	// 建立连接
	client, err := clientv3.New(config)
	if err != nil {
		return err
	}

	// 本机IP
	localIP, err := getLocalIP()
	if err != nil {
		return err
	}

	// 得到kv和lease的API子集
	kv := clientv3.NewKV(client)
	lease := clientv3.NewLease(client)

	GRegister = &Register{
		client:  client,
		kv:      kv,
		lease:   lease,
		localIP: localIP,
	}

	// 服务注册
	go GRegister.keepOnline()

	return nil
}
