package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

var MysqlDb *sql.DB
var MysqlDbErr error

const (
	USER_NAME = "root"
	PASS_WORD = "123456"
	HOST      = "192.168.74.121"
	PORT      = "3306"
	DATABASE  = "test"
	CHARSET   = "utf8"
)

type Test struct {
	ID   int    `db:"id" json:"id"`
	Name string `db:"name" json:"name"`
}

func main() {
	dbDSN := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=%s", USER_NAME, PASS_WORD, HOST, PORT, DATABASE, CHARSET)

	// 打开连接失败
	MysqlDb, MysqlDbErr = sql.Open("mysql", dbDSN)
	//defer MysqlDb.Close();
	if MysqlDbErr != nil {
		log.Println("dbDSN: " + dbDSN)
		panic("数据源配置不正确: " + MysqlDbErr.Error())
	}

	// 最大连接数
	MysqlDb.SetMaxOpenConns(100)
	// 闲置连接数
	MysqlDb.SetMaxIdleConns(20)
	// 最大连接周期
	MysqlDb.SetConnMaxLifetime(100 * time.Second)

	if MysqlDbErr = MysqlDb.Ping(); nil != MysqlDbErr {
		panic("数据库链接失败: " + MysqlDbErr.Error())
	}

	{
		test := Test{}
		row := MysqlDb.QueryRow("select id, name from `test` where id=?", 1)
		err := row.Scan(&test.ID, &test.Name)
		if err != nil {
			fmt.Printf("scan failed, err:%v", err)
			return
		}
		fmt.Printf("%+v\n", test)
	}

	{
		// 通过切片存储
		tests := make([]Test, 0)
		rows, _ := MysqlDb.Query("SELECT * FROM `test` limit ?", 100)
		// 遍历
		var test Test
		for rows.Next() {
			rows.Scan(&test.ID, &test.Name)
			tests = append(tests, test)
		}
		fmt.Printf("%+v\n", tests)

		res, err := json.Marshal(tests)
		fmt.Println(err)
		fmt.Println(string(res))
	}
}
