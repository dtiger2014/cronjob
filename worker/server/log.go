package server

import (
	"cronjob/common"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

type LogMgr struct {
	db      *sql.DB
	logChan chan *common.JobLog
}

var (
	GLogMgr *LogMgr
)

func (logMgr *LogMgr) writeLoop() {

	for {
		select {
		case log := <-logMgr.logChan:
			logMgr.addLogToDB(log)
		}
	}
}

func InitLogMgr() error {
	dbDSN := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=%s",
		GConfig.MysqlUser, GConfig.MysqlPass, GConfig.MysqlHost,
		GConfig.MysqlPort, GConfig.MysqlDatabase, GConfig.MysqlCharset)

	// 打开连接失败
	db, err := sql.Open("mysql", dbDSN)
	if err != nil {
		return err
	}

	// 最大连接数
	db.SetMaxOpenConns(100)
	// 闲置连接数
	db.SetMaxIdleConns(20)
	// 最大连接周期
	db.SetConnMaxLifetime(100 * time.Second)

	GLogMgr = &LogMgr{
		db:      db,
		logChan: make(chan *common.JobLog, 1000),
	}

	// 启动携程
	go GLogMgr.writeLoop()

	return nil
}

func (logmgr *LogMgr) Append(jobLog *common.JobLog) {
	select {
	case logmgr.logChan <- jobLog:
	default:
		// 队列满了就丢弃
	}
}

func (logmgr *LogMgr) addLogToDB(job *common.JobLog) {
	ret, _ := logmgr.db.Exec("INSERT INTO cronjob_log(job_name,command,err,output,plan_time,schedule_time,start_time,end_time) values(?,?,?,?,?,?,?,?)",
		&job.JobName, &job.Command, &job.Err, &job.Output,
		&job.PlanTime, &job.ScheduleTime, &job.StartTime, &job.EndTime)

	//插入数据的主键id
	lastInsertID, _ := ret.LastInsertId()
	fmt.Println("addLogToDB LastInsertID:", lastInsertID)

	//影响行数
	rowsaffected, _ := ret.RowsAffected()
	fmt.Println("addLogToDB RowsAffected:", rowsaffected)
}
