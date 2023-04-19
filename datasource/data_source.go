package datasource

import (
	"context"
	"errors"
	"github.com/onlyzzg/gentol/dboperator"
	"github.com/onlyzzg/gentol/gormx"
)

var dsMap = make(map[gormx.DBType]*DS)

type DS struct {
	Operator dboperator.IOperator
}

// Open open database by config
func (ds *DS) Open(config *gormx.Config) error {
	return ds.Operator.Open(config)
}

// Ping verifies a connection to the database is still alive, establishing a connection if necessary
func (ds *DS) Ping(dbName string) error {
	return ds.Operator.Ping(dbName)
}

// Close database by name
func (ds *DS) Close(dbName string) error {
	return ds.Operator.Close(dbName)
}

// GetTablesUnderDB 获取该库下所有逻辑库及表名
func (ds *DS) GetTablesUnderDB(ctx context.Context, dbName string) (dbTableMap map[string]*dboperator.LogicDBInfo, err error) {
	return ds.Operator.GetTablesUnderDB(ctx, dbName)
}

// GetColumns 获取指定库所有逻辑库及表下字段列表
func (ds *DS) GetColumns(ctx context.Context, dbName string) (dbTableColMap map[string]map[string]*dboperator.TableColInfo, err error) {
	return ds.Operator.GetColumns(ctx, dbName)
}

// GetColumnsUnderTable 获取指定库表下字段列表
func (ds *DS) GetColumnsUnderTable(ctx context.Context, dbName, logicDBName string, tableNames []string) (tableColMap map[string]*dboperator.TableColInfo, err error) {
	return ds.Operator.GetColumnsUnderTables(ctx, dbName, logicDBName, tableNames)
}

// CreateSchema 创建逻辑库
func (ds *DS) CreateSchema(ctx context.Context, dbName, schemaName, commentInfo string) (err error) {
	return ds.Operator.CreateSchema(ctx, dbName, schemaName, commentInfo)
}

// ExecuteDDL 执行DDL
func (ds *DS) ExecuteDDL(ctx context.Context, dbName, logicDBName, tableName, ddlStatement string) (err error) {
	return ds.Operator.ExecuteDDL(ctx, dbName, ddlStatement)
}

func GetDS(dataSourceType gormx.DBType) (ds *DS, err error) {
	var ok bool
	ds, ok = dsMap[dataSourceType]
	if !ok {
		err = errors.New("unsupported db_type")
		return
	}
	return
}

func init() {
	// oracle
	dsMap[gormx.DBTypeOracle] = &DS{
		Operator: dboperator.NewOracleOperator(),
	}
	// postgresql
	dsMap[gormx.DBTypePostgres] = &DS{
		Operator: dboperator.NewPGOperator(),
	}
	// mysql
	dsMap[gormx.DBTypeMySQL] = &DS{
		Operator: dboperator.NewMySQLOperator(),
	}

	// greenplum
	dsMap[gormx.DBTypeGreenplum] = &DS{
		Operator: dboperator.NewGPOperator(),
	}

	// sqlserver
	dsMap[gormx.DBTypeSqlserver] = &DS{
		Operator: dboperator.NewSqlserverOperator(),
	}
}
