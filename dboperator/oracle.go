package dboperator

import (
	"context"
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
	//TODO implement me
	panic("implement me")
}

func (o OracleOperator) GetColumns(ctx context.Context, dbName string) (dbTableColMap map[string]map[string]*TableColInfo, err error) {
	//TODO implement me
	panic("implement me")
}

func (o OracleOperator) GetColumnsUnderTables(ctx context.Context, dbName, logicDBName string, tableNames []string) (tableColMap map[string]*TableColInfo, err error) {
	//TODO implement me
	panic("implement me")
}

func (o OracleOperator) CreateSchema(ctx context.Context, dbName, schemaName, commentInfo string) (err error) {
	//TODO implement me
	panic("implement me")
}

func (o OracleOperator) ExecuteDDL(ctx context.Context, dbName, ddlStatement string) (err error) {
	//TODO implement me
	panic("implement me")
}
