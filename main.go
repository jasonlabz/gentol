package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/jasonlabz/gentol/configx"
	"github.com/jasonlabz/gentol/datasource"
	"github.com/jasonlabz/gentol/gormx"
)

func main() {
	tableConfigs := configx.TableConfigs
	//gormAnnotation := tableConfigs.AddGormAnnotation
	//protobufAnnotation := tableConfigs.AddProtobufAnnotation
	//gormAnnotation := tableConfigs.RunGoFmt
	//jsonFormat := tableConfigs.JsonFormat
	//xmlFormat := tableConfigs.XMLFormat
	//protobufFormat := tableConfigs.ProtobufFormat

	for _, dbInfo := range tableConfigs.Configs {
		dbInfo.ModelModule = "//TODO: fix path/"
		dbInfo.DaoModule = "//TODO: fix path/"
		modPath, _ := filepath.Abs(fmt.Sprintf(".%sgo1.mod", string(os.PathSeparator)))
		var relationModelPath = dbInfo.ModelPath
		var relationDaoPath = dbInfo.DaoPath
		if filepath.IsAbs(dbInfo.ModelPath) && tableConfigs.GoModule != "" {
			lastIndex := strings.LastIndex(dbInfo.ModelPath, tableConfigs.GoModule)
			if lastIndex != -1 {
				modPath = dbInfo.ModelPath[:lastIndex+len(tableConfigs.GoModule)] + fmt.Sprintf("%sgo.mod", string(os.PathSeparator))
				relationModelPath = strings.ReplaceAll(dbInfo.ModelPath[lastIndex+len(tableConfigs.GoModule)+1:], string(os.PathSeparator), "/")
			}
		}
		if filepath.IsAbs(dbInfo.DaoPath) && tableConfigs.GoModule != "" {
			lastIndex := strings.LastIndex(dbInfo.DaoPath, tableConfigs.GoModule)
			if lastIndex != -1 {
				modPath = dbInfo.DaoPath[:lastIndex+len(tableConfigs.GoModule)] + fmt.Sprintf("%sgo.mod", string(os.PathSeparator))
				relationDaoPath = strings.ReplaceAll(dbInfo.DaoPath[lastIndex+len(tableConfigs.GoModule)+1:], string(os.PathSeparator), "/")
			}
		}
		if IsExist(modPath) {
			modFile, err := os.Open(modPath)
			if err != nil {
				goto process
			}
			defer modFile.Close()
			scanner := bufio.NewScanner(modFile)
			for scanner.Scan() {
				lineText := scanner.Text()
				if strings.Contains(lineText, "module ") {
					relationModelPath = strings.ReplaceAll(relationModelPath, modPath[:strings.LastIndex(modPath, string(os.PathSeparator))], "")
					relationDaoPath = strings.ReplaceAll(relationDaoPath, modPath[:strings.LastIndex(modPath, string(os.PathSeparator))], "")
					dbInfo.ModelModule = strings.TrimSpace(strings.ReplaceAll(lineText, "module ", "")) +
						"/" + strings.TrimLeft(strings.ReplaceAll(relationModelPath, string(os.PathSeparator), "/"), "/")
					dbInfo.DaoModule = strings.TrimSpace(strings.ReplaceAll(lineText, "module ", "")) +
						"/" + strings.TrimLeft(strings.ReplaceAll(relationDaoPath, string(os.PathSeparator), "/"), "/")
					break
				}
			}
		} else {
			modelAbsPath, err := filepath.Abs(dbInfo.ModelPath)
			daoAbsPath, err := filepath.Abs(dbInfo.DaoPath)
			if err != nil {
				goto process
			}
			var rangePath = modelAbsPath
			for len(rangePath) > 0 {
				modPath = filepath.Join(rangePath, "go.mod")
				if IsExist(modPath) {
					relationModelPath = strings.ReplaceAll(strings.ReplaceAll(modelAbsPath, rangePath, ""), string(os.PathSeparator), "/")
					relationDaoPath = strings.ReplaceAll(strings.ReplaceAll(daoAbsPath, rangePath, ""), string(os.PathSeparator), "/")
					modFile, err := os.Open(modPath)
					if err != nil {
						goto process
					}
					defer modFile.Close()
					scanner := bufio.NewScanner(modFile)
					for scanner.Scan() {
						lineText := scanner.Text()
						if strings.Contains(lineText, "module ") {
							dbInfo.ModelModule = strings.TrimSpace(strings.ReplaceAll(lineText, "module ", "")) +
								"/" + strings.TrimLeft(strings.ReplaceAll(relationModelPath, string(os.PathSeparator), "/"), "/")
							dbInfo.DaoModule = strings.TrimSpace(strings.ReplaceAll(lineText, "module ", "")) +
								"/" + strings.TrimLeft(strings.ReplaceAll(relationDaoPath, string(os.PathSeparator), "/"), "/")
							break
						}
					}
					goto process
				}
				rangePath = rangePath[:strings.LastIndex(rangePath, string(os.PathSeparator))]
			}
		}
	process:
		dbConfig := &gormx.Config{DBName: dbInfo.DBName}
		dbConfig.DSN = dbInfo.DSN
		dbConfig.DBType = gormx.DBType(dbInfo.DBType)
		db, err := gormx.LoadDBInstance(dbConfig)
		if err != nil {
			panic(err)
		}

		ds, err := datasource.GetDS(gormx.DBType(dbInfo.DBType))
		if err != nil {
			panic(err)
		}
		checkDupTableMap := make(map[string]map[string]bool, 0)
		for _, tableInfo := range dbInfo.Tables {
			schemaName := strings.Trim(tableInfo.SchemaName, "\"")

			if len(tableInfo.TableList) == 0 {
				dbTableMap, err := ds.GetTablesUnderDB(context.TODO(), dbConfig.DBName)
				if err != nil {
					panic(err)
				}
				for schemaItem, dbMeta := range dbTableMap {
					if schemaName == "" {
						continue
					}

					if schemaItem != schemaName {
						continue
					}

					for _, tableItem := range dbMeta.TableInfoList {
						if tableMap, ok := checkDupTableMap[schemaName]; !ok {
							checkDupTableMap[schemaName] = map[string]bool{
								tableItem.TableName: true,
							}
						} else {
							tableMap[tableItem.TableName] = true
						}
					}
				}
			} else {
				for _, tableName := range tableInfo.TableList {
					tableNameNext := strings.Trim(tableName, "\"")
					if tableMap, ok := checkDupTableMap[schemaName]; !ok {
						checkDupTableMap[schemaName] = map[string]bool{
							tableNameNext: true,
						}
					} else {
						tableMap[tableNameNext] = true
					}
				}
			}

			if len(checkDupTableMap) == 0 {
				continue
			}
			for schemaHandle, tableMap := range checkDupTableMap {
				for tableName := range tableMap {
					joinTableName := func() string {
						if schemaHandle == "" {
							return fmt.Sprintf("%s", tableName)
						}
						return fmt.Sprintf("%s.%s", schemaHandle, tableName)
					}()

					columnTypes, getColumnErr := db.Migrator().ColumnTypes(joinTableName)
					if getColumnErr != nil {
						log.Printf(getColumnErr.Error())
						continue
					}
					WriteModel(dbInfo, schemaHandle, tableName, columnTypes)

					if !dbInfo.OnlyModel {
						WriteDao(dbInfo, schemaHandle, tableName, columnTypes)
					}
				}
			}

		}
	}
}
