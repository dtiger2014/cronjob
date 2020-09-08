package main

import (
	"cronjob/worker/server"
	"runtime"
	"time"
)

const (
	confilePaht = "etc/worker.json"
)

func initEnv() {
	runtime.GOMAXPROCS(runtime.NumCPU())
}

func main() {
	var err error

	// 初始化 env
	initEnv()

	// 配置文件
	err = server.InitConfig(confilePaht)
	if err != nil {
		panic(err)
	}

	// 服务注册
	err = server.InitRegister()
	if err != nil {
		panic(err)
	}

	// 执行器
	err = server.InitExecutor()
	if err != nil {
		panic(err)
	}

	// 日志 logMgr
	err = server.InitLogMgr()
	if err != nil {
		panic(err)
	}

	// 调度器
	err = server.InitScheduler()
	if err != nil {
		panic(err)
	}

	// 初始化任务管理器
	err = server.InitJobMgr()
	if err != nil {
		panic(err)
	}

	for {
		time.Sleep(1 * time.Second)
	}
}
