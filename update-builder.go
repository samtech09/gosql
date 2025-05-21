package gosql

import (
	"strconv"
	"strings"
)

// UpdateBuilder create new instance of UpdateBuilder.
// It allows to create UPDATE sql statements.
func UpdateBuilder() *updateBuilder {
	u := updateBuilder{}
	u.calcfields = make(map[string]string)
	u.conditionGroups = make(map[int]conditionGroup)
	u.initEnv()
	return &u
}

// Table sets name of table in which data to be updated.
func (u *updateBuilder) Table(tablename string) *updateBuilder {
	u.table = tablename
	return u
}

// Columns sets name of columns/fields to be updated.
func (u *updateBuilder) Columns(cols ...string) *updateBuilder {
	for _, v := range cols {
		u.fields = append(u.fields, v)
	}
	return u
}

// CalcColumn sets name of columns/fields to be updated with calculated value.
// Can be used for inplace updation like
//
//	set points=points+10
func (u *updateBuilder) CalcColumn(col, value string) *updateBuilder {
	u.calcfields[strings.Trim(col, " ")] = strings.Trim(value, " ")
	return u
}

// Where specifies the WHERE clause of sql, it appends WHERE keyword itself.
func (u *updateBuilder) Where(c ...ICondition) *updateBuilder {
	cg := conditionGroup{}
	cg.outer_op = opdefault
	//cg.conditions = c
	cg.conditions = make([]Condition, 0, len(c))
	for _, cd := range c {
		cg.conditions = append(cg.conditions, cd.(Condition))
	}

	l := len(u.conditionGroups)
	u.conditionGroups[l] = cg
	return u
}

// WhereGroup adds another grouped condition with AND or OR where clause after the default where clause.
// For example
//
//	where (a=1) OR (b=2)
//
// outerOp defined operator between two WhereGroups or between a WhereGroup and main where block.
//
// innerOp defines operator between two conditions within the WhereGroup
func (u *updateBuilder) WhereGroup(outerOp Operator, innerOp Operator, c ...ICondition) *updateBuilder {
	l := len(u.conditionGroups)
	if l < 1 {
		panic("default Where condition must be added first")
	}

	cg := conditionGroup{}
	cg.outer_op = outerOp
	cg.inner_op = innerOp
	//cg.conditions = c
	cg.conditions = make([]Condition, 0, len(c))
	for _, cd := range c {
		cg.conditions = append(cg.conditions, cd.(Condition))
	}

	u.conditionGroups[l] = cg
	return u
}

// Returning sets columns to incude in returning clause.
func (u *updateBuilder) Returning(cols ...string) *updateBuilder {
	for _, v := range cols {
		u.returningFields = append(u.returningFields, v)
	}
	return u
}

// Build generates the update sql statement along with meta information.
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
		sql.WriteString("=")
		sql.WriteString(u.paramChar)
		if u.paramNumeric {
			sql.WriteString(strconv.Itoa(u.paramCounter + 1))
		}

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
		//
		// if param char not already "?" and also there is no need to add numbers in param
		// then do nothing, otherwise replace parameter format
		if strings.Contains(v, "?") && u.paramChar != "?" && u.paramNumeric {
			tmp := strings.Split(v, "?")
			if len(tmp) > 1 {
				for _, str := range tmp {
					if len(str) < 1 {
						continue
					}

					sql.WriteString(str)
					sql.WriteString(u.paramChar)
					if u.paramNumeric {
						sql.WriteString(strconv.Itoa(u.paramCounter + 1))
					}
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

	if len(u.returningFields) > 0 && u.dbtype == DbTypeMsSQL {
		sql.Write(space)
		sql.WriteString("output ")
		i := 0
		for _, fld := range u.returningFields {
			if i > 0 {
				sql.Write(comma)
			}
			sql.WriteString("inserted.")
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
