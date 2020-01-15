//Copyright (c) Santosh Gupta <github.com/samtech09>

package gosql

import (
	"fmt"
	"testing"
)

func TestWriteJSON(t *testing.T) {
	fmt.Println("\n\nTestWhereClause ***")

	stmt1 := SelectBuilder().Select("q.ID", "qd.Title").
		From("Questions", "q").
		From("QuestionData", "qd").
		Where(C().EQ("q.ID", "qd.QID"), C().EQ("q.TopicID", "$1")).
		OrderBy("qd.QID", true).
		RowCount().
		Build(true)

	stmt2 := SelectBuilder().
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

	fw := NewFileWriter(5)
	fw.Queue(stmt1, "Ques", "qlist", "Gives list of questions to populate dropdown")
	fw.Queue(stmt2, "Ques", "qdata", "Gives complete question data for all questions")

	fw.Write("/tmp", "sqlbuilder", "sqltest", WriteGoCodeAndJSON)

	// if !ret {
	// 	t.Errorf("Write completed with errors")
	// }
}

func TestWriteJSONLoader(t *testing.T) {
	fmt.Println("\n\nTestWhereClause ***")

	stmt1 := SelectBuilder().Select("q.ID", "qd.Title").
		From("Questions", "q").
		From("QuestionData", "qd").
		Where(C().EQ("q.ID", "qd.QID"), C().EQ("q.TopicID", "$1")).
		OrderBy("qd.QID", true).
		RowCount().NoReadOnly().
		Build(true)

	stmt2 := SelectBuilder().
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

	fw := NewFileWriter(5)
	fw.Queue(stmt1, "Ques", "qlist", "Gives list of questions to populate dropdown")
	fw.Queue(stmt2, "Ques", "qdata", "Gives complete question data for all questions")

	fw.Write("/tmp", "sqlloader", "sqltest", WriteJSONandJSONLoaderGoCode)

	// if !ret {
	// 	t.Errorf("Write completed with errors")
	// }
}
