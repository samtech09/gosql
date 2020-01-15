package main

import (
	"fmt"

	"github.com/samtech09/gosql/Examples/JsonLoader/sqls"
)

func main() {
	//load sqls from json file
	sqls.LoadSQLs("sqls/sqlbuilder.json")

	// Execute sql
	ExecuteQuery(sqls.UserCreate(), "Testuser", "22")

	ExecuteQuery(sqls.QuesListForDD(), nil)
}

func ExecuteQuery(stmt sqls.Statement, param ...interface{}) {
	// code to execute SQL
	// ...
	fmt.Printf("ReadOnly: %v\nSQL: %s\n", stmt.ReadOnly, stmt.SQL)
}
