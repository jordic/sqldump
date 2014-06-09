package sqldump

import (
	"bytes"
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql" // import mysql driver
	"io"
	"strings"
)

// SQLDumper is an interface that can be implemented by
// and used by apps, that will dump database contents
// to io.Writer
type SQLDumper interface {
	GetTables() ([]string, error)
	GetColumnsFromTable(table string) ([]string, error)
	DumpCreateTable(table string) error
	DumpTableData(table string) error
}

// MySQLDump is actually, the only dumper released.. perhaps
// will build new ones on future.
// to init it, just, create a new type:
//      dumper := MysqlDump{db:sql.DB, w:io.Writer}
//  and use it.
// on apps directory i will build some examples...
// a commandline tool for downloading databases from server...
// a web tool, for the same prupose..
type MySQLDump struct {
	db *sql.DB
	w  io.Writer
}

// GetTables returns the list of tables on the database
// @todo: return views also
func (t MySQLDump) GetTables() ([]string, error) {

	var tables = make([]string, 0)
	rows, err := t.db.Query("SHOW FULL TABLES")
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var tableName string
		var tableType string
		err = rows.Scan(&tableName, &tableType)
		if err != nil {
			return nil, err
		}
		if tableType == "BASE TABLE" {
			tables = append(tables, tableName)
		}

	}

	if rows.Err() != nil {
		return nil, rows.Err()
	}

	return tables, nil

}

// GetColumnsFromTable returns the column names of a give table
func (t MySQLDump) GetColumnsFromTable(table string) ([]string, error) {
	var columns = make([]string, 0)
	query := fmt.Sprintf("DESCRIBE %s", table)
	rows, err := t.db.Query(query)
	if err != nil {
		return nil, err
	}

	var (
		field string
		tipo  string
		null  sql.NullString
		key   sql.NullString
		def   sql.NullString
		extra sql.NullString
	)

	for rows.Next() {
		err = rows.Scan(&field, &tipo, &key, &null, &def, &extra)
		if err != nil {
			return nil, err
		}
		columns = append(columns, field)
	}

	if rows.Err() != nil {
		return nil, rows.Err()
	}

	return columns, nil
}

// DumpCreateTable dumps sql table structure to w struct
func (t MySQLDump) DumpCreateTable(table string) error {
	row := t.db.QueryRow(fmt.Sprintf("SHOW CREATE TABLE `%s`", table))
	var tname, ddl string
	err := row.Scan(&tname, &ddl)
	if err != nil {
		return err
	}
	fmt.Fprintf(t.w, "\n--\n")
	fmt.Fprintf(t.w, "-- Structure for table `%s`\n", table)
	fmt.Fprintf(t.w, "--\n\n")
	fmt.Fprintf(t.w, "DROP TABLE IF EXISTS `%s`;\n", table)
	fmt.Fprintf(t.w, "%s;\n", ddl)
	return nil
}

// DumpTableData will dump data for a given table
func (t MySQLDump) DumpTableData(table string) error {
	fmt.Fprintf(t.w, "\n--\n-- Data for table `%s`\n", table)
	// check if table has data..
	row := t.db.QueryRow(fmt.Sprintf("SELECT COUNT(*) from `%s`", table))

	var ct uint64
	err := row.Scan(&ct)
	if err != nil {
		return err
	}
	if ct == 0 {
		fmt.Fprintf(t.w, "\n--Empty table\n")
		return nil
	}

	fmt.Fprintf(t.w, "LOCK TABLES `%s` WRITE;\n", table)
	//var columns []string
	cols, err := t.GetColumnsFromTable(table)
	if err != nil {
		return err
	}
	collist := buildColumnList(cols)
	query := fmt.Sprintf("INSERT INTO `%s` (%s)\nVALUES\n", table, collist)
	fmt.Fprint(t.w, query)

	rows, derr := t.db.Query(fmt.Sprintf("SELECT * FROM `%s`", table))
	if derr != nil {
		return derr
	}

	columns, err := rows.Columns()
	if err != nil {
		return err
	}

	values := make([]*sql.RawBytes, len(columns))
	scanArgs := make([]interface{}, len(values))
	for i := range values {
		scanArgs[i] = &values[i]
	}

	var data string
	for rows.Next() {
		// if data is present write to Writer
		if data != "" {
			fmt.Fprint(t.w, fmt.Sprintf("%s,\n", data))
		}

		err = rows.Scan(scanArgs...)
		if err != nil {
			return err
		}
		var vals []string
		for _, col := range values {
			val := "NULL"
			if col != nil {
				val = fmt.Sprintf("'%s'", escape(string(*col)))
			}
			vals = append(vals, val)
		}
		// write line to writer
		data = fmt.Sprintf("( %s )", strings.Join(vals, ", "))
	}
	// write last row
	fmt.Fprint(t.w, fmt.Sprintf("%s;\n", data))
	return nil
}

func getSelectQueryFor(t MySQLDump, table string) (string, error) {
	columns, err := t.GetColumnsFromTable(table)
	if err != nil {
		return "", err
	}

	//out := "`" + strings.Join(columns, "`, `") + "`"
	out := buildColumnList(columns)
	sql := fmt.Sprintf("SELECT %s FROM %s", out, table)
	return sql, nil
}

func buildColumnList(columns []string) string {
	out := "`" + strings.Join(columns, "`, `") + "`"
	return out
}

func escape(str string) string {
	var esc string
	var buf bytes.Buffer
	last := 0
	for i, c := range str {
		switch c {
		case 0:
			esc = `\0`
		case '\n':
			esc = `\n`
		case '\r':
			esc = `\r`
		case '\\':
			esc = `\\`
		case '\'':
			esc = `\'`
		case '"':
			esc = `\"`
		case '\032':
			esc = `\Z`
		default:
			continue
		}
		io.WriteString(&buf, str[last:i])
		io.WriteString(&buf, esc)
		last = i + 1
	}
	io.WriteString(&buf, str[last:])
	return buf.String()
}

// DumpAllTables will dump a database, content and structure to io.Writer
func DumpAllTables(t MySQLDump) error {

	tables, err := t.GetTables()
	if err != nil {
		return err
	}

	for _, table := range tables {
		err = t.DumpCreateTable(table)
		if err != nil {
			return err
		}
		err = t.DumpTableData(table)
		if err != nil {
			return err
		}
	}

	return nil

}
