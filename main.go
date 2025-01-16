package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/jasonlabz/gentol/configx"
	"github.com/jasonlabz/gentol/datasource"
	"github.com/jasonlabz/gentol/gormx"
)

func main() {
	if len(os.Args) < 2 {
		process()
		return
	}
	switch os.Args[1] {
	case "init", "new":
		projectName := os.Args[2]
		if projectName == "" {
			projectName = "demo"
		}
		match, _ := regexp.MatchString("^[a-zA-Z0-9]+$", projectName)
		if !match {
			log.Fatalf("project name is not valid, only in [a-zA-Z0-9]")
		}
		handleNewProject(projectName)
	//case "update":
	//	version := os.Args[2]
	//	if version == "" {
	//		version = "master"
	//	}
	default:
		process()
	}
}

func process() {
	// 参数获取
	argHandler()
	// 分析数据库
	handleDB()
}

func handleDB() {
	tableConfigs := configx.TableConfigs

	for _, dbInfo := range tableConfigs.Configs {
		modPath, _ := filepath.Abs(fmt.Sprintf(".%sgo.mod", string(os.PathSeparator)))
		var relativeModelPath = dbInfo.ModelPath
		var relativeDaoPath = dbInfo.DaoPath

		// 是否已经读取go.mod文件
		var findModule bool
		if filepath.IsAbs(dbInfo.ModelPath) && tableConfigs.GoModule != "" {
			// 截断golang项目名称
			lastIndex := strings.LastIndex(dbInfo.ModelPath, tableConfigs.GoModule)
			if lastIndex != -1 {
				modPath = dbInfo.ModelPath[:lastIndex+len(tableConfigs.GoModule)] + fmt.Sprintf("%sgo.mod", string(os.PathSeparator))
				relativeModelPath = strings.ReplaceAll(dbInfo.ModelPath[lastIndex+len(tableConfigs.GoModule)+1:], string(os.PathSeparator), "/")
			}
		}
		if filepath.IsAbs(dbInfo.DaoPath) && tableConfigs.GoModule != "" {
			// 截断golang项目名称
			lastIndex := strings.LastIndex(dbInfo.DaoPath, tableConfigs.GoModule)
			if lastIndex != -1 {
				modPath = dbInfo.DaoPath[:lastIndex+len(tableConfigs.GoModule)] + fmt.Sprintf("%sgo.mod", string(os.PathSeparator))
				relativeDaoPath = strings.ReplaceAll(dbInfo.DaoPath[lastIndex+len(tableConfigs.GoModule)+1:], string(os.PathSeparator), "/")
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
					if strings.Contains(lineText, "github.com/jasonlabz/gentol") {
						break
					}
					findModule = true
					relativeModelPath = strings.ReplaceAll(relativeModelPath, modPath[:strings.LastIndex(modPath, string(os.PathSeparator))], "")
					relativeDaoPath = strings.ReplaceAll(relativeDaoPath, modPath[:strings.LastIndex(modPath, string(os.PathSeparator))], "")
					dbInfo.ModelModule = strings.TrimSpace(strings.ReplaceAll(lineText, "module ", "")) +
						"/" + strings.TrimLeft(strings.ReplaceAll(relativeModelPath, string(os.PathSeparator), "/"), "/")
					dbInfo.DaoModule = strings.TrimSpace(strings.ReplaceAll(lineText, "module ", "")) +
						"/" + strings.TrimLeft(strings.ReplaceAll(relativeDaoPath, string(os.PathSeparator), "/"), "/")
					break
				}
			}
		}

		if !findModule {
			// 没有找到，则遍历目录
			modelAbsPath := func() string {
				if filepath.IsAbs(dbInfo.ModelPath) {
					return dbInfo.ModelPath
				}
				trimLeft := strings.TrimLeft(dbInfo.ModelPath, "."+string(os.PathSeparator))
				trimLeft = strings.TrimLeft(trimLeft, "./")
				abs, err := filepath.Abs(trimLeft)
				if err != nil {
					abs = dbInfo.ModelPath
				}
				return abs
			}()
			daoAbsPath := func() string {
				if filepath.IsAbs(dbInfo.DaoPath) {
					return dbInfo.DaoPath
				}
				trimLeft := strings.TrimLeft(dbInfo.DaoPath, "."+string(os.PathSeparator))
				trimLeft = strings.TrimLeft(trimLeft, "./")
				abs, err := filepath.Abs(trimLeft)
				if err != nil {
					abs = dbInfo.DaoPath
				}
				return abs
			}()

			var rangePath = modelAbsPath
			for len(rangePath) > 0 {
				modPath = filepath.Join(rangePath, "go.mod")
				if IsExist(modPath) {
					relativeModelPath = strings.ReplaceAll(strings.ReplaceAll(modelAbsPath, rangePath, ""), string(os.PathSeparator), "/")
					relativeDaoPath = strings.ReplaceAll(strings.ReplaceAll(daoAbsPath, rangePath, ""), string(os.PathSeparator), "/")
					modFile, err := os.Open(modPath)
					if err != nil {
						break
					}
					defer modFile.Close()
					scanner := bufio.NewScanner(modFile)
					for scanner.Scan() {
						lineText := scanner.Text()
						if strings.Contains(lineText, "module ") {
							dbInfo.ModelModule = strings.TrimSpace(strings.ReplaceAll(lineText, "module ", "")) +
								"/" + strings.TrimLeft(strings.ReplaceAll(relativeModelPath, string(os.PathSeparator), "/"), "/")
							dbInfo.DaoModule = strings.TrimSpace(strings.ReplaceAll(lineText, "module ", "")) +
								"/" + strings.TrimLeft(strings.ReplaceAll(relativeDaoPath, string(os.PathSeparator), "/"), "/")
							goto process
						}
					}
				}
				lastIndex := strings.LastIndex(rangePath, string(os.PathSeparator))
				if lastIndex == -1 {
					break
				}
				rangePath = rangePath[:lastIndex]
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
		if len(dbInfo.Tables) == 0 {
			dbInfo.Tables = []*configx.TableInfo{
				{SchemaName: "", TableList: []string{}},
			}
		}
		for _, tableInfo := range dbInfo.Tables {
			schemaName := strings.Trim(tableInfo.SchemaName, "\"")

			if len(tableInfo.TableList) == 0 {
				dbTableMap, err := ds.GetTablesUnderDB(context.TODO(), dbConfig.DBName)
				if err != nil {
					panic(err)
				}
				for schemaItem, dbMeta := range dbTableMap {
					if schemaName != "" && schemaItem != schemaName {
						continue
					}

					for _, tableItem := range dbMeta.TableInfoList {
						if tableMap, ok := checkDupTableMap[schemaItem]; !ok {
							checkDupTableMap[schemaItem] = map[string]bool{
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
