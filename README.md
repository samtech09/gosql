# gosql

[![Documentation](https://godoc.org/github.com/samtech09/gosql?status.svg)](http://godoc.org/github.com/samtech09/gosql)

SQL builder with `GO` code generation.


## Why another sql builder
`gosql` is not just another sql generator, instead it helps developers to visualize SQLs that they are going to call from their application.

`gosql` allows to generate `GO` code for generated SQLs that can be embeded into application or can also be side-loaded through JSON file for keeping SQLs out from application code.

The benefit of both the approaches (embedding or side-loading) is that, it makes is extremely easy to call that SQL by just it's name. It makes code clean and easy to manage.

```
...
rows, err := db.Query(sqls.UserInsert, name, pwd)
...
```

### Visualize SQLs
The most wonderful feature for which i wrote this tool is to visualize SQL in editor while coding. Just hover over the name of sql, it will give you details of fields that it will return/select, parameters that developer has to pass to execute that SQL alongwith the actual SQL too.


![SQL details in popup](doc/loader-info.png?raw=true)



## Features
- Fluent style syntax
- Generate `SELECT`, `INSERT`, `UPDATE` and `DELETE` SQLs
- Support for sub SQLs
- Groupby and OrderBy supported
- **Visualize SQL while coding**
- Generate PostgreSQL, MySQL and MS-Sql friendly SQLs
- Add rowcount with result to allow developer efficiently create slice with exact capacity during scanning to avoid repetitive allocations
- Provide ReadOnly flag with statement so developer could choose to run SQL on master or on read-only replica.


## Installation
`go get github.com/samtech09/gosql`


## Usage
There are two use cases
1. Generate inplace code: generate code right before executing SQL
2. Generate, write/export to `GO` code and/or JSON

### Generating inplace code
Use builder to create statement, finally call `Build()` function to generate statement.

```
package main

import (
	"fmt"

	gs "github.com/samtech09/gosql"
)

func main() {
	// build sql
	stmt := gs.SelectBuilder().Select("q.ID", "qd.Title").
		From("Questions", "q").
		From("QuestionData", "qd").
		Where(gs.C().EQ("q.ID", "qd.QID"),
			gs.C().EQ("q.TopicID", "$1")).
		OrderBy("qd.QID", true).
		RowCount().
		Build(true)

	// Execute sql
	ExecuteQuery(stmt.SQL, "Testuser", "22")
}

func ExecuteQuery(sql string, param ...interface{}) {
	// code to execute SQL
	//
	fmt.Println(sql)
}
```

### Generate, write to `GO` code and/or JSON
The basic flow is
- Create a cmd tool
- Write code to build SQLs through `gosql` and export to files
- Embed `GO` code in application

Create a go file either inside you application hierarchy or somewhere else, as it will be just to generate `GO` code and/JSON with SQLs, so it doesn't matter wheter you genertate it inside you application structure or somewhere else.

Lets start with following application structure

```
Gocode
    |-sqlgenerator  (folder)
    |-sqls          (folder, generated code will be placed here)
    |-main.go       (application entry point)
```

Create new file `generator.go` inside `sqlgenerator` folder and put following code there

```
package main

import (
	sb "github.com/samtech09/gosql"
)

func main() {
	fw := sb.NewFileWriter(5)

	stmt := sb.InsertBuilder().Table("users").
		Columns("name", "age").Returning("id").
		Build(true)
	// queue for writing
	fw.Queue(stmt, "user", "Create", "Create new user.")

	stmt = sb.SelectBuilder().Select("q.ID", "qd.Title").
		From("Questions", "q").
		From("QuestionData", "qd").
		Where(sb.C().EQ("q.ID", "qd.QID"), sb.C().EQ("q.TopicID", "21")).
		OrderBy("qd.QID", true).
		RowCount().
		Build(true)
	// queue for writing
	fw.Queue(stmt, "ques", "listForDD", "Gives list of question ID and Title only to fill dropdowns.")

	// Write as GO code to ../sqls folder
    	//  exported filename = sqlbuilder
    	//  exported gocode package = sqls
	fw.Write("../sqls", "sqlbuilder", "sqls", sb.WriteGoCode)
}
```

Run it to build and generate code

`go run generator.go`

It will generate `sqlbuilder.go` file inside `sqls` folder. Now Project structure should be like below

```
Gocode
    |-sqlgenerator
        |-generator.go
    |-sqls
        |-sqlbuilder.go
    |-main.go
```

Now add code in `main.go` to use those generated SQLs

```
package main

import (
	"fmt"

	"github.com/samtech09/gosql/Examples/Gocode/sqls"
)

func main() {
	// Execute sql
	ExecuteQuery(sqls.UserCreate, "Testuser", "22")

	ExecuteQuery(sqls.QuesListForDD, nil)
}

func ExecuteQuery(sql string, param ...interface{}) {
	// code to execute SQL
	// ...
	fmt.Println(sql)
}
```

After executing main.go you will see following output

```
$ go run .
insert into users(name, age) values($1, $2) returning id;
select q.ID, qd.Title, count(*) over() as rowcount from questions q, questiondata qd where (q.ID=qd.QID and q.TopicID=21) order by qd.QID desc;
```

If you are using editor which support displaying relevant section of documentation (like VSCode), you could see detail that you need to call a particular SQL.

Hover mouse over `sqls.UserCreate`, and you will see SQL details as popup-info

![sqls.UserCreate details in popup](doc/const-info.png?raw=true)


For more details view [Examples](https://github.com/samtech09/gosql/tree/master/Examples).


## Setting Database Type and paramer format to generate supported SQL
`gosql` support to generated SQLs for `PostgreSQL`, `Ms-SQL` and `MySQL`. It can be set by environment variable `DATABASE_TYPE`.

It can be set right before generating SQL as below

```
os.Setenv("DATABASE_TYPE", DbTypePostgreSQL)
or
os.Setenv("DATABASE_TYPE", DbTypeMySQL)
or
os.Setenv("DATABASE_TYPE", DbTypeMsSQL)

...
...

sql := SelectBuilder().From(...)
```

By default `gosql` will use following paramter format for generating sqls

Database Type | Parameter format
------------- | ----------------
PostgreSQL | `$1, $2, ...`
MsSQL | `@p1, @p2, ...`
MySQL | `?, ?, ...`

<br />

Parameter character can be overwritten by setting following environment variables

Database Type | Parameter format
------------- | ----------------
PARAM_CHAR | Overwrite paramter string for current DATABASE_TYPE. <br />e.g.<br />`os.Setenv("PARAM_CHAR", "$p)`
PARAM_APPEND_NUMBER | Set it to `1` to enable appending sequence number to parameters e.g. `$1, $2, ...`. To disable set to '0'


<br />

Feedback and suggestions are always welcomed.

<br />

## TODO
- ~~Add support to genrate SQLs for `ms-sql`~~ [ support added ]
- ~~Add support to generate SQLs for `mysql`~~ [ support added ]

