package server

import (
	"cronjob/common"
	"errors"
	"fmt"
	"time"
)

// 任务调度
type Scheduler struct {
	jobEventChan      chan *common.JobEvent              // etcd任务事件队列
	jobPlanTable      map[string]*common.JobSchedulePlan // 任务调度计划表
	jobExecutingTable map[string]*common.JobExecuteInfo  // 任务执行表
	jobResultChan     chan *common.JobExecuteResult      // 任务结果队列
}

var (
	GScheduler *Scheduler
)

// 推送任务变化事件
func (scheduler *Scheduler) PushJobEvent(jobEvent *common.JobEvent) {
	scheduler.jobEventChan <- jobEvent
}

// 回传任务执行结果
func (scheduler *Scheduler) PushJobResult(jobResult *common.JobExecuteResult) {
	scheduler.jobResultChan <- jobResult
}

// 初始化调度器
func InitScheduler() error {
	GScheduler = &Scheduler{
		jobEventChan:      make(chan *common.JobEvent, 1000),
		jobPlanTable:      make(map[string]*common.JobSchedulePlan),
		jobExecutingTable: make(map[string]*common.JobExecuteInfo),
		jobResultChan:     make(chan *common.JobExecuteResult, 1000),
	}

	// 启动调度协程
	go GScheduler.scheduleLoop()

	return nil
}

// 调度协程
func (scheduler *Scheduler) scheduleLoop() {

	// 初始化一次（1秒）
	scheduleAfter := scheduler.TrySchedule()

	// 调度的延迟定时器
	scheduleTimer := time.NewTimer(scheduleAfter)

	// 定时任务common.Job
	for {
		select {
		case jobEvent := <-scheduler.jobEventChan: // 监听任务变化事件
			// 对内存中维护的任务列表做增删改查
			scheduler.handleJobEvent(jobEvent)
		case <-scheduleTimer.C: // 最近的任务到期了
		case jobResult := <-scheduler.jobResultChan: // 监听任务执行结果
			scheduler.handleJobResult(jobResult)
		}
		// 调度一次任务
		scheduleAfter = scheduler.TrySchedule()
		// 重置调度间隔
		scheduleTimer.Reset(scheduleAfter)
	}
}

func (scheduler *Scheduler) TrySchedule() time.Duration {

	// 如果任务表为空的话，随便睡眠多久
	if len(scheduler.jobPlanTable) == 0 {
		return 1 * time.Second
	}

	// 当前时间
	now := time.Now()

	var nearTime *time.Time

	// 便利所有任务
	for _, jobPlan := range scheduler.jobPlanTable {
		if jobPlan.NextTime.Before(now) || jobPlan.NextTime.Equal(now) {
			scheduler.TryStartJob(jobPlan)
			jobPlan.NextTime = jobPlan.Expr.Next(now) // 更新下次执行时间
		}

		// 统计最近一个要过期的任务时间
		if nearTime == nil || jobPlan.NextTime.Before(*nearTime) {
			nearTime = &jobPlan.NextTime
		}
	}

	// 返回 下次调度间隔（最近要执行的任务调度时间 - 当前时间）
	return (*nearTime).Sub(now)
}

// 尝试执行任务
func (scheduler *Scheduler) TryStartJob(jobPlan *common.JobSchedulePlan) {
	// 执行的任务可能运行很久, 1分钟会调度60次，但是只能执行1次, 防止并发！

	// 如果任务正在执行，跳过本次调度
	if _, jobExecuting := scheduler.jobExecutingTable[jobPlan.Job.Name]; jobExecuting {
		fmt.Println("尚未结束，跳过执行：", jobPlan.Job.Name)
		return
	}

	// 构建执行状态信息
	jobExecuteInfo := common.BuildJobExecuteInfo(jobPlan)

	// 保存执行状态
	scheduler.jobExecutingTable[jobPlan.Job.Name] = jobExecuteInfo

	// 执行任务
	fmt.Println("执行任务：", jobExecuteInfo.Job.Name, jobExecuteInfo.PlanTime, jobExecuteInfo.RealTime)
	GExecutor.ExecuteJob(jobExecuteInfo)
}

// 处理任务事件
func (scheduler *Scheduler) handleJobEvent(jobEvent *common.JobEvent) {
	switch jobEvent.EventType {
	case common.JobEventSave: // 保存任务事件
		jobSchedulePlan, err := common.BuildJobSchedulePlan(jobEvent.Job)
		if err != nil {
			return
		}
		scheduler.jobPlanTable[jobEvent.Job.Name] = jobSchedulePlan
	case common.JobEventDelete: // 删除任务事件
		if _, jobExecuting := scheduler.jobExecutingTable[jobEvent.Job.Name]; jobExecuting {
			delete(scheduler.jobPlanTable, jobEvent.Job.Name)
		}
	case common.JobEventKill: // 强杀任务事件
		// 取消掉Command执行，判断任务是否在执行中
		if jobExecuteInfo, jobExecuting := scheduler.jobExecutingTable[jobEvent.Job.Name]; jobExecuting {
			jobExecuteInfo.CancelFunc() // 触发command杀死shell子进程，任务得到退出
		}
	}
}

// 处理任务结果
func (scheduler *Scheduler) handleJobResult(result *common.JobExecuteResult) {
	// 删除执行状态
	delete(scheduler.jobExecutingTable, result.ExecuteInfo.Job.Name)

	// 生成执行日志
	if result.Err != errors.New("err lock already required") {
		jobLog := &common.JobLog{
			JobName:      result.ExecuteInfo.Job.Name,
			Command:      result.ExecuteInfo.Job.Command,
			Output:       string(result.Output),
			PlanTime:     result.ExecuteInfo.PlanTime.UnixNano() / 1000 / 1000,
			ScheduleTime: result.ExecuteInfo.RealTime.UnixNano() / 1000 / 1000,
			StartTime:    result.StartTime.UnixNano() / 1000 / 1000,
			EndTime:      result.EndTime.UnixNano() / 1000 / 1000,
		}
		if result.Err != nil {
			jobLog.Err = result.Err.Error()
		} else {
			jobLog.Err = ""
		}
		GLogMgr.Append(jobLog)
	}
	// fmt.Println("任务执行完成:", result.ExecuteInfo.Job.Name, string(result.Output), result.Err)
}
