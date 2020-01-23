package gosql

import (
	"os"
	"strconv"
	"strings"
)

//ProcBuilder creates new instance of ProcBuilder.
//It allows to generate SELECT sql statements.
func ProcBuilder() *procBuilder {
	s := procBuilder{}
	s.limitRows = 0
	s.readonly = true
	paramFormat := os.Getenv("DATABASE_TYPE")

	switch paramFormat {
	case DbTypePostgreSQL:
		s.paramChar = "$"
		s.paramNumeric = true
	case DbTypeMsSQL:
		s.paramChar = "@"
		s.paramNumeric = true
	default:
		s.paramChar = "?"
		s.paramNumeric = false
	}

	//fmt.Printf("para format: %s, Char: %s", paramFormat, s.paramChar)

	return &s
}

//Select specifies the fields for select clause.
func (s *procBuilder) Select(fields ...string) *procBuilder {
	for _, v := range fields {
		s.selectsql = append(s.selectsql, strings.Trim(v, " "))
	}
	return s
}

//Select specifies the fields for select clause.
func (s *procBuilder) Perform(procname string) *procBuilder {
	if procname == "" {
		panic("invalid procname")
	}
	s.proc = procname
	s.perform = true
	return s
}

//Exec specifies name of proc to be executed along with count of parameter count.
//It adds table that is being used in sql, also allow to use table name alias.
func (s *procBuilder) FromProc(procname string) *procBuilder {
	if procname == "" {
		panic("invalid procname")
	}
	s.proc = procname
	s.perform = false
	return s
}

//Select specifies the fields for select clause.
func (s *procBuilder) Param(paraNames ...string) *procBuilder {
	for _, v := range paraNames {
		s.args = append(s.args, strings.Trim(v, " "))
	}
	return s
}

//OrderBy specifies the ORDER BY clause of sql. Different fields may have different ordering (asc or desc).
func (s *procBuilder) OrderBy(fieldname string, descending bool) *procBuilder {
	if descending {
		s.orderBy = append(s.orderBy, strings.Trim(fieldname, " ")+" desc")
	} else {
		s.orderBy = append(s.orderBy, strings.Trim(fieldname, " ")+" asc")
	}
	return s
}

//Limit limits number of resultant rows.
func (s *procBuilder) Limit(numRows int) *procBuilder {
	s.limitRows = numRows
	return s
}

//RowCount appends rowcount field at select result with count of rows in resultset.
//During scanning rows, it helps to create slice of exact capacity and avoid repetitive allocations.
func (s *procBuilder) RowCount() *procBuilder {
	s.rowcount = true
	return s
}

//Select specifies the fields for select clause.
func (s *procBuilder) NoReadOnly() *procBuilder {
	s.readonly = false
	return s
}

// Build generates the select SQL along with meta information.
func (s *procBuilder) Build(terminateWithSemiColon bool) StatementInfo {
	return s.build(terminateWithSemiColon, 0)
}

func (s *procBuilder) build(terminateWithSemiColon bool, startParam int) StatementInfo {
	var sql strings.Builder
	s.paramCounter = startParam
	s.fieldCounter = 0

	cnt := len(s.selectsql)
	if cnt < 1 && !s.perform {
		return StatementInfo{SQL: "no fields to select"}
	}

	if s.perform {
		sql.WriteString("perform")
	} else {
		sql.WriteString("select ")
		for i, sSQL := range s.selectsql {
			if i > 0 {
				sql.Write(comma)
			}
			sql.WriteString(sSQL)
			s.addFieldToCSV(sSQL)
		}
	}

	if s.rowcount && !s.perform {
		sql.WriteString(", count(*) over() as rowscount")
		s.addFieldToCSV("rowscount")
	}

	sql.Write(space)
	if !s.perform {
		sql.WriteString("from")
		sql.Write(space)
	}
	sql.WriteString(s.proc)
	sql.Write(openbrace)

	for i, arg := range s.args {
		if i > 0 {
			sql.Write(comma)
		}

		sql.WriteString(s.paramChar)
		if s.paramNumeric {
			sql.WriteString(strconv.Itoa(s.paramCounter + 1))
		} else {
			sql.WriteString(arg)
		}
		s.paramCounter++
	}
	sql.Write(closebrace)

	// add order by
	if len(s.orderBy) > 0 {
		sql.Write(space)
		sql.WriteString("order by ")
		for i, str := range s.orderBy {
			if i > 0 {
				sql.Write(comma)
			}
			sql.WriteString(str)
		}
	}

	if s.limitRows > 0 {
		sql.Write(space)
		sql.WriteString("limit " + strconv.Itoa(s.limitRows))
	}

	if terminateWithSemiColon {
		sql.Write(closure)
	}

	stmt := StatementInfo{}
	stmt.ParamCount = s.paramCounter
	stmt.ParamFields = s.paramCsv.String()
	stmt.Fields = s.fieldCsv.String()
	stmt.FieldsCount = s.fieldCounter
	stmt.SQL = sql.String()
	stmt.ReadOnly = s.readonly
	return stmt
}
