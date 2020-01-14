package main

import (
	sb "github.com/samtech09/gosql"
)

func main() {
	fw := sb.NewFileWriter(5)

	stmt := sb.InsertBuilder().Table("users").
		Columns("name", "age").Returning("id").
		Build(true)
	// queue for writing
	fw.Queue(stmt, "user", "Create", "Create new user.")

	stmt = sb.SelectBuilder().Select("q.ID", "qd.Title").
		From("Questions", "q").
		From("QuestionData", "qd").
		Where(sb.C().EQ("q.ID", "qd.QID"), sb.C().EQ("q.TopicID", "21")).
		OrderBy("qd.QID", true).
		RowCount().
		Build(true)
	// queue for writing
	fw.Queue(stmt, "ques", "listForDD", "Gives list of question ID and Title only to fill dropdowns.")

	// export as GO code to ../sqls folder
	//  exported filename = sqlbuilder
	//  exported gocode package = sqls
	fw.Write("../sqls", "sqlbuilder", "sqls", sb.WriteGoCode)
}
