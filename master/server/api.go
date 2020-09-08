package server

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/gorhill/cronexpr"

	common "cronjob/common"
)

var (
	GAPI *gin.Engine
)

func InitAPIServer() {

	GAPI = gin.Default()

	// 中间件：跨域
	GAPI.Use(Cors())

	GAPI.POST("/job/add", handleJobAdd)
	GAPI.POST("/job/save", handleJobSave)
	GAPI.POST("/job/delete", handleJobDelete)
	GAPI.GET("/job/list", handleJobList)
	GAPI.POST("/job/kill", handleJobKill)
	GAPI.GET("/job/log", handleJobLog)
	GAPI.GET("/worker/list", handleWorkerList)

	// GAPI.GET("/test/test", handlerTest)

	return
}

func handlerTest(c *gin.Context) {

}

// 跨域请求
func Cors() gin.HandlerFunc {
	return func(c *gin.Context) {
		method := c.Request.Method

		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Headers", "Content-Type,AccessToken,X-CSRF-Token, Authorization, Token")
		c.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		c.Header("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers, Content-Type")
		c.Header("Access-Control-Allow-Credentials", "true")

		//放行所有OPTIONS方法
		if method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
		}
		// 处理请求
		c.Next()
	}
}

// json返回：默认
func ResponseDefault(c *gin.Context, retcode int, msg string) {
	c.JSON(200, gin.H{
		"retcode": retcode,
		"msg":     msg,
	})
}

// json返回：数据
func ResponseByData(c *gin.Context, retcode int, msg string, data interface{}) {
	retData := gin.H{}
	retData["retcode"] = retcode
	retData["msg"] = msg
	if data != nil {
		retData["data"] = data
	}

	c.JSON(200, retData)
}

// /job/add 添加任务
func handleJobAdd(c *gin.Context) {
	// post 参数
	postJob := c.DefaultPostForm("job", "")
	if postJob == "" {
		ResponseDefault(c, -1, "postform field is empty")
		return
	}

	jobObj := common.Job{}
	// json 解析
	err := json.Unmarshal([]byte(postJob), &jobObj)
	if err != nil {
		ResponseDefault(c, -1, "postform field error")
		return
	}

	// 检测当前任务是否存在？
	if GJobMgr.CheckJobExist(jobObj.Name) == false {
		ResponseDefault(c, -1, "当前任务已存在")
		return
	}

	// 检测cronexpr
	if err := checkCronExpr(jobObj.CronExpr); err != nil {
		ResponseDefault(c, -1, "cronexpr格式错误")
		return
	}

	// save job
	oldJob, err := GJobMgr.SaveJob(&jobObj)
	if err != nil {
		ResponseDefault(c, -1, "postform field error")
		return
	}

	ResponseByData(c, 0, "", oldJob)
	return

}

// /job/save 保存任务
func handleJobSave(c *gin.Context) {
	// post 参数
	postJob := c.DefaultPostForm("job", "")
	if postJob == "" {
		ResponseDefault(c, -1, "postform field is empty")
		return
	}

	jobObj := common.Job{}
	// json 解析
	err := json.Unmarshal([]byte(postJob), &jobObj)
	if err != nil {
		ResponseDefault(c, -1, "postform field error")
		return
	}

	// 检测cronexpr
	if err := checkCronExpr(jobObj.CronExpr); err != nil {
		ResponseDefault(c, -1, "cronexpr格式错误")
		return
	}

	// save job
	oldJob, err := GJobMgr.SaveJob(&jobObj)
	if err != nil {
		ResponseDefault(c, -1, "postform field error")
		return
	}

	ResponseByData(c, 0, "", oldJob)
	return
}

// /job/delete 删除任务
func handleJobDelete(c *gin.Context) {
	// post 参数
	postJob := c.DefaultPostForm("name", "")
	if postJob == "" {
		ResponseDefault(c, -1, "postform field is empty")
		return
	}

	// 删除
	oldJob, err := GJobMgr.DeleteJob(postJob)
	if err != nil {
		ResponseDefault(c, -1, "postform field error")
		return
	}
	ResponseByData(c, 0, "", oldJob)
	return
}

// /job/list 获取任务列表
func handleJobList(c *gin.Context) {
	// 获取 列表
	listJob, err := GJobMgr.ListJobs()
	if err != nil {
		ResponseDefault(c, -1, "get list job error")
		return
	}
	ResponseByData(c, 0, "", listJob)
	return
}

// /job/kill 强杀任务
func handleJobKill(c *gin.Context) {
	// post 参数
	postJob := c.DefaultPostForm("name", "")
	if postJob == "" {
		ResponseDefault(c, -1, "postform field is empty")
		return
	}

	err := GJobMgr.KillJob(postJob)
	if err != nil {
		ResponseDefault(c, -1, "kill job error")
		return
	}
	ResponseDefault(c, 0, "")
	return
}

// /job/log 获取任务日志
func handleJobLog(c *gin.Context) {
	// post 参数
	name := c.DefaultQuery("name", "")
	if name == "" {
		ResponseDefault(c, -1, "postform field is empty")
		return
	}
	limitStr := c.DefaultQuery("limit", "")
	limit := 10
	if limitStr != "" {
		if limitNum, err := strconv.Atoi(limitStr); err == nil {
			limit = limitNum
		}
	}

	logs, err := GLogMgr.ListLog(name, limit)
	if err != nil {
		ResponseDefault(c, -1, "get log list error")
		return
	}

	ResponseByData(c, 0, "", logs)
}

// /worker/list 获取节点信息
func handleWorkerList(c *gin.Context) {
	workerArr, err := GWorkerMgr.ListWorkers()
	if err != nil {
		ResponseDefault(c, -1, "get worker list error")
		return
	}

	ResponseByData(c, 0, "", workerArr)
	return
}

// 检测cronexpr是否合法？
func checkCronExpr(cron string) error {
	if _, err := cronexpr.Parse(cron); err != nil {
		return err
	}
	return nil
}
