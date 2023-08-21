package main

import (
	"context"
	"fmt"
	"github.com/jasonlabz/gentol/configx"
	"github.com/jasonlabz/gentol/datasource"
	"github.com/jasonlabz/gentol/gormx"
	"github.com/jasonlabz/gentol/metadata"
	"gorm.io/gorm"
	"log"
	"os"
	"path/filepath"
	"strings"
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
					fmt.Printf("%+v", columnTypes)
				}
			}

		}
	}
}

func WriteModel(dbInfo *configx.Database, schemaName, tableName string, columnTypes []gorm.ColumnType) {
	modelData := &metadata.ModelMeta{
		ModelPackageName: func() string {
			if dbInfo.ModelPath == "" {
				dbInfo.ModelPath = "model"
			}
			return metadata.ToLower(filepath.Base(dbInfo.ModelPath))
		}(),
		ModelStructName: metadata.UnderscoreToUpperCamelCase(tableName),
	}

	for _, columnType := range columnTypes {
		modelData.ColumnList = append(modelData.ColumnList, &metadata.ColumnInfo{
			ColumnName: columnType.Name(),
			ColumnType: func() string {
				columnTypeName, ok := columnType.ColumnType()
				if ok {
					return columnTypeName
				}
				return ""
			}(),
			DataBaseType: columnType.DatabaseTypeName(),
			IsPrimaryKey: func() bool {
				if prime, ok := columnType.PrimaryKey(); ok {
					return prime
				}
				return false
			}(),
			AutoIncrement: func() bool {
				if increment, ok := columnType.AutoIncrement(); ok {
					return increment
				}
				return false
			}(),
			Length: func() int64 {
				if length, ok := columnType.Length(); ok {
					return length
				}
				return 0
			}(),
			Nullable: func() bool {
				null, ok := columnType.Nullable()
				if ok {
					return null
				}
				return false
			}(),
			Comment: func() string {
				comment, ok := columnType.Comment()
				if ok {
					return comment
				}
				return ""
			}(),
			DefaultValue: func() string {
				defaultVal, ok := columnType.DefaultValue()
				if ok {
					return defaultVal
				}
				return ""
			}(),
		})
	}
	modelData.DBType = dbInfo.DBType
	modelData.SchemaName = schemaName
	modelData.TableName = tableName
	modelData.ModelPath = dbInfo.ModelPath
	tpl, ok := metadata.LoadTpl("model")
	if !ok {
		panic("undefined template" + "model")
	}
	exist := IsExist(modelData.ModelPath)
	if !exist {
		_ = os.MkdirAll(modelData.ModelPath, 0666)
	}
	ff, _ := filepath.Abs(fmt.Sprintf("%s/%s.go", modelData.ModelPath, metadata.CamelCaseToUnderscore(modelData.TableName)))
	err := RenderingTemplate(tpl, modelData, ff, true)
	if err != nil {
		panic(err)
	}
	return
}
