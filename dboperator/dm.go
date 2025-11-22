package dboperator

import (
	"context"
	"errors"
	"fmt"

	"github.com/jasonlabz/gentol/gormx"
)

func NewDMOperator() IOperator {
	return &DMOperator{}
}

type DMOperator struct{}

func (o DMOperator) Open(config *gormx.Config) error {
	return gormx.InitConfig(config)
}

func (o DMOperator) Ping(dbName string) error {
	return gormx.Ping(dbName)
}

func (o DMOperator) Close(dbName string) error {
	return gormx.Close(dbName)
}

func (o DMOperator) GetDataBySQL(ctx context.Context, dbName, sqlStatement string) (rows []map[string]interface{}, err error) {
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

func (o DMOperator) GetTableData(ctx context.Context, dbName, schemaName, tableName string, pageInfo *Pagination) (rows []map[string]interface{}, err error) {
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

func (o DMOperator) GetTablesUnderDB(ctx context.Context, dbName string) (dbTableMap map[string]*LogicDBInfo, err error) {
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
		Raw("SELECT OWNER as \"table_schema\", " +
			"TABLE_NAME as \"table_name\", " +
			"COMMENTS as \"comments\" " +
			"FROM all_tab_comments " +
			"WHERE OWNER IN " +
			"(select SYS_CONTEXT('USERENV','CURRENT_SCHEMA') CURRENT_SCHEMA from dual) " +
			"ORDER BY OWNER, TABLE_NAME").
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

func (o DMOperator) GetColumns(ctx context.Context, dbName string) (dbTableColMap map[string]map[string]*TableColInfo, err error) {
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
		Raw("SELECT atc.OWNER as \"table_schema\", " +
			"atc.TABLE_NAME as \"table_name\", " +
			"atc.Column_Name as \"column_name\"," +
			" acc.COMMENTS as \"comments\"," +
			"atc.Data_TYPE  as \"data_type\" " +
			"FROM ALL_TAB_COLUMNS atc " +
			"left join all_col_comments acc " +
			"on acc.TABLE_NAME = atc.TABLE_NAME and acc.COLUMN_NAME = atc.COLUMN_NAME " +
			"WHERE atc.OWNER IN (select SYS_CONTEXT('USERENV','CURRENT_SCHEMA') CURRENT_SCHEMA from dual) " +
			"ORDER BY atc.OWNER, atc.TABLE_NAME, atc.Column_Name").
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

func (o DMOperator) GetColumnsUnderTables(ctx context.Context, dbName, logicDBName string, tableNames []string) (tableColMap map[string]*TableColInfo, err error) {
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
	err = db.WithContext(ctx).
		Raw("SELECT atc.OWNER as \"table_schema\", "+
			"atc.TABLE_NAME as \"table_name\", "+
			"atc.Column_Name as \"column_name\","+
			" acc.COMMENTS as \"comments\","+
			"atc.Data_TYPE  as \"data_type\" "+
			"FROM ALL_TAB_COLUMNS atc "+
			"left join all_col_comments acc "+
			"on acc.TABLE_NAME = atc.TABLE_NAME and acc.COLUMN_NAME = atc.COLUMN_NAME "+
			"WHERE atc.OWNER = ? "+
			"AND atc.TABLE_NAME IN ? "+
			"ORDER BY atc.OWNER, atc.TABLE_NAME, atc.Column_Name", logicDBName, tableNames).
		Find(&gormTableColumns).Error
	if err != nil {
		return
	}
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

func (o DMOperator) CreateSchema(ctx context.Context, dbName, schemaName, commentInfo string) (err error) {
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

func (o DMOperator) ExecuteDDL(ctx context.Context, dbName, ddlStatement string) (err error) {
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
