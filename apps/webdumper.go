package main

// This is an app demostrating the functionalitis of
// SQLDumper
// Listen on port 3000 accepts a dns / database / table form, and
// dumps the corresponding sql.

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jordic/sqldump"
	"html/template"
	"log"
	"net/http"
)

var (
	server   string
	user     string
	password string
	db       string
	table    string
)

func main() {

	mux := http.NewServeMux()
	mux.Handle("/", http.HandlerFunc(handleRequest))
	log.Println("Server listening on :3000")
	http.ListenAndServe(":3000", mux)

}

func handleRequest(w http.ResponseWriter, r *http.Request) {

	if r.FormValue("action") == "download" {

		// Read form values @todo validate them
		server = r.FormValue("server")
		user = r.FormValue("user")
		password = r.FormValue("password")
		db = r.FormValue("db")
		table = r.FormValue("table")

		// create dsn connection for mysql server
		dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s?timeout=30s&strict=true", user,
			password, server, db)

		db, err := sql.Open("mysql", dsn)
		if err != nil {
			fmt.Fprintf(w, "Unable to connect to database", err)
		}

		// generate a new instance of MysqlDumper
		dumper := sqldump.NewMySQLDump(db, w)

		// dumps create table
		err = dumper.DumpCreateTable(table)
		if err != nil {
			fmt.Fprintf(w, "Error dumping database", err)
			return
		}
		// dumps table data
		err = dumper.DumpTableData(table)
		if err != nil {
			fmt.Fprintf(w, "Error dumping database", err)
			return
		}

		return

	}

	t := template.Must(template.New("listing").Parse(TPL))
	v := map[string]interface{}{
		"Test": "test",
	}
	t.Execute(w, v)

}

const TPL = `
    <html>
        <head>
        <title>Mysqldumper</title>

        </head>
        <body>
        <form action="" method="POST" >
            <input type="hidden" name="action" value="download" /><br />
            <label for="id_server">Server Name: localhost:3306</label><br />
            <input type="text" name="server" value="localhost:3306" id="id_server" /><br />
            <label for="id_user">Username</label><br />
            <input type="text" name="user" value="root" id="id_user" /><br />
            <label for="id_password">Password</label><br />
            <input type="text" name="password" value="" id="id_password" /><br />
            <label for="id_db">Database</label><br />
            <input type="text" name="db" value="" id="id_db" /><br />
            <label for="id_db">Table</label><br />
            <input type="text" name="table" value="" id="id_table" /><br />
            
            <button type="submit">Download</button>

        </form>
        </body>
    </html>`
