//sqls package is autogenerated by gosql.
//Do not edit this file.
//Changes will be lost on next auto-generate.
package sqls

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

var _jsonsqls map[string]string

type Statement struct {
	ReadOnly  bool
	SQL       string
}

func fatal(msg ...string) {
	fmt.Println(msg)
	os.Exit(1)
}

//LoadSQLs load sql into map from json file generated by gosql
func LoadSQLs(f string) {
	file, err := os.Open(f)
	if err != nil {
		fatal("error reading file ", f, " ", err.Error())
	}
	defer file.Close()

	byteValue, err := ioutil.ReadAll(file)
	if err != nil {
		fatal("read failure ", err.Error())
	}

	err = json.Unmarshal([]byte(byteValue), &_jsonsqls)
	if err != nil {
		fatal("parse failure ", err.Error())
	}
}

func parseStmt(s string) Statement {
	stmt := Statement{}
	ro := s[:1]
	if ro == "T" {
		stmt.ReadOnly = true
	}
	stmt.SQL = s[2:len(s)]
	return stmt
}

//Creates new user.
//
//Fields: 2, Parameters: 2
//
//  Fields: name, age
//
//  ParamFields: name, age
//
//  ReturningFields: id
//SQL:
//  insert into users(name, age) values($1, $2) returning id;
func UserCreate() Statement {
	sql, ok := _jsonsqls["UserCreate"]
	if !ok {
		return Statement{}
	}
	return parseStmt(sql)
}

//Gives list of question ID and Title only to fill dropdowns.
//
//Fields: 3, Parameters: 0
//
//  Fields: q.ID, qd.Title, rowcount
//SQL:
//  select q.ID, qd.Title, count(*) over() as rowcount from questions q, questiondata qd where (q.ID=qd.QID and q.TopicID=21) order by qd.QID desc;
func QuesListForDD() Statement {
	sql, ok := _jsonsqls["QuesListForDD"]
	if !ok {
		return Statement{}
	}
	return parseStmt(sql)
}

