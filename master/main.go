package main

import (
	"cronjob/master/server"
	"fmt"
	"runtime"
)

const (
	confilePath = "etc/master.json"
)

// initEnv : init thread
func initEnv() error {
	runtime.GOMAXPROCS(runtime.NumCPU())
	return nil
}

func main() {
	var err error

	// 加载配置文件
	if err = server.InitConfig(confilePath); err != nil {
		panic(err)
	}

	fmt.Println(server.GConfig)

	// init Env
	if err = initEnv(); err != nil {
		panic(err)
	}

	// 初始化 JobMgr
	if err = server.InitJobMgr(); err != nil {
		panic(err)
	}

	// 初始化 workermgr
	if err = server.InitWorkerMgr(); err != nil {
		panic(err)
	}

	// 初始化 logMgr
	if err = server.InitLogMgr(); err != nil {
		panic(err)
	}

	// API 初始化
	server.InitAPIServer()

	// 启动API服务
	port := fmt.Sprintf(":%d", server.GConfig.Port)
	server.GAPI.Run(port)
}
