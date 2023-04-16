package dboperator

import (
	"context"
	"errors"
	"fmt"
	"github.com/onlyzzg/gentol/gormx"
)

func NewPGOperator() IOperator {
	return &PGOperator{}
}

type PGOperator struct{}

func (P PGOperator) Open(config *gormx.Config) error {
	return gormx.InitWithConfig(config)
}

func (P PGOperator) Ping(dbName string) error {
	return gormx.Ping(dbName)
}

func (P PGOperator) Close(dbName string) error {
	return gormx.Close(dbName)
}

func (P PGOperator) GetTablesUnderDB(ctx context.Context, dbName string) (dbTableMap map[string]*LogicDBInfo, err error) {
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
		Raw("SELECT tb.schemaname as table_schema, " +
			"tb.tablename as table_name, " +
			"d.description as comment " +
			"FROM pg_tables tb " +
			"JOIN pg_class c ON c.relname = tb.tablename " +
			"LEFT JOIN pg_description d ON d.objoid = c.oid AND d.objsubid = '0' " +
			"WHERE schemaname <> 'information_schema' " +
			"AND tablename NOT LIKE 'pg%' " +
			"AND tablename NOT LIKE 'gp%' " +
			"AND tablename NOT LIKE 'sql_%' ").
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

func (P PGOperator) GetColumns(ctx context.Context, dbName string) (dbTableColMap map[string]map[string]*TableColInfo, err error) {
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
		Raw("select " +
			"t.table_schema, " +
			"t.table_name, " +
			"c.column_name, " +
			"c.udt_name data_type " +
			"from " +
			"information_schema.tables t " +
			"inner join information_schema.columns c on " +
			"t.table_name = c.table_name " +
			"and t.table_schema = c.table_schema " +
			"where " +
			"t.table_schema <> 'information_schema' " +
			"AND t.table_name NOT LIKE 'pg%' " +
			"AND t.table_name NOT LIKE 'gp%' " +
			"AND t.table_name NOT LIKE 'sql_%'").
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
						DataType:   row.DataType,
					}},
				},
			}
		} else if tableColInfo, ok_ := dbTableColInfoMap[row.TableName]; !ok_ {
			dbTableColInfoMap[row.TableName] = &TableColInfo{
				TableName: row.TableName,
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
	return
}

func (P PGOperator) GetColumnsUnderTables(ctx context.Context, dbName, logicDBName string, tableNames []string) (tableColMap map[string]*TableColInfo, err error) {
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
			"t.table_name, "+
			"c.column_name, "+
			"c.udt_name data_type "+
			"from "+
			"information_schema.tables t "+
			"inner join information_schema.columns c on "+
			"t.table_name = c.table_name "+
			"where "+
			"t.table_schema = ? "+
			"and c.table_name in ?", logicDBName, tableNames).
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
	return
}

func (P PGOperator) CreateSchema(ctx context.Context, dbName, schemaName, commentInfo string) (err error) {
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
	commentStr := fmt.Sprintf("comment on schema %s is '%s'", schemaName, commentInfo)
	err = db.WithContext(ctx).Exec(commentStr).Error
	if err != nil {
		return
	}
	return
}

func (P PGOperator) ExecuteDDL(ctx context.Context, dbName, ddlStatement string) (err error) {
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
