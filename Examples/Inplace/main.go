package main

import (
	"fmt"

	gs "github.com/samtech09/gosql"
)

func main() {
	// build sql
	stmt := gs.SelectBuilder().Select("q.ID", "qd.Title").
		From("Questions", "q").
		From("QuestionData", "qd").
		Where(gs.C().EQ("q.ID", "qd.QID"),
			gs.C().EQ("q.TopicID", "$1")).
		OrderBy("qd.QID", true).
		RowCount().
		Build(true)

	// Execute sql
	ExecuteQuery(stmt.SQL, "Testuser", "22")
}

func ExecuteQuery(sql string, param ...interface{}) {
	// code to execute SQL
	//
	fmt.Println(sql)
}
