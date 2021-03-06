//Copyright (c) Santosh Gupta <github.com/samtech09>

package gosql

import (
	"fmt"
	"os"
	"path"
	"strconv"
	"strings"
)

//WriteOption configure output files to write
type WriteOption int

const (
	//WriteJSON outputs SQL and metadata to JSON file
	WriteJSON WriteOption = iota
	//WriteGoCode outputs SQL and metadata to GO code file
	WriteGoCode
	//WriteGoCodeAndJSON outputs SQL and metadata to JSON file and GO code file both
	WriteGoCodeAndJSON
	//WriteJSONandJSONLoaderGoCode outputs SQL and key to JSON and create GO code to load SQL statements from that JSON file.
	// It is much flexible way that allows to sideload SQLs from JSON file
	WriteJSONandJSONLoaderGoCode
)

type FileWriter struct {
	writequeue  []sqlEntry
	jsonBuilder strings.Builder
	codeBuilder strings.Builder
	entry       int
	writeoption WriteOption
}

type sqlEntry struct {
	StatementInfo
	Key         string
	Description string
}

//NewFileWriter create new writer to write generated SQL and metadata to disk file.
//
// queueSize: sets initial capacity of internal Slice that holds Statements,
// setting it with appropriate capacity will reduce repetitive memory allocations.
// Default queue size is set to 50.
func NewFileWriter(queueSize int) *FileWriter {
	w := FileWriter{}
	w.writequeue = make([]sqlEntry, 0, 50)
	return &w
}

//Queue adds given StatementInfo to write queue for writing to file later.
func (w *FileWriter) Queue(si StatementInfo, group, key, purpose string) {
	ukey := strings.Title(group) + strings.Title(key)
	se := sqlEntry{si, ukey, purpose}
	w.writequeue = append(w.writequeue, se)
}

//Write write SQL and metadata to files 'sqlbuilder.*' in given folder.
//
// packageName: set package for generated GO code. If writing only to JSON file then pass empty string.
//
//    Note: existing files with same name will be overwritten in outFolder.
func (w *FileWriter) Write(outFolder, outfileName, packageName string, option WriteOption) {
	w.writeoption = option

	for _, se := range w.writequeue {
		switch option {
		case WriteJSON:
			w.writeJSON(&se)

		case WriteGoCode:
			w.writeCode(&se)

		case WriteJSONandJSONLoaderGoCode:
			if w.writeJSON(&se) {
				w.writeCode(&se)
			}

		default: //WriteGoCodeAndJSON is default
			if w.writeJSON(&se) {
				w.writeCode(&se)
			}

		}
	}

	// write output file
	switch option {
	case WriteJSON:
		w.writeJSONFile(path.Join(outFolder, outfileName+".json"))

	case WriteGoCode:
		w.writeCodeFile(packageName, path.Join(outFolder, outfileName+".go"))

	case WriteJSONandJSONLoaderGoCode:
		w.writeJSONLoaderFile(outFolder, outfileName, packageName)

	default: //WriteGoCodeAndJSON is default
		w.writeJSONFile(path.Join(outFolder, outfileName+".json"))
		w.writeCodeFile(packageName, path.Join(outFolder, outfileName+".go"))

	}
}

func (w *FileWriter) writeJSONFile(file string) {
	writeFile(file, "[", "]", &w.jsonBuilder)
}

func (w *FileWriter) writeCodeFile(pkg, file string) {
	header := "//" + pkg + " package is autogenerated by gosql.\n//Do not edit this file.\n//Changes will be lost on next auto-generate.\n"
	header += "package " + pkg + "\n\n"
	writeFile(file, header, "", &w.codeBuilder)
}

//writeJSONLoaderFile write json file as well as GO code to load sqls from JSON
func (w *FileWriter) writeJSONLoaderFile(folder, filename, pkg string) {
	writeFile(path.Join(folder, filename+".json"), "{", "}", &w.jsonBuilder)

	header := "//" + pkg + " package is autogenerated by gosql.\n//Do not edit this file.\n//Changes will be lost on next auto-generate.\n"
	header += "package " + pkg + "\n\n"
	header += `import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

var _jsonsqls map[string]string

type Statement struct {
	ReadOnly  bool
	SQL       string
}

func fatal(msg ...string) {
	fmt.Println(msg)
	os.Exit(1)
}

//LoadSQLs load sql into map from json file generated by gosql
func LoadSQLs(f string) {
	file, err := os.Open(f)
	if err != nil {
		fatal("error reading file ", f, " ", err.Error())
	}
	defer file.Close()

	byteValue, err := ioutil.ReadAll(file)
	if err != nil {
		fatal("read failure ", err.Error())
	}

	err = json.Unmarshal([]byte(byteValue), &_jsonsqls)
	if err != nil {
		fatal("parse failure ", err.Error())
	}
}

func parseStmt(s string) Statement {
	stmt := Statement{}
	ro := s[:1]
	if ro == "T" {
		stmt.ReadOnly = true
	}
	stmt.SQL = s[2:len(s)]
	return stmt
}` + "\n\n"

	writeFile(path.Join(folder, filename+".go"), header, "", &w.codeBuilder)

}

func writeFile(file, header, footer string, content *strings.Builder) {
	f, err := os.Create(file)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	// writer header
	_, err = f.WriteString(header)
	if err != nil {
		panic(err)
	}

	// write content
	_, err = f.WriteString(content.String())
	if err != nil {
		panic(err)
	}

	// writer footer
	_, err = f.WriteString(footer)
	if err != nil {
		panic(err)
	}

	err = f.Sync()
	if err != nil {
		panic(err)
	}
}

//writeJSON create and write JSON to builder for given sqlEntry
func (w *FileWriter) writeJSON(se *sqlEntry) bool {
	json, err := structToJSONString(se)
	if err != nil {
		fmt.Printf("Error writing [%s] : %s", se.Key, err.Error())
		return false
	}

	if w.entry > 0 {
		w.jsonBuilder.WriteString(",")
	}
	if w.writeoption == WriteJSONandJSONLoaderGoCode {
		readonly := "T"
		if !se.ReadOnly {
			readonly = "F"
		}
		w.jsonBuilder.WriteString(fmt.Sprintf("\"%s\":\"%s|%s\"", se.Key, readonly, se.SQL))
	} else {
		w.jsonBuilder.WriteString(json)
	}
	w.entry++

	return true
}

// writeCode create code block for given sqlEntry and write to builder
func (w *FileWriter) writeCode(se *sqlEntry) {
	w.writeCodeComment(se)

	if w.writeoption == WriteJSONandJSONLoaderGoCode {
		w.codeBuilder.WriteString(fmt.Sprintf(`func %s() Statement {
	sql, ok := _jsonsqls["%s"]
	if !ok {
		return Statement{}
	}
	return parseStmt(sql)
}`, se.Key, se.Key))
		w.codeBuilder.WriteString("\n\n")
	} else {
		w.codeBuilder.WriteString("const " + se.Key + " string = \"" + se.SQL + "\"\n\n")
	}
}

//writeCodeComment writes comments for code in builder
func (w *FileWriter) writeCodeComment(se *sqlEntry) {
	w.codeBuilder.WriteString("//" + se.Description + "\n")

	w.codeBuilder.WriteString("//\n")
	w.codeBuilder.WriteString("//Fields: " + strconv.Itoa(se.FieldsCount))
	w.codeBuilder.WriteString(", Parameters: " + strconv.Itoa(se.ParamCount) + "\n")

	if len(se.Fields) > 0 {
		w.codeBuilder.WriteString("//\n")
		w.codeBuilder.WriteString("//  Fields: " + se.Fields + "\n")
	}

	if len(se.ParamFields) > 0 {
		w.codeBuilder.WriteString("//\n")
		w.codeBuilder.WriteString("//  ParamFields: " + se.ParamFields + "\n")
	}

	if len(se.ReturningFields) > 0 {
		w.codeBuilder.WriteString("//\n")
		w.codeBuilder.WriteString("//  ReturningFields: " + se.ReturningFields + "\n")
	}

	if w.writeoption == WriteJSONandJSONLoaderGoCode {
		w.codeBuilder.WriteString("//SQL:\n")
		w.codeBuilder.WriteString("//  " + se.SQL + "\n")
	}
}
