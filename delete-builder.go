package gosql

import (
	"strings"
)

//DeleteBuilder creates new instance of DeleteBuilder.
//It allows to create DELETE sql statements.
func DeleteBuilder() *deleteBuilder {
	u := deleteBuilder{}
	u.conditionGroups = make(map[int]conditionGroup)
	u.initEnv()
	return &u
}

//Table sets name of table in which data to be updated
func (u *deleteBuilder) Table(tablename string) *deleteBuilder {
	u.table = tablename
	return u
}

//Where specifies the WHERE clause of sql, it appends WHERE keyword itself.
func (u *deleteBuilder) Where(c ...ICondition) *deleteBuilder {
	cg := conditionGroup{}
	cg.operator = opdefault
	//cg.conditions = c
	cg.conditions = make([]Condition, 0, len(c))
	for _, cd := range c {
		cg.conditions = append(cg.conditions, cd.(Condition))
	}

	l := len(u.conditionGroups)
	u.conditionGroups[l] = cg
	return u
}

//WhereGroup adds another grouped condition with AND or OR where clause after the default where clause
// e.g. where (a=1) OR (b=2)
func (u *deleteBuilder) WhereGroup(op Operator, c ...ICondition) *deleteBuilder {
	l := len(u.conditionGroups)
	if l < 1 {
		panic("default Where condition must be added first")
	}

	cg := conditionGroup{}
	cg.operator = op
	//cg.conditions = c
	cg.conditions = make([]Condition, 0, len(c))
	for _, cd := range c {
		cg.conditions = append(cg.conditions, cd.(Condition))
	}

	u.conditionGroups[l] = cg
	return u
}

//Returning sets columns to incude in returning clause
func (u *deleteBuilder) Returning(cols ...string) *deleteBuilder {
	for _, v := range cols {
		u.returningFields = append(u.returningFields, v)
	}
	return u
}

//Build generates the insert sql statement
func (u *deleteBuilder) Build(terminateWithSemiColon bool) StatementInfo {
	var sql strings.Builder
	u.paramCounter = 0
	u.fieldCounter = 0

	sql.WriteString("delete from ")
	sql.WriteString(u.table)

	if len(u.returningFields) > 0 && u.dbtype == DbTypeMsSQL {
		sql.Write(space)
		sql.WriteString("output ")
		i := 0
		for _, fld := range u.returningFields {
			if i > 0 {
				sql.Write(comma)
			}
			sql.WriteString("deleted.")
			sql.WriteString(fld)
			i++
			u.addReturningCSV(fld)
		}
	}

	if len(u.conditionGroups) > 0 {
		sql.Write(space)
		sql.WriteString(u.getWhereClause())
	}

	if len(u.returningFields) > 0 && u.dbtype == DbTypePostgreSQL {
		sql.Write(space)
		sql.WriteString("returning ")
		i := 0
		for _, fld := range u.returningFields {
			if i > 0 {
				sql.Write(comma)
			}
			sql.WriteString(fld)
			i++
			u.addReturningCSV(fld)
		}
	}

	if terminateWithSemiColon {
		sql.Write(closure)
	}

	stmt := StatementInfo{}
	stmt.ParamCount = u.paramCounter
	stmt.ParamFields = u.paramCsv.String()
	stmt.Fields = u.fieldCsv.String()
	stmt.FieldsCount = u.fieldCounter
	stmt.ReturningFields = u.returningCsv.String()
	stmt.SQL = sql.String()

	return stmt
}
