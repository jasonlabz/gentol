package dboperator

import (
	"context"
	"errors"
	"fmt"
	"github.com/onlyzzg/gentol/gormx"
)

func NewOracleOperator() IOperator {
	return &OracleOperator{}
}

type OracleOperator struct{}

func (o OracleOperator) Open(config *gormx.Config) error {
	return gormx.InitWithConfig(config)
}

func (o OracleOperator) Ping(dbName string) error {
	return gormx.Ping(dbName)
}

func (o OracleOperator) Close(dbName string) error {
	return gormx.Close(dbName)
}

func (o OracleOperator) GetTablesUnderDB(ctx context.Context, dbName string) (dbTableMap map[string]*LogicDBInfo, err error) {
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
	db.WithContext(ctx).
		Raw("SELECT OWNER as table_schema, " +
			"TABLE_NAME as table_name, " +
			"ALL_TAB_COMMENTS.COMMENTS as comment " +
			"FROM all_tab_comments " +
			"WHERE OWNER IN " +
			"(select SYS_CONTEXT('USERENV','CURRENT_SCHEMA') CURRENT_SCHEMA from dual) " +
			"ORDER BY OWNER, TABLE_NAME").
		Find(&gormDBTables)
	if len(gormDBTables) == 0 {
		return
	}
	for _, row := range gormDBTables {
		if logicDBInfo, ok := dbTableMap[row.TableSchema]; !ok {
			dbTableMap[row.TableSchema] = &LogicDBInfo{
				SchemaName: row.TableSchema,
				TableInfoList: []*TableInfo{{
					TableName: row.TableName,
					Comment:   row.Comment,
				}},
			}
		} else {
			logicDBInfo.TableInfoList = append(logicDBInfo.TableInfoList,
				&TableInfo{
					TableName: row.TableName,
					Comment:   row.Comment,
				})
		}
	}
	return
}

func (o OracleOperator) GetColumns(ctx context.Context, dbName string) (dbTableColMap map[string]map[string]*TableColInfo, err error) {
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
	db.WithContext(ctx).
		Raw("SELECT OWNER as table_schema, " +
			"TABLE_NAME as table_name, " +
			"Column_Name as column_name, " +
			"COMMENTS as comment " +
			"DATA_TYPE as data_type " +
			"FROM all_col_comments " +
			"WHERE OWNER IN " +
			"(select SYS_CONTEXT('USERENV','CURRENT_SCHEMA') CURRENT_SCHEMA from dual) " +
			"ORDER BY OWNER, TABLE_NAME").
		Find(&gormTableColumns)
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
						Comment:    row.Comment,
						DataType:   row.DataType,
					}},
				},
			}
		} else if tableColInfo, ok_ := dbTableColInfoMap[row.TableName]; !ok_ {
			dbTableColInfoMap[row.TableName] = &TableColInfo{
				TableName: row.TableName,
				ColumnInfoList: []*ColumnInfo{{
					ColumnName: row.ColumnName,
					Comment:    row.Comment,
					DataType:   row.DataType,
				}},
			}
		} else {
			tableColInfo.ColumnInfoList = append(tableColInfo.ColumnInfoList, &ColumnInfo{
				ColumnName: row.ColumnName,
				Comment:    row.Comment,
				DataType:   row.DataType,
			})
		}
	}
	return
}

func (o OracleOperator) GetColumnsUnderTables(ctx context.Context, dbName, logicDBName string, tableNames []string) (tableColMap map[string]*TableColInfo, err error) {
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
		Raw("SELECT OWNER as table_schema, "+
			"TABLE_NAME as table_name, "+
			"Column_Name as column_name, "+
			"COMMENTS as comment "+
			"DATA_TYPE as data_type "+
			"FROM all_col_comments "+
			"WHERE OWNER = ? "+
			"AND TABLE_NAME IN ? "+
			"ORDER BY OWNER, TABLE_NAME", logicDBName, tableNames).
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
					Comment:    row.Comment,
					DataType:   row.DataType,
				}},
			}
		} else {
			tableColInfo.ColumnInfoList = append(tableColInfo.ColumnInfoList, &ColumnInfo{
				ColumnName: row.ColumnName,
				Comment:    row.Comment,
				DataType:   row.DataType,
			})
		}
	}
	return
}

func (o OracleOperator) CreateSchema(ctx context.Context, dbName, schemaName, commentInfo string) (err error) {
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
	config, err := gormx.GetDBConfig(dbName)
	if err != nil {
		return
	}
	password := config.Password
	err = db.WithContext(ctx).Exec(fmt.Sprintf("create user %s identified by %s", schemaName, password)).Error
	if err != nil {
		return
	}
	err = db.WithContext(ctx).Exec(fmt.Sprintf("grant connect, resource to %s", schemaName)).Error
	if err != nil {
		return
	}
	return
}

func (o OracleOperator) ExecuteDDL(ctx context.Context, dbName, ddlStatement string) (err error) {
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
