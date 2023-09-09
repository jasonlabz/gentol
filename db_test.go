package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/jasonlabz/gentol/datasource"
	"github.com/jasonlabz/gentol/gormx"
	"github.com/jasonlabz/gentol/metadata"
	"log"
	"reflect"
	"regexp"
	"runtime"
	"strings"
	"testing"
	"time"
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
	columnTypes, err := db.Migrator().ColumnTypes("test.user1")
	indexes, err := db.Migrator().GetIndexes("user1")
	if err != nil {
		panic(err)
	}
	fmt.Println(indexes)
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
	inValues := []bool{true, true, false}
	bytes, _ := json.Marshal(inValues)
	inValues1 := []int32{1, 3, 5}
	bytes1, _ := json.Marshal(inValues1)
	res := string(bytes)
	res1 := string(bytes1)
	fmt.Print(res)
	fmt.Print(res1)
}
func TestStruct(t *testing.T) {
	user := struct{}{}
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
func Values(value any) string {
	switch value.(type) {
	case int, int8, int16, int32, int64, bool, float32, float64:
		return fmt.Sprintf("%v", value)
	default:
		return fmt.Sprintf("'%v'", value)
	}
}

func TransInCondition[T any](prefix string, values []T) string {
	res := make([]string, 0)
	numbers := len(values) / 1000
	for i := 0; i < numbers; i++ {
		items := make([]string, 0)
		for j := i * 1000; j < (i+1)*1000; j++ {
			items = append(items, Values(values[j]))
		}
		res = append(res, fmt.Sprintf("%s (%s)", prefix, strings.Join(items, ",")))
	}
	items := make([]string, 0)
	for i := numbers * 1000; i < numbers*1000+len(values)%1000; i++ {
		items = append(items, Values(values[i]))
	}
	res = append(res, fmt.Sprintf("%s (%s)", prefix, strings.Join(items, ",")))
	return strings.Join(res, " or ")
}

func TestTransInCondition(t *testing.T) {
	inValues := []bool{true, true, false}
	inValues0 := []int32{}
	inValues1 := []float64{1.0, 3.0, 5.0}
	inValues2 := []string{"这种短手", "sdasdsd", "sdasdsd", "sdasdsd", "sad"}
	inValues3 := []time.Time{time.Now(), time.Now(), time.Now(), time.Now(), time.Now()}
	condition := TransInCondition("name in", inValues)
	condition1 := TransInCondition("name in", inValues0)
	condition2 := TransInCondition("name in", inValues1)
	condition3 := TransInCondition("name in", inValues2)
	condition4 := TransInCondition("name in", inValues3)
	fmt.Print(condition)
	fmt.Print(condition1)
	fmt.Print(condition2)
	fmt.Print(condition3)
	fmt.Print(condition4)
}

func TestListDir(t *testing.T) {
	version := runtime.Version()
	reg := regexp.MustCompile(`(\d+\.\d+\.*\d*)`)
	if reg == nil {
		log.Panicln("正则表达式解析失败")
	}
	versionStr := reg.FindString(version)
	versionSlice := strings.Split(versionStr, ".")
	fmt.Print(versionStr)
	fmt.Print(versionSlice)
}
