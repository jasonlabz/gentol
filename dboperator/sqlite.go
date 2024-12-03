package dboperator

import (
	"context"
	"errors"
	"fmt"
	
	"github.com/jasonlabz/gentol/gormx"
)

func NewSQLiteOperator() IOperator {
	return &SQLiteOperator{}
}

type SQLiteOperator struct{}

func (m SQLiteOperator) Open(config *gormx.Config) error {
	return gormx.InitConfig(config)
}

func (m SQLiteOperator) Ping(dbName string) error {
	return gormx.Ping(dbName)
}

func (m SQLiteOperator) Close(dbName string) error {
	return gormx.Close(dbName)
}

func (m SQLiteOperator) GetDataBySQL(ctx context.Context, dbName, sqlStatement string) (rows []map[string]interface{}, err error) {
	rows = make([]map[string]interface{}, 0)
	db, err := gormx.GetDB(dbName)
	if err != nil {
		return
	}
	err = db.WithContext(ctx).
		Raw(sqlStatement).
		Find(&rows).Error
	return
}

func (m SQLiteOperator) GetTableData(ctx context.Context, dbName, schemaName, tableName string, pageInfo *Pagination) (rows []map[string]interface{}, err error) {
	rows = make([]map[string]interface{}, 0)
	db, err := gormx.GetDB(dbName)
	if err != nil {
		return
	}
	queryTable := fmt.Sprintf("\"%s\"", tableName)
	if schemaName != "" {
		queryTable = fmt.Sprintf("\"%s\".\"%s\"", schemaName, tableName)
	}
	var count int64
	tx := db.WithContext(ctx).
		Table(queryTable)
	if pageInfo != nil {
		tx = tx.Count(&count).
			Offset(int(pageInfo.GetOffset())).
			Limit(int(pageInfo.PageSize))
	}
	err = tx.Scan(&rows).Error
	if pageInfo != nil {
		pageInfo.Total = count
		pageInfo.SetPageCount()
	}
	return
}

func (m SQLiteOperator) GetTablesUnderDB(ctx context.Context, dbName string) (dbTableMap map[string]*LogicDBInfo, err error) {
	dbTableMap = make(map[string]*LogicDBInfo)
	if dbName == "" {
		err = errors.New("empty dnName")
		return
	}
	defaultName := "sqlite_default"
	gormDBTables := make([]*GormDBTable, 0)
	db, err := gormx.GetDB(dbName)
	if err != nil {
		return
	}
	err = db.WithContext(ctx).
		Raw("SELECT name as table_name" +
			"FROM sqlite_master " +
			"WHERE type = 'table'").
		Find(&gormDBTables).Error
	if err != nil {
		return
	}
	if len(gormDBTables) == 0 {
		return
	}
	for _, row := range gormDBTables {
		if logicDBInfo, ok := dbTableMap[defaultName]; !ok {
			dbTableMap[defaultName] = &LogicDBInfo{
				SchemaName: defaultName,
				TableInfoList: []*TableInfo{{
					TableName: row.TableName,
					Comment:   row.Comments,
				}},
			}
		} else {
			logicDBInfo.TableInfoList = append(logicDBInfo.TableInfoList,
				&TableInfo{
					TableName: row.TableName,
					Comment:   row.Comments,
				})
		}
	}
	return
}

func (m SQLiteOperator) GetColumns(ctx context.Context, dbName string) (dbTableColMap map[string]map[string]*TableColInfo, err error) {
	dbTableColMap = make(map[string]map[string]*TableColInfo, 0)
	if dbName == "" {
		err = errors.New("empty dnName")
		return
	}
	sqliteTableColumn := make([]*SQLiteTableColumn, 0)
	tableMap, err := m.GetTablesUnderDB(ctx, dbName)
	if err != nil {
		return
	}
	defaultName := "sqlite_default"

	for _, schemaTableInfo := range tableMap {
		for _, tableInfo := range schemaTableInfo.TableInfoList {
			db, getErr := gormx.GetDB(dbName)
			if getErr != nil {
				err = getErr
				return
			}
			err = db.WithContext(ctx).
				Raw("PRAGMA table_info(?)", tableInfo.TableName).
				Find(&sqliteTableColumn).Error
			if err != nil {
				return
			}
			if len(sqliteTableColumn) == 0 {
				return
			}

			for _, row := range sqliteTableColumn {
				if dbTableColInfoMap, ok := dbTableColMap[defaultName]; !ok {
					dbTableColMap[defaultName] = map[string]*TableColInfo{
						tableInfo.TableName: {
							TableName: tableInfo.TableName,
							ColumnInfoList: []*ColumnInfo{{
								ColumnName: row.ColumnName,
								DataType:   row.DataType,
							}},
						},
					}
				} else if tableColInfo, ok_ := dbTableColInfoMap[tableInfo.TableName]; !ok_ {
					dbTableColInfoMap[tableInfo.TableName] = &TableColInfo{
						TableName: tableInfo.TableName,
						ColumnInfoList: []*ColumnInfo{{
							ColumnName: row.ColumnName,
							DataType:   row.DataType,
						}},
					}
				} else {
					tableColInfo.ColumnInfoList = append(tableColInfo.ColumnInfoList, &ColumnInfo{
						ColumnName: row.ColumnName,
						DataType:   row.DataType,
					})
				}
			}
		}

	}
	return
}

func (m SQLiteOperator) GetColumnsUnderTables(ctx context.Context, dbName, logicDBName string, tableNames []string) (tableColMap map[string]*TableColInfo, err error) {
	tableColMap = make(map[string]*TableColInfo, 0)
	if dbName == "" {
		err = errors.New("empty dnName")
		return
	}
	if len(tableNames) == 0 {
		err = errors.New("empty tableNames")
		return
	}

	sqliteTableColumns := make([]*SQLiteTableColumn, 0)
	db, err := gormx.GetDB(dbName)
	if err != nil {
		return
	}
	for _, table := range tableNames {
		db.WithContext(ctx).
			Raw("PRAGMA table_info(?)", table).
			Find(&sqliteTableColumns)
		if len(sqliteTableColumns) == 0 {
			continue
		}

		for _, row := range sqliteTableColumns {
			if tableColInfo, ok := tableColMap[table]; !ok {
				tableColMap[table] = &TableColInfo{
					TableName: table,
					ColumnInfoList: []*ColumnInfo{{
						ColumnName: row.ColumnName,
						DataType:   row.DataType,
					}},
				}
			} else {
				tableColInfo.ColumnInfoList = append(tableColInfo.ColumnInfoList, &ColumnInfo{
					ColumnName: row.ColumnName,
					DataType:   row.DataType,
				})
			}
		}
	}

	return
}

func (m SQLiteOperator) CreateSchema(ctx context.Context, dbName, schemaName, commentInfo string) (err error) {

	return
}

func (m SQLiteOperator) ExecuteDDL(ctx context.Context, dbName, ddlStatement string) (err error) {

	return
}
