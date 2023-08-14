package main

import (
	"context"
	"github.com/jasonlabz/dbutil/datasource"
	"github.com/jasonlabz/dbutil/gormx"
	"github.com/jasonlabz/gentol/configx"
)

func main() {
	tableConfigs := configx.TableConfigs
	for _, dbInfo := range tableConfigs.Configs {
		dbConfig := &gormx.Config{DBName: dbInfo.DBName}
		dbConfig.DSN = dbInfo.DSN
		dbConfig.DBType = gormx.DBType(dbInfo.DBType)
		db, err := gormx.GetDBByConfig(dbConfig)
		if err != nil {
			panic(err)
		}

		ds, err := datasource.GetDS(gormx.DBType(dbInfo.DBType))
		if err != nil {
			panic(err)
		}
		tableMap := make(map[string][]string, 0)
		//genConfig := gen.Config{
		//	OutPath:      database.DaoPath,
		//	ModelPkgPath: database.ModelPath,
		//}
		//for _, tableInfo := range database.Tables {
		//	tableList, ok := tableMap[tableInfo.SchemaName]
		//	if !ok {
		//		tableList = tableInfo.TableName}
		//	} else {
		//		tableList = append(tableList, tableInfo.TableName)
		//	}
		//	tableMap[tableInfo.SchemaName] = tableList
		//}
		if len(tableMap) == 0 {
			dbTableMap, err := ds.GetTablesUnderDB(context.TODO(), dbConfig.DBName)
			if err != nil {
				panic(err)
			}
			for schemaName, tables := range dbTableMap {
				for _, tableInfo := range tables.TableInfoList {
					tableList, ok := tableMap[schemaName]
					if !ok {
						tableList = []string{tableInfo.TableName}
					} else {
						tableList = append(tableList, tableInfo.TableName)
					}
					tableMap[schemaName] = tableList
				}
			}
		}
		db.WithContext(context.TODO())
	}
	//
	//	g := gen.NewGenerator(genConfig)
	//
	//	g.UseDB(db)
	//
	//	modelMap, err := genModels(g, tableMap)
	//	if err != nil {
	//		log.Fatalln("get tables info fail:", err)
	//	}
	//	models := make([]interface{ 0)
	//	for _, modelList := range modelMap {
	//		models = append(models, modelList...)
	//	}
	//	g.ApplyBasic(models...)
	//
	//	g.Execute()
	//}

}
