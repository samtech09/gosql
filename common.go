package gosql

import (
	"encoding/json"
	"strconv"
	"strings"
)

//sliceToStringInt convert slice to int to comma separated string
func sliceToStringInt(a []int, sep string) string {
	if len(a) == 0 {
		return ""
	}

	b := make([]string, len(a))
	for i, v := range a {
		b[i] = strconv.Itoa(v)
	}
	return strings.Join(b, sep)
}

//sliceToStringFloat convert slice to float64 to comma separated string
func sliceToStringFloat(a []float64, sep string) string {
	if len(a) == 0 {
		return ""
	}

	b := make([]string, len(a))
	for i, v := range a {
		b[i] = strconv.FormatFloat(v, 'f', 6, 64)
	}
	return strings.Join(b, sep)
}

func concat(args ...string) string {
	var b strings.Builder
	for _, s := range args {
		b.WriteString(s)
	}
	return b.String()
}

func structToJSONString(d interface{}) (string, error) {
	data, err := json.Marshal(d)
	if err != nil {
		return "", err
	}
	jsonstring := string(data[:])
	return jsonstring, nil
}
