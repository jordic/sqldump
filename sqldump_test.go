package sqldump

import (
	"database/sql"
	"os"
	"testing"
	//"strings"
	"bytes"
	"fmt"
	_ "github.com/go-sql-driver/mysql" // import mysql driver
	"net"
	"strings"
	//"io"
)

var (
	user    string
	pass    string
	prot    string
	addr    string
	dbname  string
	dsn     string
	netAddr string
)

func init() {

	env := func(key, defaultValue string) string {
		if value := os.Getenv(key); value != "" {
			return value
		}
		return defaultValue
	}

	user = env("MYSQL_TEST_USER", "root")
	pass = env("MYSQL_TEST_PASS", "")
	prot = env("MYSQL_TEST_PROT", "tcp")
	addr = env("MYSQL_TEST_ADDR", "localhost:3306")
	dbname = env("MYSQL_TEST_DBNAME", "gotest")
	netAddr = fmt.Sprintf("%s(%s)", prot, addr)
	dsn = fmt.Sprintf("%s:%s@%s/%s?timeout=30s&strict=true", user, pass, netAddr, dbname)
	c, err := net.Dial(prot, addr)
	if err == nil {
		//available = true
		c.Close()
	}

}

type DBTest struct {
	*testing.T
	db *sql.DB
}

var CREATE_TABLE = "CREATE TABLE `%s` (\n  `id` int(11) unsigned NOT NULL AUTO_INCREMENT,\n  `test` varchar(255) DEFAULT NULL,\n  `name` varchar(255) DEFAULT NULL,\n  PRIMARY KEY (`id`)\n) ENGINE=MyISAM DEFAULT CHARSET=utf8;"

func runTests(t *testing.T, dsn string, tests ...func(dbt *DBTest)) {

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		t.Fatalf("Error connecting %s", dsn)
	}
	defer db.Close()

	db.Exec("DROP TABLE IF EXISTS test")
	dbt := &DBTest{t, db}
	for _, test := range tests {
		test(dbt)
		dbt.db.Exec("DROP TABLE IF EXISTS test")
		dbt.db.Exec("DROP TABLE IF EXISTS test2")
	}

}

func createTableTest(dbt *DBTest, id string) error {

	query := fmt.Sprintf(CREATE_TABLE, id)

	_, err := dbt.db.Exec(query)

	if err != nil {
		return err
	}

	return nil
}

func getDmapInstance(db *sql.DB) MySQLDump {

	dial := MySQLDump{db: db, w: os.Stdout}
	return dial

}

func TestGetTables(t *testing.T) {

	runTests(t, dsn, func(dbt *DBTest) {

		createTableTest(dbt, "test")
		dmap := getDmapInstance(dbt.db)
		tab, err := dmap.GetTables()

		if err != nil {
			panic(err)
		}

		fmt.Println(strings.Join(tab, ", "))
		if len(tab) != 1 {
			t.Errorf("%v!=%v", 1, len(tab))
		}

		if tab[0] != "test" {
			t.Errorf("tab name not equal")
		}

	})

}

func TestColumnList(t *testing.T) {
	runTests(t, dsn, func(dbt *DBTest) {
		createTableTest(dbt, "test")
		dmap := getDmapInstance(dbt.db)
		columns, err := dmap.GetColumnsFromTable("test")
		if err != nil {
			panic(err)
		}
		if len(columns) != 3 {
			t.Errorf("%v!=%v", 1, len(columns))
		}
	})
}

func TestCreateTable(t *testing.T) {
	runTests(t, dsn, func(dbt *DBTest) {
		createTableTest(dbt, "test")
		b := &bytes.Buffer{}
		c := &bytes.Buffer{}

		dumper := MySQLDump{db: dbt.db, w: c}
		dumper.DumpCreateTable("test")
		table := "test"
		fmt.Fprintf(b, "\n--\n")
		fmt.Fprintf(b, "-- Structure for table `%s`\n", table)
		fmt.Fprintf(b, "--\n\n")
		fmt.Fprintf(b, "DROP TABLE IF EXISTS `%s`;\n", table)
		fmt.Fprintf(b, CREATE_TABLE, table)
		fmt.Fprintf(b, "\n")

		res := bytes.Compare(b.Bytes(), c.Bytes())
		if res != 0 {
			t.Errorf("%s!=%s", b, c)
		}

	})
}

func TestGetSelectQueryFor(t *testing.T) {
	runTests(t, dsn, func(dbt *DBTest) {
		createTableTest(dbt, "test")
		dumper := getDmapInstance(dbt.db)
		s, err := getSelectQueryFor(dumper, "test")

		if err != nil {
			panic(err)
		}

		QU := "SELECT `id`, `test`, `name` FROM test"

		if s != QU {
			t.Errorf("%s!=%s", s, QU)
		}

	})
}

func TestDumpTableData(t *testing.T) {
	runTests(t, dsn, func(dbt *DBTest) {
		createTableTest(dbt, "test")
		q := "INSERT INTO `test` (`id`, `test`, `name`)\nVALUES\n"
		// add two rows
		dbt.db.Exec(fmt.Sprintf("%s('1', 'test', 'name');", q))
		dbt.db.Exec(fmt.Sprintf("%s('2', 'test', 'name');", q))

		b := &bytes.Buffer{}
		c := &bytes.Buffer{}

		dumper := MySQLDump{db: dbt.db, w: c}

		row := dbt.db.QueryRow("SELECT count(*) from `test`")
		var r uint64
		err := row.Scan(&r)

		if r != 2 {
			t.Errorf("%v!=%v", r, 2)
		}

		err = dumper.DumpTableData("test")
		if err != nil {
			panic(err)
		}

		fmt.Fprintf(b, "\n--\n-- Data for table `test`\n")
		fmt.Fprintf(b, "LOCK TABLES `test` WRITE;\n")
		fmt.Fprintf(b, "INSERT INTO `test` (`id`, `test`, `name`)\nVALUES\n")
		fmt.Fprintf(b, "( '1', 'test', 'name' ),\n")
		fmt.Fprintf(b, "( '2', 'test', 'name' );\n")

		res := bytes.Compare(b.Bytes(), c.Bytes())
		if res != 0 {
			t.Errorf("%s!=%s", b, c)
		}

	})
}
func TestDumpDataEmptyTable(t *testing.T) {
	runTests(t, dsn, func(dbt *DBTest) {

		createTableTest(dbt, "test")

		b := &bytes.Buffer{}
		c := &bytes.Buffer{}

		table := "test"

		dumper := MySQLDump{db: dbt.db, w: c}
		err := dumper.DumpTableData("test")
		if err != nil {
			panic(err)
		}

		fmt.Fprintf(b, "\n--\n-- Data for table `%s`\n", table)
		fmt.Fprintf(b, "\n--Empty table\n")

		res := bytes.Compare(b.Bytes(), c.Bytes())
		if res != 0 {
			t.Errorf("%s!=%s", b, c)
		}

	})

}

func TestNoTablesInDatabase(t *testing.T) {
	runTests(t, dsn, func(dbt *DBTest) {

		dumper := getDmapInstance(dbt.db)
		tables, err := dumper.GetTables()

		if len(tables) != 0 {
			t.Error("No tables for dumping error detected")
		}

		_, err = dumper.GetColumnsFromTable("test")
		if strings.Contains(err.Error(), "Error 1146") != true {
			t.Error("Wrong error number")
		}

		err = dumper.DumpCreateTable("test")
		if strings.Contains(err.Error(), "Error 1146") != true {
			t.Error("Wrong error number")
		}

	})

}

func TestDumpAllTables(t *testing.T) {
	runTests(t, dsn, func(dbt *DBTest) {
		createTableTest(dbt, "test")
		createTableTest(dbt, "test2")
		q := "INSERT INTO `test` (`id`, `test`, `name`)\nVALUES\n"
		dbt.db.Exec(fmt.Sprintf("%s('1', 'test', 'name');", q))
		dbt.db.Exec(fmt.Sprintf("%s('1', 'test2', 'name');", q))

		c := &bytes.Buffer{}
		dumper := MySQLDump{db: dbt.db, w: c}

		err := DumpAllTables(dumper)
		if err != nil {
			panic(err)
		}

		d := c.String()

		//fmt.Println(d)

		if strings.Contains(d, "CREATE TABLE `test`") != true {
			t.Error("Incorrect dump")
		}

		if strings.Contains(d, "CREATE TABLE `test2`") != true {
			t.Error("Incorrect dump")
		}

		if strings.Contains(d, "INSERT INTO `test` (`id`") != true {
			t.Error("Incorrect dump")
		}

		if strings.Contains(d, "DROP TABLE IF EXISTS `test`;") != true {
			t.Error("Incorrect dump")
		}

		if strings.Contains(d, "INSERT INTO `test2`") != false {
			t.Error("Incorrect dump")
		}

	})
}
