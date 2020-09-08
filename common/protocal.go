package common

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/gorhill/cronexpr"
)

// 定时任务
type Job struct {
	Name     string `json:"name"`
	Command  string `json:"command"`  // shell command
	CronExpr string `json:"cronExpr"` // cron expression
}

// 任务调度计划
type JobSchedulePlan struct {
	Job      *Job                 // 要调度的任务信息
	Expr     *cronexpr.Expression // 解析好的cronexpr表达式
	NextTime time.Time            // 下次调度时间
}

// 任务执行状态
type JobExecuteInfo struct {
	Job        *Job               // 任务信息
	PlanTime   time.Time          // 理论上的调度时间
	RealTime   time.Time          // 实际的调度时间
	CancelCtx  context.Context    // 任务command的context
	CancelFunc context.CancelFunc // 用于取消command执行的cancel函数
}

// 变化事件
type JobEvent struct {
	EventType int // SAVE, DELETE
	Job       *Job
}

type JobLog struct {
	ID           int    `db:"id"`
	JobName      string `db:"job_name" json:"job_name"`           // job name
	Command      string `db:"command" json:"command"`             // job command
	Err          string `db:"err" json:"err"`                     // error string
	Output       string `db:"output" json:"output"`               // output string
	PlanTime     int64  `db:"plan_time" json:"plan_time"`         // job plan time
	ScheduleTime int64  `db:"schedule_time" json:"schedule_time"` // job schedule time
	StartTime    int64  `db:"start_time" json:"start_time"`       // job start time
	EndTime      int64  `db:"end_time" json:"end_time"`           // job end time
}

// 任务执行结果
type JobExecuteResult struct {
	ExecuteInfo *JobExecuteInfo // 执行状态
	Output      []byte          // 输出
	Err         error           // 错误
	StartTime   time.Time       // 启动时间
	EndTime     time.Time       // 结束时间
}

func UnpackJob(value []byte) (*Job, error) {
	job := &Job{}

	err := json.Unmarshal(value, job)
	if err != nil {
		return nil, err
	}
	return job, nil
}

// 从etcd的key中提取任务名称
// /cron/jobs/xxx -> xxx
func ExtractJobName(jobKey string) string {
	return strings.TrimPrefix(jobKey, JobSaveDir)
}

// /cron/killer/xxx -> xxx
func ExtractKillerName(killerKey string) string {
	return strings.TrimPrefix(killerKey, JobKillerDir)
}

// 提取worker的IP
func ExtractWorkerIP(regKey string) string {
	return strings.TrimPrefix(regKey, JobWorkerDir)
}

// 任务变化事件有2种：1）更新任务 2）删除任务
func BuildJobEvent(eventType int, job *Job) *JobEvent {
	return &JobEvent{
		EventType: eventType,
		Job:       job,
	}
}

// 构造任务执行计划
func BuildJobSchedulePlan(job *Job) (*JobSchedulePlan, error) {
	// 解析Job的cron表达式
	expr, err := cronexpr.Parse(job.CronExpr)
	if err != nil {
		return nil, err
	}

	// 生成任务调度计划对象
	jobSchedulePlan := &JobSchedulePlan{
		Job:      job,
		Expr:     expr,
		NextTime: expr.Next(time.Now()),
	}

	return jobSchedulePlan, nil
}

// 构造任务执行状态信息
func BuildJobExecuteInfo(jobSchedulePlan *JobSchedulePlan) *JobExecuteInfo {
	jobExecuteInfo := &JobExecuteInfo{
		Job:      jobSchedulePlan.Job,
		PlanTime: jobSchedulePlan.NextTime,
		RealTime: time.Now(),
	}

	jobExecuteInfo.CancelCtx, jobExecuteInfo.CancelFunc = context.WithCancel(context.TODO())
	return jobExecuteInfo
}
