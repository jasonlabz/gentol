package main

import (
	"context"
	"fmt"
	"github.com/jasonlabz/gentol/dal/db/model"
	"github.com/jasonlabz/gentol/datasource"
	"github.com/jasonlabz/gentol/gormx"
	"github.com/jasonlabz/gentol/metadata"
	"reflect"
	"runtime"
	"strings"
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
func TestPostgresOperator(t *testing.T) {
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
	db, err := gormx.GetDB(dbInfo.DBName)
	if err != nil {
		panic(err)
	}
	columnTypes, err := db.Migrator().ColumnTypes("public.user")
	if err != nil {
		panic(err)
	}
	fmt.Println(columnTypes)
}

func TestDemo(t *testing.T) {
	baseName := metadata.GetFuncNamePath(metadata.LoadTpl)
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
func TestStruct(t *testing.T) {
	user := model.User{}
	val := reflect.ValueOf(user)

	fmt.Println(val.Type().PkgPath())
	for i := 0; i < val.Type().NumField(); i++ {
		fmt.Println("gorm:" + val.Type().Field(i).Tag.Get("gorm"))
		column := ""
		gormValList := strings.Split(val.Type().Field(i).Tag.Get("gorm"), ";")
		for _, item := range gormValList {
			if strings.Contains(item, "column") {
				column = strings.Split(item, ":")[1]
			}
		}
		fmt.Println("column:" + column)
		fmt.Println("json:" + val.Type().Field(i).Tag.Get("json"))
	}
}

func test() string {
	_, filename, _, _ := runtime.Caller(2)
	fmt.Print(filename)
	return filename
}
