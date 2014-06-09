

# sqldump

A package for dumping database contents. For the momment only MySQL driver had been developed. Perhaps more in the future

## Install

go get github.com/jordic/sqldump

## Example

Simple example dumping one table from a MySQL database

```go
    // dumping all tables to stdout
    dumper := sqldump.NewMySQLDump(db, os.Stdout)
    DumpAllTables(dumper)

    // dumping just table test structure
    dumper.DumpCreateTable("test")

    // dumping table test content
    dumper.DumpTableData("test")

    // w.. is a io.Writer, and it can works with 


```

## Examples in apps

### webdumper.go 
Demostrates the use of the dumper as a web dumper..

#### Install:
go get github.com/jordic/sqldump
go run apps/webdumper.go

Fires an http server on port :3000, default, shows the form for
input params... 
if param action="dowload" as a hidden or get, param... 
dumps the requested database

