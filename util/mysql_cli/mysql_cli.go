package mysql_cli

import (
	"database/sql"
	_ "go-sql-driver/mysql"

	"github.com/gochenzl/chess/util/log"
)

var db *sql.DB
var sqlChan chan string

func Init(dsn string, maxOpenConns int) bool {
	var err error
	db, err = sql.Open("mysql", dsn)
	if err != nil {
		log.Error("open mysql fail:", err)
		return false
	}

	db.SetMaxIdleConns(maxOpenConns)
	db.SetMaxOpenConns(maxOpenConns)

	sqlChan = make(chan string, 500)
	for i := 0; i < 5; i++ {
		go process()
	}

	return true
}

func Exec(sqlStr string) bool {
	_, err := db.Exec(sqlStr)
	if err != nil {
		log.Error("sql = %s, err = %s", sqlStr, err.Error())
		return false
	}

	return true
}

// 异步执行
func AsyncExec(sqlStr string) {
	sqlChan <- sqlStr
}

func process() {
	for {
		sqlStr := <-sqlChan
		Exec(sqlStr)
	}
}
