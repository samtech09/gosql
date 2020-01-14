package gosql

import (
	"strconv"
	"strings"
)

//UpdateBuilder create new instance to updateBuilder
func UpdateBuilder() *updateBuilder {
	u := updateBuilder{}
	u.calcfields = make(map[string]string)
	u.conditionGroups = make(map[int]conditionGroup)
	return &u
}

//Table sets name of table in which data to be updated
func (u *updateBuilder) Table(tablename string) *updateBuilder {
	u.table = tablename
	return u
}

//Columns sets name of columns/fields to be updated
func (u *updateBuilder) Columns(cols ...string) *updateBuilder {
	for _, v := range cols {
		u.fields = append(u.fields, v)
	}
	return u
}

//CalcColumn sets name of columns/fields to be updated with calculated value
func (u *updateBuilder) CalcColumn(col, value string) *updateBuilder {
	u.calcfields[strings.Trim(col, " ")] = strings.Trim(value, " ")
	return u
}

//Where specifies the WHERE clause of sql, it appends WHERE keyword itself.
func (u *updateBuilder) Where(c ...Condition) *updateBuilder {
	cg := conditionGroup{}
	cg.operator = opdefault
	cg.conditions = c

	l := len(u.conditionGroups)
	u.conditionGroups[l] = cg
	return u
}

//WhereGroup adds another grouped condition with AND or OR where clause after the default where clause
// e.g. where (a=1) OR (b=2)
func (u *updateBuilder) WhereGroup(op Operator, c ...Condition) *updateBuilder {
	l := len(u.conditionGroups)
	if l < 1 {
		panic("default Where condition must be added first")
	}

	cg := conditionGroup{}
	cg.operator = op
	cg.conditions = c

	u.conditionGroups[l] = cg
	return u
}

//Returning sets columns to incude in returning clause
func (u *updateBuilder) Returning(cols ...string) *updateBuilder {
	for _, v := range cols {
		u.returningFields = append(u.returningFields, v)
	}
	return u
}

//Build generates the update sql statement
func (u *updateBuilder) Build(terminateWithSemiColon bool) StatementInfo {
	return u.build(terminateWithSemiColon)
}

func (u *updateBuilder) build(terminateWithSemiColon bool) StatementInfo {
	var sql strings.Builder

	// get count of fields
	cnt := len(u.fields)
	if cnt < 1 {
		return StatementInfo{SQL: "no fields to update"}
	}

	u.paramCounter = 0
	u.fieldCounter = 0

	sql.WriteString("update ")
	sql.WriteString(u.table)
	sql.WriteString(" set ")

	for _, fld := range u.fields {
		if u.paramCounter > 0 {
			sql.Write(comma)
		}

		sql.WriteString(fld)
		sql.WriteString("=$")
		sql.WriteString(strconv.Itoa(u.paramCounter + 1))

		// // add field and param to CSV
		u.addFieldToCSV(fld)
		u.addParamToCSV(fld)

		// // increase field and param counters
		// u.fieldCounter++
	}

	for k, v := range u.calcfields {
		if u.fieldCounter > 0 {
			sql.Write(comma)
		}
		sql.WriteString(k)
		sql.WriteString("=")
		// replace '?' with pg param i.e $1, $2 ...
		if strings.Contains(v, "?") {
			tmp := strings.Split(v, "?")
			if len(tmp) > 1 {
				for _, str := range tmp {
					if len(str) < 1 {
						continue
					}

					sql.WriteString(str)
					sql.WriteString("$")
					sql.WriteString(strconv.Itoa(u.paramCounter + 1))

					// add param to CSV
					u.addParamToCSV(v)
				}
			}
		} else {
			sql.WriteString(v)
		}
		// add field to CSV
		u.addFieldToCSV(v)
	}

	if len(u.conditionGroups) > 0 {
		sql.Write(space)
		sql.WriteString(u.getWhereClause())
	}

	if len(u.returningFields) > 0 {
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