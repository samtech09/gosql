package gosql

import "strings"

//inBuilder helps to create partial sql to use with IN clause
type inBuilder struct {
	_usePgArray bool
}

//InBuilder returns new instance of selectBuilder
func InBuilder(usePgArray bool) *inBuilder {
	s := inBuilder{usePgArray}
	return &s
}

//BuildIntIN returns IN clause for given field and array (slice) for selected DBMS format
func (s *inBuilder) BuildIntIN(fieldName string, in []int) string {
	csv := sliceToStringInt(in, ",")

	if s._usePgArray {
		return fieldName + "=ANY('{" + csv + "}'::integer[])"
	}
	return fieldName + " IN (" + csv + ")"
}

//BuildFloatIN returns IN clause for given field and array (slice) for selected DBMS format
func (s *inBuilder) BuildFloatIN(fieldName string, in []float64) string {
	csv := sliceToStringFloat(in, ",")

	if s._usePgArray {
		return fieldName + "=ANY('{" + csv + "}'::numeric[])"
	}
	return fieldName + " IN (" + csv + ")"
}

//BuildStrIN returns IN clause for given field and array (slice) for selected DBMS format
func (s *inBuilder) BuildStrIN(fieldName string, in []string) string {
	if s._usePgArray {
		csv := strings.Join(in, ",")
		return fieldName + "=ANY('{" + csv + "}'::text[])"
	}
	csv := strings.Join(in, "','")
	return fieldName + " IN ('" + csv + "')"
}
