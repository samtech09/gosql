package gosql

import (
	"strconv"
	"strings"
)

//InsertBuilder creates new instance of InsertBuilder.
//It allows to create INSERT sql statements.
func InsertBuilder() *insertBuilder {
	n := insertBuilder{}
	n.initEnv()
	return &n
}

//Table sets name of table in which data to be inserted.
func (n *insertBuilder) Table(tablename string) *insertBuilder {
	n.table = tablename
	return n
}

//Columns sets name of columns/fields to be inserted.
func (n *insertBuilder) Columns(cols ...string) *insertBuilder {
	for _, v := range cols {
		n.fields = append(n.fields, v)
	}
	return n
}

//Returning sets columns to incude in returning clause supported by PostgreSQL.
func (n *insertBuilder) Returning(cols ...string) *insertBuilder {
	for _, v := range cols {
		n.returningFields = append(n.returningFields, v)
	}
	return n
}

//Build generates the insert sql statement along with meta information.
func (n *insertBuilder) Build(terminateWithSemiColon bool) StatementInfo {
	var sql strings.Builder

	// get count of fields
	cnt := len(n.fields)
	if cnt < 1 {
		return StatementInfo{SQL: "no fields to insert"}
	}

	sql.WriteString("insert into ")
	sql.WriteString(n.table)
	sql.Write(openbrace)

	i := 0
	for _, fld := range n.fields {
		if i > 0 {
			sql.Write(comma)
		}
		sql.WriteString(fld)
		i++
		n.addFieldToCSV(fld)
		n.addParamToCSV(fld)
	}
	sql.Write(closebrace)

	if len(n.returningFields) > 0 && n.dbtype == DbTypeMsSQL {
		sql.Write(space)
		sql.WriteString("output ")
		i = 0
		for _, fld := range n.returningFields {
			if i > 0 {
				sql.Write(comma)
			}
			sql.WriteString("inserted.")
			sql.WriteString(fld)
			i++
			n.addReturningCSV(fld)
		}
	}

	sql.WriteString(" values(")

	for i = 1; i <= cnt; i++ {
		if i > 1 {
			sql.Write(comma)
		}
		sql.WriteString(n.paramChar)
		if n.paramNumeric {
			sql.WriteString(strconv.Itoa(i))
		}
	}
	sql.Write(closebrace)

	if len(n.returningFields) > 0 && n.dbtype == DbTypePostgreSQL {
		sql.Write(space)
		sql.WriteString("returning ")
		i = 0
		for _, fld := range n.returningFields {
			if i > 0 {
				sql.Write(comma)
			}
			sql.WriteString(fld)
			i++
			n.addReturningCSV(fld)
		}
	}

	if terminateWithSemiColon {
		sql.Write(closure)
	}

	stmt := StatementInfo{}
	stmt.ParamCount = n.paramCounter
	stmt.ParamFields = n.paramCsv.String()
	stmt.Fields = n.fieldCsv.String()
	stmt.FieldsCount = n.fieldCounter
	stmt.ReturningFields = n.returningCsv.String()
	stmt.SQL = sql.String()

	return stmt
}

//func (n *insertBuilder) getReturning()
