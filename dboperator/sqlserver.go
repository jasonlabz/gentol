package dboperator

import (
	"context"
	"github.com/onlyzzg/gentol/gormx"
)

func NewSqlserverOperator() IOperator {
	return &SqlServerOperator{}
}

type SqlServerOperator struct{}

func (s SqlServerOperator) Open(config *gormx.Config) error {
	return gormx.InitWithConfig(config)
}

func (s SqlServerOperator) Ping(dbName string) error {
	return gormx.Ping(dbName)
}

func (s SqlServerOperator) Close(dbName string) error {
	return gormx.Close(dbName)
}

func (s SqlServerOperator) GetTablesUnderDB(ctx context.Context, dbName string) (dbTableMap map[string]*LogicDBInfo, err error) {
	//TODO implement me
	panic("implement me")
}

func (s SqlServerOperator) GetColumns(ctx context.Context, dbName string) (dbTableColMap map[string]map[string]*TableColInfo, err error) {
	//TODO implement me
	panic("implement me")
}

func (s SqlServerOperator) GetColumnsUnderTables(ctx context.Context, dbName, logicDBName string, tableNames []string) (tableColMap map[string]*TableColInfo, err error) {
	//TODO implement me
	panic("implement me")
}

func (s SqlServerOperator) CreateSchema(ctx context.Context, dbName, schemaName, commentInfo string) (err error) {
	//TODO implement me
	panic("implement me")
}

func (s SqlServerOperator) ExecuteDDL(ctx context.Context, dbName, ddlStatement string) (err error) {
	//TODO implement me
	panic("implement me")
}
