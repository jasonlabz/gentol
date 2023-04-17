package main

import (
	"context"
	"fmt"
	"github.com/onlyzzg/gentol/datasource"
	"github.com/onlyzzg/gentol/gormx"
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
	err := gormx.InitWithConfig(dbInfo)
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
