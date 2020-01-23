package gosql

import (
	"os"
	"sort"
	"strconv"
	"strings"
)

type selectSQL struct {
	//issub      bool
	sql        string
	subBuilder *procBuilder // for sub-sql building
}

//var _usePgArray bool

//ProcBuilder creates new instance of ProcBuilder.
//It allows to generate SELECT sql statements.
func ProcBuilder() *procBuilder {
	s := selectBuilder{}
	s.tables = make(map[string]string)
	s.conditionGroups = make(map[int]conditionGroup)
	s.limitRows = 0
	s.readonly = true
	paramFormat := os.Getenv("SQL_PARAM_FORMAT")

	switch paramFormat {
	case ParamPostgreSQL:
		s.paramChar = "$"
		s.paramNumeric = true
	case ParamMsSQL:
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
		sql := selectSQL{strings.Trim(v, " "), nil}
		s.selectsql = append(s.selectsql, sql)
	}
	return s
}

//From specifies the FROM clause of sql.
//It adds table that is being used in sql, also allow to use table name alias.
func (s *procBuilder) From(tblname, alias string) *procBuilder {
	if tblname != "" {
		s.tables[alias] = strings.ToLower(tblname)
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
	return s.build(terminateWithSemiColon, 0, false)
}

func (s *procBuilder) build(terminateWithSemiColon bool, startParam int, issub bool) StatementInfo {
	var sql strings.Builder
	s.paramCounter = startParam
	s.fieldCounter = 0

	cnt := len(s.selectsql)
	if cnt < 1 {
		return StatementInfo{SQL: "no fields to select"}
	}

	sql.WriteString("select ")
	for i, sSQL := range s.selectsql {
		if i > 0 {
			sql.Write(comma)
		}

		if sSQL.subBuilder == nil {
			sql.WriteString(sSQL.sql)
		} else {
			sql.Write(openbrace)

			// generate sub-sql
			subStmp := sSQL.subBuilder.build(false, s.paramCounter, true)
			// update param, paracount etc as per sub SQL
			s.paramCounter = subStmp.ParamCount
			s.addParamToCSV(subStmp.ParamFields)

			sql.WriteString(subStmp.SQL)
			sql.Write(closebrace)

			// here sql may have alias for sub-select-statement [ (select ....) as field1 ]
			if sSQL.sql != "" {
				sql.Write(space)
				sql.WriteString(sSQL.sql)
			}
		}

		// do not add fileds to csv for sub-sqls
		if !issub {
			s.addFieldToCSV(sSQL.sql)
		}
	}

	if s.rowcount && !issub {
		sql.WriteString(", count(*) over() as rowscount")
		s.addFieldToCSV("rowscount")
	}

	if len(s.tables) > 0 {
		sql.Write(space)
		x := 0
		sql.WriteString("from ")

		// To store the keys in slice in sorted order
		var keys []string
		for k := range s.tables {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		for _, k := range keys {
			if x > 0 {
				sql.Write(comma)
			}
			sql.WriteString(s.tables[k])
			if k != "" {
				sql.Write(space)
				sql.WriteString(k)
			}
			x++
		}
	}

	// get where clause
	if len(s.conditionGroups) > 0 {
		sql.Write(space)
		sql.WriteString(s.getWhereClause())
	}

	// add group by
	if len(s.groupBy) > 0 {
		sql.Write(space)
		sql.WriteString("group by ")
		for i, str := range s.groupBy {
			if i > 0 {
				sql.Write(comma)
			}
			sql.WriteString(str)
		}
	}

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
