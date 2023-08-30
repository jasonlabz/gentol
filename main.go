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

					if !dbInfo.OnlyModel {
						WriteDao(dbInfo, schemaHandle, tableName, columnTypes)
					}
				}
			}

		}
	}
}

func WriteModel(dbInfo *configx.Database, schemaName, tableName string, columnTypes []gorm.ColumnType) {
	modelData := &metadata.ModelMeta{
		ModelPackageName: func() string {
			if dbInfo.ModelPath == "" {
				dbInfo.ModelPath = "dal/db/model"
			}
			return metadata.ToLower(filepath.Base(dbInfo.ModelPath))
		}(),
		ModelStructName: metadata.UnderscoreToUpperCamelCase(tableName),
	}
	columnTempList := make([]*metadata.ColumnInfo, 0)
	getColumnInfo(columnTypes, &columnTempList)
	modelData.ColumnList = columnTempList
	modelData.DBType = dbInfo.DBType
	modelData.SchemaName = schemaName
	modelData.TableName = tableName
	modelData.ModelPath = dbInfo.ModelPath
	modelTpl, ok := metadata.LoadTpl("model")
	if !ok {
		panic("undefined template" + "model")
	}
	exist := IsExist(modelData.ModelPath)
	if !exist {
		_ = os.MkdirAll(modelData.ModelPath, 0666)
	}
	ff, _ := filepath.Abs(filepath.Join(modelData.ModelPath, metadata.CamelCaseToUnderscore(modelData.TableName)+".go"))
	err := RenderingTemplate(modelTpl, modelData, ff, true)
	if err != nil {
		panic(err)
	}

	hookFile := filepath.Join(modelData.ModelPath, metadata.CamelCaseToUnderscore(modelData.TableName)+"_hook.go")
	exist = IsExist(hookFile)
	if !exist {
		ff, _ = filepath.Abs(hookFile)
		modelHookTpl, ok := metadata.LoadTpl("model_hook")
		if !ok {
			panic("undefined template" + "model_hook")
		}
		err = RenderingTemplate(modelHookTpl, modelData, ff, true)
		if err != nil {
			panic(err)
		}
	}
	baseFile := filepath.Join(modelData.ModelPath, "base.go")
	exist = IsExist(baseFile)
	if !exist {
		ff, _ = filepath.Abs(baseFile)
		modelBaseTpl, ok := metadata.LoadTpl("model_base")
		if !ok {
			panic("undefined template" + "model_hook")
		}
		err = RenderingTemplate(modelBaseTpl, modelData, ff, true)
		if err != nil {
			panic(err)
		}
	}
	return
}

func WriteDao(dbInfo *configx.Database, schemaName, tableName string, columnTypes []gorm.ColumnType) {
	daoData := &metadata.DaoMeta{
		ModelPackageName: func() string {
			if dbInfo.ModelPath == "" {
				dbInfo.ModelPath = "dal/db/model"
			}
			return metadata.ToLower(filepath.Base(dbInfo.ModelPath))
		}(),
		DaoPackageName: func() string {
			if dbInfo.ModelPath == "" {
				dbInfo.ModelPath = "dal/db/dao"
			}
			return metadata.ToLower(filepath.Base(dbInfo.DaoPath))
		}(),
		ModelModulePath: "TODO:" + "/" + strings.TrimLeft(dbInfo.ModelPath, "/"),
		DaoModulePath:   "TODO:" + "/" + strings.TrimLeft(dbInfo.DaoPath, "/"),
		ModelStructName: metadata.UnderscoreToUpperCamelCase(tableName),
	}
	columnTempList := make([]*metadata.ColumnInfo, 0)
	getColumnInfo(columnTypes, &columnTempList)
	daoData.ColumnList = columnTempList
	daoData.DBType = dbInfo.DBType
	daoData.SchemaName = schemaName
	daoData.TableName = tableName
	daoData.ModelPath = dbInfo.ModelPath
	daoData.DaoPath = dbInfo.DaoPath
	daoTpl, ok := metadata.LoadTpl("dao")
	if !ok {
		panic("undefined template" + "dao")
	}
	daoInterfacePath := filepath.Join(daoData.DaoPath, "interfaces")
	exist := IsExist(daoInterfacePath)
	if !exist {
		_ = os.MkdirAll(daoInterfacePath, 0666)
	}
	ff, _ := filepath.Abs(filepath.Join(daoInterfacePath, metadata.CamelCaseToUnderscore(daoData.TableName)+"_dao.go"))
	err := RenderingTemplate(daoTpl, daoData, ff, true)
	if err != nil {
		panic(err)
	}

	daoImplFile := filepath.Join(daoData.DaoPath, metadata.CamelCaseToUnderscore(daoData.TableName)+"_dao_impl.go")
	ff, _ = filepath.Abs(daoImplFile)
	daoImplTpl, ok := metadata.LoadTpl("dao_impl")
	if !ok {
		panic("undefined template" + "dao_impl")
	}
	err = RenderingTemplate(daoImplTpl, daoData, ff, true)
	if err != nil {
		panic(err)
	}
	baseFile := filepath.Join(daoData.DaoPath, "db.go")
	exist = IsExist(baseFile)
	ff, _ = filepath.Abs(baseFile)
	daoBaseTpl, ok := metadata.LoadTpl("database")
	if !ok {
		panic("undefined template" + "database")
	}
	err = RenderingTemplate(daoBaseTpl, daoData, ff, true)
	if err != nil {
		panic(err)
	}
	return
}

func getColumnInfo(columnTypes []gorm.ColumnType, columnInfoList *[]*metadata.ColumnInfo) {
	for _, columnType := range columnTypes {
		*columnInfoList = append(*columnInfoList, &metadata.ColumnInfo{
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
}
