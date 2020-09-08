package server

import (
	"cronjob/common"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

type LogMgr struct {
	db *sql.DB
}

var (
	GLogMgr *LogMgr
)

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
		db: db,
	}
	return nil
}

func (logMgr *LogMgr) ListLog(name string, limit int) ([]common.JobLog, error) {

	logs := make([]common.JobLog, 0)
	rows, err := logMgr.db.Query("SELECT * FROM `cronjob_log` WHERE `job_name`=? ORDER BY `id` DESC limit ?", name, limit)
	if err != nil {
		return logs, err
	}

	var log common.JobLog
	for rows.Next() {
		// log := common.JobLog{}
		rows.Scan(&log.ID, &log.JobName, &log.Command, &log.Err, &log.Output,
			&log.PlanTime, &log.ScheduleTime, &log.StartTime, &log.EndTime)
		logs = append(logs, log)
	}
	fmt.Printf("Logs: \n%+v\n", logs)
	return logs, nil
}
