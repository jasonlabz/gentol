package main

import (
	"context"
	"github.com/onlyzzg/gentol/src/config"
	"github.com/onlyzzg/gentol/src/datasource"
	"github.com/onlyzzg/gentol/src/gen"
	"github.com/onlyzzg/gentol/src/gormx"
	"log"
)

// genModels is gorm/gen generated models
func genModels(g *gen.Generator, tableMap map[string][]string) (models map[string][]interface{}, err error) {
	// Execute some data table tasks
	models = make(map[string][]interface{}, 0)
	for schemaName, tableList := range tableMap {
		for _, tableName := range tableList {
			model := g.GenerateModel(schemaName, tableName)
			modelList, ok := models[schemaName]
			if !ok {
				modelList = []interface{}{model}
			} else {
				modelList = append(modelList, model)
			}
			models[schemaName] = modelList
		}
	}
	return models, nil
}

func main() {
	tableConfigs := config.TableConfigs
	for _, database := range tableConfigs.Configs {
		dbConfig := &gormx.Config{DBName: database.DBName}
		dbConfig.DSN = database.DSN
		dbConfig.DBType = gormx.DBType(database.DBType)
		db, err := gormx.GetDBByConfig(dbConfig)
		if err != nil {
			panic(err)
		}
		ds, err := datasource.GetDS(gormx.DBType(database.DBType))
		if err != nil {
			panic(err)
		}
		tableMap := make(map[string][]string, 0)
		genConfig := gen.Config{
			OutPath:      database.DaoPath,
			ModelPkgPath: database.ModelPath,
		}
		for _, tableInfo := range database.Tables {
			tableList, ok := tableMap[tableInfo.SchemaName]
			if !ok {
				tableList = []string{tableInfo.TableName}
			} else {
				tableList = append(tableList, tableInfo.TableName)
			}
			tableMap[tableInfo.SchemaName] = tableList
		}
		if len(tableMap) == 0 {
			dbTableMap, err := ds.GetTablesUnderDB(context.Background(), dbConfig.DBName)
			if err != nil {
				panic(err)
			}
			for schema, tables := range dbTableMap {
				for _, tableInfo := range tables.TableInfoList {
					tableList, ok := tableMap[schema]
					if !ok {
						tableList = []string{tableInfo.TableName}
					} else {
						tableList = append(tableList, tableInfo.TableName)
					}
					tableMap[schema] = tableList
				}
			}
		}

		g := gen.NewGenerator(genConfig)

		g.UseDB(db)

		modelMap, err := genModels(g, tableMap)
		if err != nil {
			log.Fatalln("get tables info fail:", err)
		}
		models := make([]interface{}, 0)
		for _, modelList := range modelMap {
			models = append(models, modelList...)
		}
		g.ApplyBasic(models...)

		g.Execute()
	}
}

//func main() {
//	tableConfigs := config.TableConfigs
//	for _, database := range tableConfigs.Configs {
//		dbConfig := &gormx.Config{DBName: database.DBName}
//		dbConfig.DSN = database.DSN
//		dbConfig.DBType = gormx.DBType(database.DBType)
//		err := gormx.InitConfig(dbConfig)
//		if err != nil {
//			panic(err)
//		}
//		ds, err := datasource.GetDS(gormx.DBType(database.DBType))
//		if err != nil {
//			panic(err)
//		}
//		for _, tableInfo := range database.Tables {
//			tableColMap, getErr := ds.GetColumnsUnderTable(context.Background(), dbConfig.DBName, tableInfo.SchemaName, []string{tableInfo.TableName})
//			if getErr != nil {
//				panic(getErr)
//			}
//			for _, info := range tableColMap {
//				tableName := info.TableName
//				columnInfoList := info.ColumnInfoList
//				fmt.Println("tableName: " + tableName)
//				fmt.Print("columnInfoList: ")
//				columns := make([]string, 0)
//				for _, columnInfo := range columnInfoList {
//					columns = append(columns, columnInfo.ColumnName)
//				}
//				fmt.Println(strings.Join(columns, ","))
//			}
//		}
//	}
//}
