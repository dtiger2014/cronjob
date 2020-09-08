package main

import (
	"fmt"
	"time"

	"github.com/gorhill/cronexpr"
)

// https://github.com/gorhill/cronexpr

/*
Field name     Mandatory?   Allowed values    Allowed special characters
----------     ----------   --------------    --------------------------
Seconds        No           0-59              * / , -
Minutes        Yes          0-59              * / , -
Hours          Yes          0-23              * / , -
Day of month   Yes          1-31              * / , - L W
Month          Yes          1-12 or JAN-DEC   * / , -
Day of week    Yes          0-6 or SUN-SAT    * / , - L #
Year           No           1970–2099         * / , -
*/

func main() {
	cron := "* * * * * *"

	// 检测 expr 是否有效？
	if _, err := cronexpr.Parse(cron); err != nil {
		fmt.Println("Err: ", err)
	}

	expr := cronexpr.MustParse(cron)
	nextTime := expr.Next(time.Now())
	fmt.Println(nextTime.Format("2006-01-02 15:04:05"))
	nextTimes := expr.NextN(time.Now(), 5)
	for i := range nextTimes {
		fmt.Println(nextTimes[i].Format(time.RFC1123))
	}
}
