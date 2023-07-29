package dboperator

import (
	"context"
	"errors"
	"fmt"
	"github.com/onlyzzg/gentol/src/gormx"
)

func NewMySQLOperator() IOperator {
	return &MySQLOperator{}
}

type MySQLOperator struct{}

func (m MySQLOperator) Open(config *gormx.Config) error {
	return gormx.InitConfig(config)
}

func (m MySQLOperator) Ping(dbName string) error {
	return gormx.Ping(dbName)
}

func (m MySQLOperator) Close(dbName string) error {
	return gormx.Close(dbName)
}

func (m MySQLOperator) GetDataBySQL(ctx context.Context, dbName, sqlStatement string) (rows []map[string]interface{}, err error) {
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

func (m MySQLOperator) GetTableData(ctx context.Context, dbName, schemaName, tableName string, pageInfo *Pagination) (rows []map[string]interface{}, err error) {
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
	err = db.WithContext(ctx).
		Table(queryTable).
		Count(&count).
		Offset(int(pageInfo.GetOffset())).
		Limit(int(pageInfo.PageSize)).
		Find(&rows).Error
	pageInfo.Total = count
	pageInfo.SetPageCount()
	return
}

func (m MySQLOperator) GetTablesUnderDB(ctx context.Context, dbName string) (dbTableMap map[string]*LogicDBInfo, err error) {
	dbTableMap = make(map[string]*LogicDBInfo)
	if dbName == "" {
		err = errors.New("empty dnName")
		return
	}
	gormDBTables := make([]*GormDBTable, 0)
	db, err := gormx.GetDB(dbName)
	if err != nil {
		return
	}
	err = db.WithContext(ctx).
		Raw("SELECT TABLE_SCHEMA as table_schema, " +
			"TABLE_NAME as table_name, " +
			"TABLE_COMMENT as comments " +
			"FROM INFORMATION_SCHEMA.TABLES " +
			"WHERE TABLE_TYPE = 'BASE TABLE' " +
			"AND TABLE_SCHEMA NOT IN ('mysql', 'sys', 'performance_schema', 'information_schema') " +
			"ORDER  BY TABLE_SCHEMA, TABLE_NAME").
		Find(&gormDBTables).Error
	if err != nil {
		return
	}
	if len(gormDBTables) == 0 {
		return
	}
	for _, row := range gormDBTables {
		if logicDBInfo, ok := dbTableMap[row.TableSchema]; !ok {
			dbTableMap[row.TableSchema] = &LogicDBInfo{
				SchemaName: row.TableSchema,
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

func (m MySQLOperator) GetColumns(ctx context.Context, dbName string) (dbTableColMap map[string]map[string]*TableColInfo, err error) {
	dbTableColMap = make(map[string]map[string]*TableColInfo, 0)
	if dbName == "" {
		err = errors.New("empty dnName")
		return
	}
	gormTableColumns := make([]*GormTableColumn, 0)
	db, err := gormx.GetDB(dbName)
	if err != nil {
		return
	}
	err = db.WithContext(ctx).
		Raw("select " +
			"t.TABLE_SCHEMA table_schema, " +
			"t.TABLE_NAME table_name, " +
			"c.COLUMN_NAME column_name, " +
			"c.COLUMN_COMMENT comments, " +
			"c.COLUMN_TYPE data_type " +
			"from " +
			"INFORMATION_SCHEMA.TABLES t " +
			"inner join INFORMATION_SCHEMA.COLUMNS c on " +
			"t.TABLE_NAME = c.TABLE_NAME " +
			"and t.TABLE_SCHEMA = c.TABLE_SCHEMA " +
			"where " +
			"t.TABLE_TYPE = 'BASE TABLE' " +
			"AND t.TABLE_SCHEMA NOT IN ('mysql', 'sys', 'performance_schema', 'information_schema') " +
			"ORDER BY t.TABLE_NAME, c.COLUMN_NAME").
		Find(&gormTableColumns).Error
	if err != nil {
		return
	}
	if len(gormTableColumns) == 0 {
		return
	}

	for _, row := range gormTableColumns {
		if dbTableColInfoMap, ok := dbTableColMap[row.TableSchema]; !ok {
			dbTableColMap[row.TableSchema] = map[string]*TableColInfo{
				row.TableName: {
					TableName: row.TableName,
					ColumnInfoList: []*ColumnInfo{{
						ColumnName: row.ColumnName,
						Comment:    row.Comments,
						DataType:   row.DataType,
					}},
				},
			}
		} else if tableColInfo, ok_ := dbTableColInfoMap[row.TableName]; !ok_ {
			dbTableColInfoMap[row.TableName] = &TableColInfo{
				TableName: row.TableName,
				ColumnInfoList: []*ColumnInfo{{
					ColumnName: row.ColumnName,
					Comment:    row.Comments,
					DataType:   row.DataType,
				}},
			}
		} else {
			tableColInfo.ColumnInfoList = append(tableColInfo.ColumnInfoList, &ColumnInfo{
				ColumnName: row.ColumnName,
				Comment:    row.Comments,
				DataType:   row.DataType,
			})
		}
	}
	return
}

func (m MySQLOperator) GetColumnsUnderTables(ctx context.Context, dbName, logicDBName string, tableNames []string) (tableColMap map[string]*TableColInfo, err error) {
	tableColMap = make(map[string]*TableColInfo, 0)
	if dbName == "" {
		err = errors.New("empty dnName")
		return
	}
	if len(tableNames) == 0 {
		err = errors.New("empty tableNames")
		return
	}

	gormTableColumns := make([]*GormTableColumn, 0)
	db, err := gormx.GetDB(dbName)
	if err != nil {
		return
	}
	db.WithContext(ctx).
		Raw("select "+
			"t.TABLE_SCHEMA table_schema, "+
			"t.TABLE_NAME table_name, "+
			"c.COLUMN_NAME column_name, "+
			"c.COLUMN_COMMENT comments, "+
			"c.COLUMN_TYPE data_type "+
			"from "+
			"INFORMATION_SCHEMA.TABLES t "+
			"inner join INFORMATION_SCHEMA.COLUMNS c on "+
			"t.TABLE_NAME = c.TABLE_NAME "+
			"and t.TABLE_SCHEMA = c.TABLE_SCHEMA "+
			"where "+
			"t.TABLE_TYPE = 'BASE TABLE' "+
			"AND t.TABLE_SCHEMA = ? "+
			"AND t.TABLE_NAME IN ? "+
			"ORDER BY t.TABLE_NAME, c.COLUMN_NAME", logicDBName, tableNames).
		Find(&gormTableColumns)
	if len(gormTableColumns) == 0 {
		return
	}

	for _, row := range gormTableColumns {
		if tableColInfo, ok := tableColMap[row.TableName]; !ok {
			tableColMap[row.TableName] = &TableColInfo{
				TableName: row.TableName,
				ColumnInfoList: []*ColumnInfo{{
					ColumnName: row.ColumnName,
					Comment:    row.Comments,
					DataType:   row.DataType,
				}},
			}
		} else {
			tableColInfo.ColumnInfoList = append(tableColInfo.ColumnInfoList, &ColumnInfo{
				ColumnName: row.ColumnName,
				Comment:    row.Comments,
				DataType:   row.DataType,
			})
		}
	}
	return
}

func (m MySQLOperator) CreateSchema(ctx context.Context, dbName, schemaName, commentInfo string) (err error) {
	if dbName == "" {
		err = errors.New("empty dnName")
		return
	}
	if commentInfo == "" {
		commentInfo = schemaName
	}
	db, err := gormx.GetDB(dbName)
	if err != nil {
		return
	}
	err = db.WithContext(ctx).Exec("create schema if not exists " + schemaName).Error
	if err != nil {
		return
	}
	return
}

func (m MySQLOperator) ExecuteDDL(ctx context.Context, dbName, ddlStatement string) (err error) {
	if dbName == "" {
		err = errors.New("empty dnName")
		return
	}
	db, err := gormx.GetDB(dbName)
	if err != nil {
		return
	}
	err = db.WithContext(ctx).Exec(ddlStatement).Error
	if err != nil {
		return
	}
	return
}
