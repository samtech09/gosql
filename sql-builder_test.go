package gosql

import (
	"fmt"
	"os"
	"testing"
)

func TestWhereClauseMsSQL(t *testing.T) {
	fmt.Println("\n\nTestWhereClause ***")

	os.Setenv("DATABASE_TYPE", DbTypeMsSQL)

	sql := SelectBuilder().
		Where(C().EQ("q.ID", "qd.QID"), C().EQ("q.TopicID", "?")).
		WhereGroup(OpOR, C().LT("qd.Seqno", "?")).
		BuildWhereClause()

	exp := "where (q.ID=qd.QID and q.TopicID=@p1) OR (qd.Seqno<@p2)"
	if sql != exp {
		t.Errorf("Expected\n %s\nGot\n %s", exp, sql)
	}
}

func TestWhereClauseMySQL(t *testing.T) {
	fmt.Println("\n\nTestWhereClause ***")

	os.Setenv("DATABASE_TYPE", DbTypeMySQL)

	sql := SelectBuilder().
		Where(C().EQ("q.ID", "qd.QID"), C().EQ("q.TopicID", "?")).
		WhereGroup(OpOR, C().LT("qd.Seqno", "?")).
		BuildWhereClause()

	exp := "where (q.ID=qd.QID and q.TopicID=?) OR (qd.Seqno<?)"
	if sql != exp {
		t.Errorf("Expected\n %s\nGot\n %s", exp, sql)
	}
}

func TestWhereClausePgSQL(t *testing.T) {
	fmt.Println("\n\nTestWhereClause ***")

	os.Setenv("DATABASE_TYPE", DbTypePostgreSQL)

	sql := SelectBuilder().
		Where(C().EQ("q.ID", "qd.QID"), C().EQ("q.TopicID", "?")).
		WhereGroup(OpOR, C().LT("qd.Seqno", "?")).
		BuildWhereClause()

	exp := "where (q.ID=qd.QID and q.TopicID=$1) OR (qd.Seqno<$2)"
	if sql != exp {
		t.Errorf("Expected\n %s\nGot\n %s", exp, sql)
	}
}

func TestBuilder(t *testing.T) {
	fmt.Println("\n\nTestBuilder ***")

	//builder := selectBuilder{}

	stmt := SelectBuilder().Select("q.ID", "qd.QID").
		From("Questions", "q").
		From("QuestionData", "qd").
		Where(C().EQ("q.ID", "qd.QID"), C().EQ("q.TopicID", "?")).
		OrderBy("qd.QID", true).
		RowCount().
		Build(true)

	// fmt.Println("Paracount: ", stmt.ParamCount)
	// fmt.Println("ParaFields: ", stmt.ParamFields)
	// fmt.Println("FieldsCount: ", stmt.FieldsCount)
	// fmt.Println("Fields: ", stmt.Fields)

	if stmt.ParamCount != 1 {
		t.Errorf("Expected Paramters\n %d\nGot\n %d", 1, stmt.ParamCount)
	}

	exp := "select q.ID, qd.QID, count(*) over() as rowscount from questions q, questiondata qd where (q.ID=qd.QID and q.TopicID=$1) order by qd.QID desc;"
	if stmt.SQL != exp {
		//fmt.Printf("Sql: %d, Exp: %d\n", len(stmt.SQL), len(exp))
		t.Errorf("Expected\n %s\nGot\n %s", exp, stmt.SQL)
	}
}

func TestSubSQL(t *testing.T) {
	fmt.Println("\n\nTestSubSQL ***")

	stmt := SelectBuilder().
		Select("tq.ID as QID").
		Sub(SelectBuilder().
			Select("left(Qdata,50)").From("QuestionData", "").
			Where(C().EQ("QID", "q.ID"), C().EQ("DataType", "1")).
			Limit(1), "as Tquestion").
		Sub(SelectBuilder().
			Select("s.Title").From("Subjects", "s").
			Where(C().EQ("s.ID", "t.SubjectID")), "as TSubject").
		Select("q.QType", "q.DifficultyLevel", "tq.CorrectMarks", "tq.NegativeMarks", "tq.QCancelMarks", "tq.seqno").
		Select("q.ID", "t.SubjectID", "ts.SeqNo AS SeqNoSubject", "tq.Addedon", "tq.Addedby", "getquestionlanguages(q.ID) as Languages").
		From("TestQuestions", "tq").
		From("Questions", "q").
		From("Topics", "t").
		From("testsubjects", "ts").
		Where(C().EQ("ts.TestID", "tq.TestID"), C().EQ("t.SubjectID", "ts.SubjectID"),
			C().EQ("t.ID", "q.TopicID"), C().EQ("tq.QID", "q.ID")).
		Build(true)

	exp := "select tq.ID as QID, (select left(Qdata,50) from questiondata where (DataType=1 and QID=q.ID) limit 1) as Tquestion, " +
		"(select s.Title from subjects s where (s.ID=t.SubjectID)) as TSubject, q.QType, q.DifficultyLevel, " +
		"tq.CorrectMarks, tq.NegativeMarks, tq.QCancelMarks, tq.seqno, q.ID, t.SubjectID, ts.SeqNo AS SeqNoSubject, tq.Addedon, tq.Addedby, " +
		"getquestionlanguages(q.ID) as Languages " +
		"from questions q, topics t, testquestions tq, testsubjects ts " +
		"where (t.ID=q.TopicID and t.SubjectID=ts.SubjectID and tq.QID=q.ID and ts.TestID=tq.TestID);"

	if stmt.SQL != exp {
		//fmt.Printf("Sql: %d, Exp: %d\n", len(stmt.SQL), len(exp))
		t.Errorf("Expected\n %s\nGot\n %s", exp, stmt.SQL)
	}
}

func TestSubSQLWhereClause(t *testing.T) {
	fmt.Println("\n\nTestSubSQLWhereClause ***")

	os.Setenv("DATABASE_TYPE", DbTypePostgreSQL)

	stmt := SelectBuilder().
		Select("tq.ID as QID").
		Sub(SelectBuilder().
			Select("left(Qdata,50)").From("QuestionData", "").
			Where(C().EQ("QID", "q.ID"), C().EQ("DataType", "1")).
			Limit(1), "as Tquestion").
		Sub(SelectBuilder().
			Select("s.Title").From("Subjects", "s").
			Where(C().EQ("s.ID", "t.SubjectID")), "as TSubject").
		Select("q.QType", "q.DifficultyLevel", "tq.CorrectMarks", "tq.NegativeMarks", "tq.QCancelMarks", "tq.seqno").
		Select("q.ID", "t.SubjectID", "ts.SeqNo AS SeqNoSubject", "tq.Addedon", "tq.Addedby", "getquestionlanguages(q.ID) as Languages").
		From("TestQuestions", "tq").
		From("Questions", "q").
		From("Topics", "t").
		From("testsubjects", "ts").
		Where(C().EQ("ts.TestID", "tq.TestID"), C().EQ("t.SubjectID", "ts.SubjectID"),
			C().EQ("t.ID", "q.TopicID"), C().EQ("tq.QID", "q.ID"),
			C().INSub("ts.testid", SelectBuilder().Select("id").
				From("tests", "").
				Where(C().GT("id", "?")))).
		Build(false)

	// fmt.Println("Paracount: ", stmt.ParamCount)
	// fmt.Println("ParaFields: ", stmt.ParamFields)
	// fmt.Println("FieldsCount: ", stmt.FieldsCount)
	// fmt.Println("Fields: ", stmt.Fields)

	if stmt.ParamCount != 1 {
		t.Errorf("Expected Paramters\n %d\nGot\n %d", 1, stmt.ParamCount)
	}

	exp := "select tq.ID as QID, (select left(Qdata,50) from questiondata where (DataType=1 and QID=q.ID) limit 1) as Tquestion, " +
		"(select s.Title from subjects s where (s.ID=t.SubjectID)) as TSubject, q.QType, q.DifficultyLevel, " +
		"tq.CorrectMarks, tq.NegativeMarks, tq.QCancelMarks, tq.seqno, q.ID, t.SubjectID, ts.SeqNo AS SeqNoSubject, tq.Addedon, tq.Addedby, " +
		"getquestionlanguages(q.ID) as Languages " +
		"from questions q, topics t, testquestions tq, testsubjects ts " +
		"where (t.ID=q.TopicID and t.SubjectID=ts.SubjectID and tq.QID=q.ID and ts.TestID=tq.TestID " +
		"and ts.testid IN (select id from tests where (id>$1)))"

	if stmt.SQL != exp {
		//fmt.Printf("Sql: %d, Exp: %d\n", len(stmt.SQL), len(exp))
		t.Errorf("Expected\n %s\nGot\n %s", exp, stmt.SQL)
	}
}

func TestRawINWhereClause(t *testing.T) {
	fmt.Println("\n\nTestRawINWhereClause ***")

	os.Setenv("DATABASE_TYPE", DbTypePostgreSQL)

	stmt := SelectBuilder().
		Select("tq.ID as QID").
		Sub(SelectBuilder().
			Select("left(Qdata,50)").From("QuestionData", "").
			Where(C().EQ("QID", "q.ID"), C().EQ("DataType", "1")).
			Limit(1), "as Tquestion").
		Sub(SelectBuilder().
			Select("s.Title").From("Subjects", "s").
			Where(C().EQ("s.ID", "t.SubjectID")), "as TSubject").
		Select("q.QType", "q.DifficultyLevel", "tq.CorrectMarks", "tq.NegativeMarks", "tq.QCancelMarks", "tq.seqno").
		Select("q.ID", "t.SubjectID", "ts.SeqNo AS SeqNoSubject", "tq.Addedon", "tq.Addedby", "getquestionlanguages(q.ID) as Languages").
		From("TestQuestions", "tq").
		From("Questions", "q").
		From("Topics", "t").
		From("testsubjects", "ts").
		Where(C().EQ("ts.TestID", "tq.TestID"), C().EQ("t.SubjectID", "ts.SubjectID"),
			C().EQ("t.ID", "q.TopicID"), C().EQ("tq.QID", "q.ID"),
			C().INAnyArray("ts.testid", false)).
		Build(false)

	// fmt.Println("Paracount: ", stmt.ParamCount)
	// fmt.Println("ParaFields: ", stmt.ParamFields)
	// fmt.Println("FieldsCount: ", stmt.FieldsCount)
	// fmt.Println("Fields: ", stmt.Fields)

	if stmt.ParamCount != 1 {
		t.Errorf("Expected Paramters\n %d\nGot\n %d", 1, stmt.ParamCount)
	}

	exp := "select tq.ID as QID, (select left(Qdata,50) from questiondata where (DataType=1 and QID=q.ID) limit 1) as Tquestion, " +
		"(select s.Title from subjects s where (s.ID=t.SubjectID)) as TSubject, q.QType, q.DifficultyLevel, " +
		"tq.CorrectMarks, tq.NegativeMarks, tq.QCancelMarks, tq.seqno, q.ID, t.SubjectID, ts.SeqNo AS SeqNoSubject, tq.Addedon, tq.Addedby, " +
		"getquestionlanguages(q.ID) as Languages " +
		"from questions q, topics t, testquestions tq, testsubjects ts " +
		"where (t.ID=q.TopicID and t.SubjectID=ts.SubjectID and tq.QID=q.ID and ts.TestID=tq.TestID " +
		"and ts.testid=ANY($1))"

	if stmt.SQL != exp {
		//fmt.Printf("Sql: %d, Exp: %d\n", len(stmt.SQL), len(exp))
		t.Errorf("Expected\n %s\nGot\n %s", exp, stmt.SQL)
	}
}

func TestBuilderMultipleClause(t *testing.T) {
	fmt.Println("\n\nTestBuilderMultipleClause ***")

	stmt := SelectBuilder().
		Select("q.ID", "qd.QID").
		From("questions", "q").
		From("QuestionData", "qd").
		Where(C().EQ("q.ID", "qd.QID"),
			C().EQ("q.TopicID", "?"),
			C().INInt("q.ID", []int{2, 4}, true)).
		WhereGroup(OpOR, C().GTE("q.ID", "?")).
		OrderBy("qd.QID", false).
		OrderBy("q.ID", true).
		Limit(2).
		Build(true)

	// fmt.Println("Paracount: ", stmt.ParamCount)
	// fmt.Println("ParaFields: ", stmt.ParamFields)
	// fmt.Println("FieldsCount: ", stmt.FieldsCount)
	// fmt.Println("Fields: ", stmt.Fields)

	if stmt.ParamCount != 2 {
		t.Errorf("Expected Paramters\n %d\nGot\n %d", 2, stmt.ParamCount)
	}

	exp := "select q.ID, qd.QID from questions q, questiondata qd where (q.ID=qd.QID and q.ID=ANY('{2,4}'::integer[]) and q.TopicID=$1) OR (q.ID>=$2) order by qd.QID asc, q.ID desc limit 2;"
	if stmt.SQL != exp {
		//fmt.Printf("Sql: %d, Exp: %d\n", len(stmt.SQL), len(exp))
		t.Errorf("Expected\n %s\nGot\n %s", exp, stmt.SQL)
	}
}

func TestInsertBuilderMsSQL(t *testing.T) {
	fmt.Println("\n\nTestInsertBuilder ***")

	os.Setenv("DATABASE_TYPE", DbTypeMsSQL)

	stmt := InsertBuilder().Table("users").
		Columns("name", "age").Returning("id").
		Build(true)

	exp := "insert into users(name, age) output inserted.id values(@p1, @p2);"
	if stmt.SQL != exp {
		//fmt.Printf("Sql: %d, Exp: %d\n", len(stmt.SQL), len(exp))
		t.Errorf("Expected\n %s\nGot\n %s", exp, stmt.SQL)
	}
}

func TestInsertBuilderPgSQL(t *testing.T) {
	fmt.Println("\n\nTestInsertBuilder ***")

	os.Setenv("DATABASE_TYPE", DbTypePostgreSQL)

	stmt := InsertBuilder().Table("users").
		Columns("name", "age").Returning("id").
		Build(true)

	exp := "insert into users(name, age) values($1, $2) returning id;"
	if stmt.SQL != exp {
		//fmt.Printf("Sql: %d, Exp: %d\n", len(stmt.SQL), len(exp))
		t.Errorf("Expected\n %s\nGot\n %s", exp, stmt.SQL)
	}
}

func TestUpdateBuilderMsSQL(t *testing.T) {
	fmt.Println("\n\nTestUpdateBuilder ***")

	os.Setenv("DATABASE_TYPE", DbTypeMsSQL)

	stmt := UpdateBuilder().Table("users").
		Columns("name", "age").
		CalcColumn("points", "points+?").
		Where(C().EQ("id", "?")).
		Returning("id").
		Build(true)

	exp := "update users set name=@p1, age=@p2, points=points+@p3 output inserted.id where (id=@p4);"
	if stmt.SQL != exp {
		//fmt.Printf("Sql: %d, Exp: %d\n", len(stmt.SQL), len(exp))
		t.Errorf("Expected\n %s\nGot\n %s", exp, stmt.SQL)
	}
}

func TestUpdateBuilderPgSQL(t *testing.T) {
	fmt.Println("\n\nTestUpdateBuilder ***")

	os.Setenv("DATABASE_TYPE", DbTypePostgreSQL)

	stmt := UpdateBuilder().Table("users").
		Columns("name", "age").
		CalcColumn("points", "points+?").
		Where(C().EQ("id", "?")).
		Returning("id").
		Build(true)

	exp := "update users set name=$1, age=$2, points=points+$3 where (id=$4) returning id;"
	if stmt.SQL != exp {
		//fmt.Printf("Sql: %d, Exp: %d\n", len(stmt.SQL), len(exp))
		t.Errorf("Expected\n %s\nGot\n %s", exp, stmt.SQL)
	}
}

func TestDeleteBuilder(t *testing.T) {
	fmt.Println("\n\nTestDeleteBuilder ***")

	os.Setenv("DATABASE_TYPE", DbTypeMsSQL)

	stmt := DeleteBuilder().Table("users").
		Where(C().EQ("ID", "?")).
		Returning("name").
		Build(true)

	exp := "delete from users output deleted.name where (ID=@p1);"
	if stmt.SQL != exp {
		//fmt.Printf("Sql: %d, Exp: %d\n", len(stmt.SQL), len(exp))
		t.Errorf("Expected\n %s\nGot\n %s", exp, stmt.SQL)
	}
}

func TestProcBuilder(t *testing.T) {
	fmt.Println("\n\nTestProcBuilder ***")

	os.Setenv("DATABASE_TYPE", DbTypeMsSQL)

	stmt := ProcBuilder().Select("id", "name").
		FromProc("proc1").
		Param("email", "regdate").
		Build(true)

	exp := "exec proc1 @p1, @p2;"
	if stmt.SQL != exp {
		t.Errorf("Expected\n %s\nGot\n %s", exp, stmt.SQL)
	}

	os.Setenv("DATABASE_TYPE", DbTypePostgreSQL)

	stmt = ProcBuilder().Perform("proc1").
		Param("email", "regdate").
		RowCount().
		Build(true)

	exp = "perform proc1($1, $2);"
	if stmt.SQL != exp {
		t.Errorf("Expected\n %s\nGot\n %s", exp, stmt.SQL)
	}

	stmt = ProcBuilder().Select("Id", "Email", "Mobile", "Role", "Name", "BillingAddress", "BillingState", "BillingCity", "BillingPincode", "DeliverySameAsBilling", "DeliveryAddress", "DeliveryState", "DeliveryCity", "DeliveryPincode", "CreatedOn", "Disabled", "Lastlogin", "BillingCountry", "DeliveryCountry", "MibsEmpId").
		FromProc("getusers").
		Param("uname", "uemail", "umobile", "urole", "uordercountcompare", "uordercount", "unoofdays", "searchbylogin", "datecrit").
		RowCount().Build(true)

	exp = "select Id, Email, Mobile, Role, Name, BillingAddress, BillingState, BillingCity, BillingPincode, DeliverySameAsBilling, DeliveryAddress, DeliveryState, DeliveryCity, DeliveryPincode, CreatedOn, Disabled, Lastlogin, BillingCountry, DeliveryCountry, MibsEmpId, count(*) over() as rowscount from getusers($1, $2, $3, $4, $5, $6, $7, $8, $9);"
	if stmt.SQL != exp {
		t.Errorf("Expected\n %s\nGot\n %s", exp, stmt.SQL)
	}
}

// ------------------------------
//
// Benchmarkking
//
// ------------------------------

func BenchmarkBuilder(b *testing.B) {
	exp := "select q.ID, qd.QID, count(*) over() as rowscount from questions q, questiondata qd where (q.ID=qd.QID and q.TopicID=$1) order by qd.QID desc;"

	for n := 0; n < b.N; n++ {
		stmt := SelectBuilder().Select("q.ID", "qd.QID").
			From("Questions", "q").
			From("QuestionData", "qd").
			Where(C().EQ("q.ID", "qd.QID"), C().EQ("q.TopicID", "$1")).
			OrderBy("qd.QID", true).
			RowCount().
			Build(true)

		if stmt.SQL != exp {
			b.Errorf("Expected\n %s\nGot\n %s", exp, stmt.SQL)
		}
	}
}

func BenchmarkSubSQLWhereClause(b *testing.B) {
	exp := "select tq.ID as QID, (select left(Qdata,50) from questiondata where (DataType=1 and QID=q.ID) limit 1) as Tquestion, " +
		"(select s.Title from subjects s where (s.ID=t.SubjectID)) as TSubject, q.QType, q.DifficultyLevel, " +
		"tq.CorrectMarks, tq.NegativeMarks, tq.QCancelMarks, tq.seqno, q.ID, t.SubjectID, ts.SeqNo AS SeqNoSubject, tq.Addedon, tq.Addedby, " +
		"getquestionlanguages(q.ID) as Languages " +
		"from questions q, topics t, testquestions tq, testsubjects ts " +
		"where (t.ID=q.TopicID and t.SubjectID=ts.SubjectID and tq.QID=q.ID and ts.TestID=tq.TestID " +
		"and ts.testid IN (select id from tests where (id>$1)))"

	for n := 0; n < b.N; n++ {
		stmt := SelectBuilder().
			Select("tq.ID as QID").
			Sub(SelectBuilder().
				Select("left(Qdata,50)").From("QuestionData", "").
				Where(C().EQ("QID", "q.ID"), C().EQ("DataType", "1")).
				Limit(1), "as Tquestion").
			Sub(SelectBuilder().
				Select("s.Title").From("Subjects", "s").
				Where(C().EQ("s.ID", "t.SubjectID")), "as TSubject").
			Select("q.QType", "q.DifficultyLevel", "tq.CorrectMarks", "tq.NegativeMarks", "tq.QCancelMarks", "tq.seqno").
			Select("q.ID", "t.SubjectID", "ts.SeqNo AS SeqNoSubject", "tq.Addedon", "tq.Addedby", "getquestionlanguages(q.ID) as Languages").
			From("TestQuestions", "tq").
			From("Questions", "q").
			From("Topics", "t").
			From("testsubjects", "ts").
			Where(C().EQ("ts.TestID", "tq.TestID"), C().EQ("t.SubjectID", "ts.SubjectID"),
				C().EQ("t.ID", "q.TopicID"), C().EQ("tq.QID", "q.ID"),
				C().INSub("ts.testid", SelectBuilder().Select("id").
					From("tests", "").
					Where(C().GT("id", "?")))).
			Build(false)

		if stmt.SQL != exp {
			b.Errorf("Expected\n %s\nGot\n %s", exp, stmt.SQL)
		}
	}
}

func BenchmarkBuilderMultipleClause(b *testing.B) {
	exp := "select q.ID, qd.QID from questions q, questiondata qd where (q.ID=qd.QID and q.ID=ANY('{2,4}'::integer[]) and q.TopicID=$1) OR (q.ID>=$2) order by qd.QID asc, q.ID desc limit 2;"

	for n := 0; n < b.N; n++ {
		stmt := SelectBuilder().
			Select("q.ID", "qd.QID").
			From("questions", "q").
			From("QuestionData", "qd").
			Where(C().EQ("q.ID", "qd.QID"),
				C().EQ("q.TopicID", "$1"),
				C().INInt("q.ID", []int{2, 4}, true)).
			WhereGroup(OpOR, C().GTE("q.ID", "$2")).
			OrderBy("qd.QID", false).
			OrderBy("q.ID", true).
			Limit(2).
			Build(true)

		if stmt.SQL != exp {
			b.Errorf("Expected\n %s\nGot\n %s", exp, stmt.SQL)
		}
	}
}
