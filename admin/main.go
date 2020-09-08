package main

import (
	"cronjob/admin/server"
	"fmt"
	"runtime"
)

const (
	confilePaht = "etc/admin.json"
)

// initEnv : init thread
func initEnv() error {
	runtime.GOMAXPROCS(runtime.NumCPU())

	return nil
}

func main() {
	var err error

	// 加载配置文件
	if err = server.InitConfig(confilePaht); err != nil {
		panic(err)
	}

	fmt.Println(server.GConfig.WebRoot)

	// init Env
	if err = initEnv(); err != nil {
		panic(err)
	}

	// API 初始化
	server.InitAPIServer()

	// 启动API服务
	port := fmt.Sprintf(":%d", server.GConfig.Port)
	server.GAPI.Run(port)
}
