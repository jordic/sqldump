package main

import (
	"compress/gzip"
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jordic/sqldump"
	"io"
	"io/ioutil"
	"os"
)

// version app number
var VERSION string = "0.2"

var (
	config   = flag.String("config", "sqldumper.json", "path to json config file")
	version  = flag.Bool("version", false, "print version number")
	output   = flag.String("output", "", "Output dir of mysql dumps")
	usegz    = flag.Bool("gzip", false, "Output in gzip format")
	database = flag.String("db", "", "Which Database to download. Default all databases")
	dsn      map[string]string
	w        io.WriteCloser
)

func DownloadDB(datab string, dsn string) {

	var filename string
	//var w interface{}

	bdd, err := sql.Open("mysql", dsn)
	if err != nil {
		fmt.Printf("Error on database %s", err)
		return
	}
	defer bdd.Close()

	// check if dsn is valid..
	err = bdd.Ping()
	if err != nil {
		fmt.Printf("Can't connect %s\n", err)
		return

	}
	if *usegz == false {
		filename = fmt.Sprintf("%s%s.sql", *output, datab)
	} else {
		filename = fmt.Sprintf("%s%s.sql.gz", *output, datab)
	}

	f, err := os.Create(filename)
	if err != nil {
		fmt.Printf("Error creating file %s\n", err)
	}
	defer f.Close()

	w = f
	if *usegz {
		w = gzip.NewWriter(w)
		defer w.Close()
	}

	dumper := sqldump.NewMySQLDump(bdd, w)
	tables, err := dumper.GetTables()
	if err != nil {
		fmt.Printf("ERROR: %s\n", err)
		return
	}

	// adds utf variable and no constrains checks
	dumper.DumpInit()

	for _, table := range tables {
		err = dumper.DumpCreateTable(table)
		if err != nil {
			fmt.Println(err)

		}
		err = dumper.DumpTableData(table)
		if err != nil {
			fmt.Println(err)
		}
	}
	// enables constrain checks again
	dumper.DumpEnd()

	fmt.Printf("Database %s succefull dumped to %s\n", datab, filename)

}

func main() {

	flag.Parse()

	if *version {
		fmt.Printf("Version %s\n", VERSION)
		return
	}

	// load config file
	file, err := ioutil.ReadFile(*config)
	if err != nil {
		fmt.Printf("%s. Confing file not found.\n", *config)
		return
	}

	err = json.Unmarshal(file, &dsn)
	if err != nil {
		fmt.Printf("Invalid config file")
		return
	}

	if *output != "" {
		if _, err := os.Stat(*output); os.IsNotExist(err) {
			fmt.Printf("No such file or directory: %s\n", *output)
			return
		}
	}

	if *database == "" {
		for k, v := range dsn {
			DownloadDB(k, v)
		}
	} else {
		if val, ok := dsn[*database]; ok {
			fmt.Printf("Dumping database .. %s\n", *database)
			DownloadDB(*database, val)
		} else {
			fmt.Printf("Dumping %s not found in server \n", *database)
			return
		}
	}

}
