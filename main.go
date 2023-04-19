package main

import (
	"context"
	"fmt"
	"github.com/onlyzzg/gentol/src/config"
	"github.com/onlyzzg/gentol/src/datasource"
	gormx "github.com/onlyzzg/gentol/src/gormx"
	"strings"
)

func main() {
	tableConfigs := config.TableConfigs
	for _, database := range tableConfigs.Configs {
		dbConfig := &gormx.Config{DBName: database.DBName}
		dbConfig.DSN = database.DSN
		dbConfig.DBType = gormx.DBType(database.DBType)
		err := gormx.InitConfig(dbConfig)
		if err != nil {
			panic(err)
		}
		ds, err := datasource.GetDS(gormx.DBType(database.DBType))
		if err != nil {
			panic(err)
		}
		for _, tableInfo := range database.Tables {
			tableColMap, getErr := ds.GetColumnsUnderTable(context.Background(), dbConfig.DBName, tableInfo.SchemaName, []string{tableInfo.TableName})
			if getErr != nil {
				panic(getErr)
			}
			for _, info := range tableColMap {
				tableName := info.TableName
				columnInfoList := info.ColumnInfoList
				fmt.Println("tableName: " + tableName)
				fmt.Print("columnInfoList: ")
				columns := make([]string, 0)
				for _, columnInfo := range columnInfoList {
					columns = append(columns, columnInfo.ColumnName)
				}
				fmt.Println(strings.Join(columns, ","))
			}
		}
	}
}
