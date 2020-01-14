package gosql

import "strings"

type conditionGroup struct {
	operator Operator
	//conditions []string
	//fields     []string
	conditions []Condition
}

//ICondition interface.
type ICondition interface {
	GetSQL() string
	GetFieldName() string
	GetBuilder() *selectBuilder
}

//Condition holds generated SQL for given condition.
type Condition struct {
	fieldname    string
	conditionsql string
	subBuilder   *selectBuilder // for sub-sql builing
}

//C creates a new Condition
func C() *Condition {
	return &Condition{}
}

//GetSQL returns generated SQL for given condition.
func (c Condition) GetSQL() string {
	return c.conditionsql
}

//GetFieldName returns generated SQL for given condition.
func (c Condition) GetFieldName() string {
	return c.fieldname
}

//GetBuilder returs sub builder otherwise nil.
func (c Condition) GetBuilder() *selectBuilder {
	return c.subBuilder
}

//
// --------------------------
//
//	Direct operators
//

//EQ generate sql with '=' operator.
func (c *Condition) EQ(col1, col2 string) Condition {
	c.fieldname = col1
	c.conditionsql = concat(col1, "=", col2)
	return *c
}

//NEQ generate sql with '!=' operator.
func (c *Condition) NEQ(col1, col2 string) Condition {
	c.fieldname = col1
	c.conditionsql = concat(col1, "!=", col2)
	return *c
}

//GT generate sql with '>' operator.
func (c *Condition) GT(col1, col2 string) Condition {
	c.fieldname = col1
	c.conditionsql = concat(col1, ">", col2)
	return *c
}

//GTE generate sql with '>=' operator.
func (c *Condition) GTE(col1, col2 string) Condition {
	c.fieldname = col1
	c.conditionsql = concat(col1, ">=", col2)
	return *c
}

//LT generate sql with '<' operator.
func (c *Condition) LT(col1, col2 string) Condition {
	c.fieldname = col1
	c.conditionsql = concat(col1, "<", col2)
	return *c
}

//LTE generate sql with '<=' operator.
func (c *Condition) LTE(col1, col2 string) Condition {
	c.fieldname = col1
	c.conditionsql = concat(col1, "<=", col2)
	return *c
}

//Between generate sql with 'between clause'.
func (c *Condition) Between(col1, col2, col3 string) Condition {
	c.fieldname = col1
	c.conditionsql = concat(col1, " between ", col2, " and ", col3)
	return *c
}

//
// --------------------------
//
//	Conditions with Sub-sqls
//

//EQSub generate sql with '=' operator along with SubSQL.
func (c *Condition) EQSub(col1 string, builder *selectBuilder) Condition {
	c.fieldname = col1
	c.subBuilder = builder
	c.conditionsql = "="
	return *c
}

//NEQSub generate sql with '!=' operator along with SubSQL.
func (c *Condition) NEQSub(col1 string, builder *selectBuilder) Condition {
	c.fieldname = col1
	c.subBuilder = builder
	c.conditionsql = "!="
	return *c
}

//GTSub generate sql with '>' operator along with SubSQL.
func (c *Condition) GTSub(col1 string, builder *selectBuilder) Condition {
	c.fieldname = col1
	c.subBuilder = builder
	c.conditionsql = ">"
	return *c
}

//GTESub generate sql with '>=' operator along with SubSQL.
func (c *Condition) GTESub(col1 string, builder *selectBuilder) Condition {
	c.fieldname = col1
	c.subBuilder = builder
	c.conditionsql = ">="
	return *c
}

//LTSub generate sql with '<' operator along with SubSQL.
func (c *Condition) LTSub(col1 string, builder *selectBuilder) Condition {
	c.fieldname = col1
	c.subBuilder = builder
	c.conditionsql = "<"
	return *c
}

//LTESub generate sql with '<=' operator along with SubSQL.
func (c *Condition) LTESub(col1 string, builder *selectBuilder) Condition {
	c.fieldname = col1
	c.subBuilder = builder
	c.conditionsql = "<="
	return *c
}

//
// --------------------------
//
//	IN clause conditions
//

//INInt create IN clause for given field and slice of int.
func (c *Condition) INInt(col string, in []int, usePgArray bool) Condition {
	c.fieldname = col
	csv := sliceToStringInt(in, ",")
	if usePgArray {
		c.conditionsql = col + "=ANY('{" + csv + "}'::integer[])"
	} else {
		c.conditionsql = col + " IN (" + csv + ")"
	}
	return *c
}

//INFloat create IN clause for given field and slice of float64.
func (c *Condition) INFloat(col string, in []float64, usePgArray bool) Condition {
	c.fieldname = col
	csv := sliceToStringFloat(in, ",")
	if usePgArray {
		c.conditionsql = col + "=ANY('{" + csv + "}'::numeric[])"
	} else {
		c.conditionsql = col + " IN (" + csv + ")"
	}
	return *c
}

//INStr create IN clause for given field and slice of string.
func (c *Condition) INStr(col string, in []string, usePgArray bool) Condition {
	c.fieldname = col
	if usePgArray {
		csv := strings.Join(in, ",")
		c.conditionsql = col + "=ANY('{" + csv + "}'::text[])"
	} else {
		csv := strings.Join(in, "','")
		c.conditionsql = col + " IN ('" + csv + "')"
	}
	return *c
}

//INSub create IN clause for given field with sub-sql.
func (c *Condition) INSub(col string, builder *selectBuilder) Condition {
	c.fieldname = col
	c.subBuilder = builder
	c.conditionsql = " IN "
	//c.conditionsql = concat(col, " IN (", builder.Build(false), ")")
	return *c
}
