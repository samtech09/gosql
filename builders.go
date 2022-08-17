//Package gosql - SQL builder with GO code generation
package gosql

import (
	"os"
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
	//OpOR is logical OR for SQL where clause
	OpOR
)

const (
	//DbTypePostgreSQL sets SQL format to PostgreSQL
	DbTypePostgreSQL string = "pgsql"
	//DbTypeMsSQL sets SQL format to MS-SQL
	DbTypeMsSQL string = "mssql"
	//DbTypeMySQL sets SQL format to MySQL
	DbTypeMySQL string = "mysql"
)

var (
	comma      = []byte(", ")
	space      = []byte(" ")
	openbrace  = []byte("(")
	closebrace = []byte(")")
	and        = []byte(" and ")
	closure    = []byte(";")
)

//StatementInfo holds meta data of generated SQL along with SQL itself.
type StatementInfo struct {
	//Fields holds name of comma separated fields for SELECT or UPDATE or INSERT statement.
	Fields string
	//FieldsCount is count of fields to be SELECT or UPDATE or INSERT.
	FieldsCount int
	//ParamFields holds name of comma separated fields required as parameters in generated SQL.
	ParamFields string
	//ParamCount is count of total parameters in generated SQL.
	ParamCount int
	//ReturningFields holds name of comma separated fields returned with PostgreSQL RETURNING clause.
	ReturningFields string
	//SQL is generated Sql statement.
	SQL string
	//ReadOnly tell whether the statement is ReadOnly or write to database. gosql auto set SQLs generated with SelectBuilder as readonly, other as write.
	//You can override this behavious by calling NoReadOnly() method of SelectBuilder
	ReadOnly bool
}

type builder struct {
	paramCounter    int
	fieldCounter    int
	fieldCsv        strings.Builder
	paramCsv        strings.Builder
	returningCsv    strings.Builder
	conditionGroups map[int]conditionGroup
	readonly        bool
	paramChar       string
	paramNumeric    bool
	dbtype          string
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

//selectBuilder allow to dynamically build SQL to query database-tables
type procBuilder struct {
	builder
	selectsql []string
	proc      string
	orderBy   []string
	limitRows int
	args      []string
	rowcount  bool
	perform   bool
}

//initEnv parse environment variables and set database type and paramter format
func (b *builder) initEnv() {
	b.dbtype = os.Getenv("DATABASE_TYPE")
	paramCharacter := os.Getenv("PARAM_CHAR")
	paramIsNumeric := os.Getenv("PARAM_APPEND_NUMBER")

	switch b.dbtype {
	case DbTypePostgreSQL:
		if paramCharacter != "" {
			b.paramChar = paramCharacter
		} else {
			b.paramChar = "$"
		}
		if paramIsNumeric != "0" {
			b.paramNumeric = true
		}
	case DbTypeMsSQL:
		if paramCharacter != "" {
			b.paramChar = paramCharacter
		} else {
			b.paramChar = "@p"
		}
		if paramIsNumeric != "0" {
			b.paramNumeric = true
		}
	default:
		b.dbtype = DbTypeMySQL
		if paramCharacter != "" {
			b.paramChar = paramCharacter
		} else {
			b.paramChar = "?"
		}
		if paramIsNumeric == "1" {
			b.paramNumeric = true
		}
	}
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
			return cg.conditions[i].GetFieldName() < cg.conditions[j].GetFieldName()
		})

		i := 0
		for _, cond := range cg.conditions {
			if i > 0 {
				sql.Write(and)
			}

			condSql := cond.GetSQL()

			if cond.GetBuilder() != nil {
				// generate sub sql
				subStmp := cond.GetBuilder().build(false, b.paramCounter, true)
				// update param, paracount etc as per sub SQL
				b.addParamToCSV(subStmp.ParamFields)
				b.paramCounter = subStmp.ParamCount

				// write sql like 'filed=(sub sql)'
				sql.WriteString(cond.GetFieldName())
				// here conditionsql holds operator like = or <= or > etc.
				sql.WriteString(condSql)
				sql.Write(openbrace)
				sql.WriteString(subStmp.SQL)
				sql.Write(closebrace)

			} else {
				// replace '?' with pg param i.e $1, $2 ...
				//
				// if param char not already "?" and also there is no need to add numbers in param
				// then do nothing, otherwise replace parameter format
				if strings.Contains(condSql, "?") && b.paramChar != "?" && b.paramNumeric {
					tmp := strings.Split(condSql, "?")
					if len(tmp) > 1 {
						for _, str := range tmp {
							if len(str) < 1 {
								continue
							}
							if str == ")" { // sub-sql or ANY(?) ends
								sql.WriteString(str)
								continue
							}

							sql.WriteString(str)
							sql.WriteString(b.paramChar)
							if b.paramNumeric {
								sql.WriteString(strconv.Itoa(b.paramCounter + 1))
							}

							//add parameter to csv
							b.addParamToCSV(cond.GetFieldName())
						}
					}
				} else {
					if len(condSql) > 0 {
						sql.WriteString(condSql)
					}
				}
			}

			i++
		}
		sql.WriteString(")")
	}

	return sql.String()
}
