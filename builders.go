package gosql

import (
	"sort"
	"strconv"
	"strings"
)

//Operator defines operators to separate where clause groups by logical AND or OR
type Operator int

const (
	opdefault Operator = iota
	//OpAND is logical AND for SQL where clause
	OpAND
	//OpOR is logical AND for SQL where clause
	OpOR
)

var (
	comma      = []byte(", ")
	space      = []byte(" ")
	openbrace  = []byte("(")
	closebrace = []byte(")")
	and        = []byte(" and ")
	closure    = []byte(";")
)

//StatementInfo holds meta data of generated SQL along with SQL itself
type StatementInfo struct {
	//Fields holds name of field for SELECT or UPDATE or INSERT statement
	Fields string
	//FieldsCount is count of fields to be SELECT or UPDATE or INSERT
	FieldsCount int
	//ParamFields holds name of fields used with parameters in generated SQL
	ParamFields string
	//ParamCount is count of total parameters in generated SQL
	ParamCount int
	//ReturningFields holds name of fields used with RETURNING clause
	ReturningFields string
	//SQL is generated Sql statement
	SQL string
}

type builder struct {
	paramCounter    int
	fieldCounter    int
	fieldCsv        strings.Builder
	paramCsv        strings.Builder
	returningCsv    strings.Builder
	conditionGroups map[int]conditionGroup
}

//selectBuilder allow to dynamically build SQL to query database-tables
type selectBuilder struct {
	builder
	selectsql []selectSQL
	fromsql   []string
	joinsql   []string
	groupBy   []string
	orderBy   []string
	limitRows int
	tables    map[string]string
	rowcount  bool
}

//insertBuilder allow to dynamically build SQL to insert record in database
type insertBuilder struct {
	builder
	table           string
	fields          []string
	returningFields []string
}

//updateBuilder allow to dynamically build SQL to update record in database
type updateBuilder struct {
	builder
	table           string
	fields          []string
	calcfields      map[string]string
	returningFields []string
}

//deleteBuilder allow to dynamically build SQL to delete record from database
type deleteBuilder struct {
	builder
	table           string
	returningFields []string
}

func (b *builder) addFieldToCSV(fld string) {
	if fld == "" {
		return
	}
	b.fieldCounter++
	if b.fieldCsv.Len() > 0 {
		b.fieldCsv.Write(comma)
	}
	b.fieldCsv.WriteString(fld)
}
func (b *builder) addParamToCSV(param string) {
	if param == "" {
		return
	}
	b.paramCounter++
	if b.paramCsv.Len() > 0 {
		b.paramCsv.Write(comma)
	}
	b.paramCsv.WriteString(param)
}
func (b *builder) addReturningCSV(fld string) {
	if fld == "" {
		return
	}
	if b.returningCsv.Len() > 0 {
		b.returningCsv.Write(comma)
	}
	b.returningCsv.WriteString(fld)
}

//getWhereClause prepare and return where clause for given conditiongroups and number of parameters added
func (b *builder) getWhereClause() string {
	var sql strings.Builder

	ln := len(b.conditionGroups)
	if ln < 1 {
		return ""
	}

	sql.WriteString("where ")

	// sort condition groups by keys
	var keys []int
	for k := range b.conditionGroups {
		keys = append(keys, k)
	}
	sort.Ints(keys)

	for _, key := range keys {
		cg := b.conditionGroups[key]
		switch cg.operator {
		case OpAND:
			sql.WriteString(" AND ")
		case OpOR:
			sql.WriteString(" OR ")
		}

		sql.WriteString("(")
		// sort array so fields will always be in same order
		sort.Slice(cg.conditions, func(i, j int) bool {
			return cg.conditions[i].fieldname < cg.conditions[j].fieldname
		})

		i := 0
		for _, cond := range cg.conditions {
			if i > 0 {
				sql.Write(and)
			}

			if cond.GetBuilder() != nil {
				// generate sub sql
				subStmp := cond.GetBuilder().build(false, b.paramCounter, true)
				// update param, paracount etc as per sub SQL
				b.addParamToCSV(subStmp.ParamFields)
				b.paramCounter = subStmp.ParamCount

				// write sql like 'filed=(sub sql)'
				sql.WriteString(cond.fieldname)
				// here conditionsql holds operator like = or <= or > etc.
				sql.WriteString(cond.conditionsql)
				sql.Write(openbrace)
				sql.WriteString(subStmp.SQL)
				sql.Write(closebrace)

			} else {
				// replace '?' with pg param i.e $1, $2 ...
				if strings.Contains(cond.conditionsql, "?") {
					tmp := strings.Split(cond.conditionsql, "?")
					if len(tmp) > 1 {
						for _, str := range tmp {
							if len(str) < 1 {
								continue
							}

							sql.WriteString(str)
							sql.WriteString("$")
							sql.WriteString(strconv.Itoa(b.paramCounter + 1))

							//add parameter to csv
							b.addParamToCSV(cond.fieldname)
						}
					}
				} else {
					if len(cond.conditionsql) > 0 {
						sql.WriteString(cond.conditionsql)
					}
				}
			}

			i++
		}
		sql.WriteString(")")
	}

	return sql.String()
}
