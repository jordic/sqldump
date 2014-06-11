# sqldump

A package for dumping database contents. For the momment only MySQL driver had been developed. Perhaps more in the future.
I get the idea from https://github.com/hgfischer/mysqlsuperdump start hacking the code as learning it, and finally, I release as a package.

Feel free to contact me, or send pull requests if this package is util for someone. I use, sqldumper.go as a tool for backing up some databases...

## Install

go get github.com/jordic/sqldump

## Example

Simple example dumping one table from a MySQL database

```go
    // dumping all tables to stdout
    dumper := sqldump.NewMySQLDump(db, os.Stdout)
    DumpAllTables(dumper)

    dumper.DumpInit()
    // dumping just table test structure
    dumper.DumpCreateTable("test")
    // dumping table test content
    dumper.DumpTableData("test")
    dumper.DumpEnd()

    // w.. is a io.Writer, and it can works with 


```


## Examples in apps

### sqldumper.go
Using a json config file, describing all your databases it can download all of them, or just one.

Params:
- config="jsonfile" defaults to sqldumper.json
- output="directory where to output in format /tmp/"
- gzip default false, put the flag if want to dump to gzip file

JsonFile:
```json
{
    "test1":"root:@tcp(127.0.0.1:3306)/test1?charset=utf8",
    "test2":"root:@tcp(127.0.0.1:3306)/test2?charset=utf8"
}
```
Where each key represents a database, and each value, a dsn string for connecting to it

Install:

go get github.com/jordic/sqldump<br/>
go run apps/sqldumper/sqldumper.go<br/> 
or
go build apps/sqldumper/sqldumper.go

### webdumper.go 
Demostrates the use of the dumper as a web dumper..

#### Install:
go get github.com/jordic/sqldump
go run apps/webdumper.go

Fires an http server on port :3000, default, shows the form for
input params... 
if param action="dowload" as a hidden or get, param... 
dumps the requested database



