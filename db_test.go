package main

import (
	"context"
	"fmt"
	"github.com/onlyzzg/dbutil/datasource"
	"github.com/onlyzzg/dbutil/gormx"
	"path/filepath"
	"runtime"
	"testing"
)

func TestNewOperator(t *testing.T) {
	dbInfo := &gormx.Config{
		DBName:   "master",
		DBType:   gormx.DBTypeGreenplum,
		Host:     "127.0.0.1",
		Port:     8432,
		User:     "postgres",
		Password: "halojeff",
		Database: "lg_server",
	}
	err := gormx.InitConfig(dbInfo)
	if err != nil {
		panic(err)
	}
	ds, _ := datasource.GetDS(gormx.DBTypeGreenplum)
	tableMap, err := ds.GetTablesUnderDB(context.Background(), dbInfo.DBName)
	colMap, err := ds.GetColumns(context.Background(), dbInfo.DBName)
	tableColMap, err := ds.GetColumnsUnderTable(context.Background(), dbInfo.DBName, "public", []string{"user"})
	fmt.Print(tableMap)
	fmt.Print(colMap)
	fmt.Print(tableColMap)
}

func TestDemo(t *testing.T) {
	baseName := filepath.Base("./hello.go")
	t1, filename, t2, ok := runtime.Caller(0)
	fmt.Print(t1)
	fmt.Print(t2)
	fmt.Print(ok)
	//fmt.Print(ok)
	fmt.Print(baseName)
	fmt.Print(filename)
	s := test()
	fmt.Print(s)
}

func test() string {
	_, filename, _, _ := runtime.Caller(2)
	fmt.Print(filename)
	return filename
}
