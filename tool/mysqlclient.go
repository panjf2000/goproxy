/*
@version: 1.0
@author: allanpan
@license:  Apache Licence
@contact: panjf2000@gmail.com  
@site: 
@file: mysqlclient.go
@time: 2017/3/20 19:39
@tag: 1,2,3
@todo: ...
*/
package tool

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/panjf2000/goproxy/config"
)

var db *sql.DB
var myErr error

func init() {
	db, myErr = sql.Open("mysql", fmt.Sprintf(config.MYSQL_TPL,
		config.MysqlConf[config.ENV]["MYSQL_USER"],
		config.MysqlConf[config.ENV]["MYSQL_PASSWD"],
		config.MysqlConf[config.ENV]["MYSQL_MUSIC_HOST"],
		config.MysqlConf[config.ENV]["MYSQL_DB_NAME"]))
	db.SetMaxOpenConns(1000)
	db.SetMaxIdleConns(500)
	db.Ping()

}

//取一行数据，注意这类取出来的结果都是string
func fetchRow(db *sql.DB, sqlstr string, args ...interface{}) (*map[string]string, error) {
	stmtOut, err := db.Prepare(sqlstr)
	if err != nil {
		panic(err.Error())
	}
	defer stmtOut.Close()

	rows, err := stmtOut.Query(args...)
	if err != nil {
		panic(err.Error())
	}

	columns, err := rows.Columns()
	if err != nil {
		panic(err.Error())
	}

	values := make([]sql.RawBytes, len(columns))
	scanArgs := make([]interface{}, len(values))
	ret := make(map[string]string, len(scanArgs))

	for i := range values {
		scanArgs[i] = &values[i]
	}
	for rows.Next() {
		err = rows.Scan(scanArgs...)
		if err != nil {
			panic(err.Error())
		}
		var value string

		for i, col := range values {
			if col == nil {
				value = "NULL"
			} else {
				value = string(col)
			}
			ret[columns[i]] = value
		}
		break //get the first row only
	}
	return &ret, nil
}

//取多行，<span style="font-family: Arial, Helvetica, sans-serif;">注意这类取出来的结果都是string </span>
func fetchRows(db *sql.DB, sqlstr string, args ...interface{}) (*[]map[string]string, error) {
	stmtOut, err := db.Prepare(sqlstr)
	if err != nil {
		panic(err.Error())
	}
	defer stmtOut.Close()

	rows, err := stmtOut.Query(args...)
	if err != nil {
		panic(err.Error())
	}

	columns, err := rows.Columns()
	if err != nil {
		panic(err.Error())
	}

	values := make([]sql.RawBytes, len(columns))
	scanArgs := make([]interface{}, len(values))

	ret := make([]map[string]string, 0)
	for i := range values {
		scanArgs[i] = &values[i]
	}

	for rows.Next() {
		err = rows.Scan(scanArgs...)
		if err != nil {
			panic(err.Error())
		}
		var value string
		vmap := make(map[string]string, len(scanArgs))
		for i, col := range values {
			if col == nil {
				value = "NULL"
			} else {
				value = string(col)
			}
			vmap[columns[i]] = value
		}
		ret = append(ret, vmap)
	}
	return &ret, nil
}

func NewMysqlClient() (*sql.DB, error) {
	return db, myErr
}
