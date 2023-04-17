package dboperator

import (
	"context"
	"github.com/onlyzzg/gentol/gormx"
	"math"
)

// IConnector 数据库连接器接口
type IConnector interface {
	Open(config *gormx.Config) error
	Ping(dbName string) error
	Close(dbName string) error
}

// IDataExplorer 数据探查
type IDataExplorer interface {
	// GetTablesUnderDB 获取该库下所有逻辑库及表名
	GetTablesUnderDB(ctx context.Context, dbName string) (dbTableMap map[string]*LogicDBInfo, err error)
	// GetColumns 获取指定库所有逻辑库及表下字段列表
	GetColumns(ctx context.Context, dbName string) (dbTableColMap map[string]map[string]*TableColInfo, err error)
	// GetColumnsUnderTables 获取指定库表下字段列表
	GetColumnsUnderTables(ctx context.Context, dbName, logicDBName string, tableNames []string) (tableColMap map[string]*TableColInfo, err error)
	// CreateSchema 创建逻辑库
	CreateSchema(ctx context.Context, dbName, schemaName, commentInfo string) (err error)
	// ExecuteDDL 执行DDL
	ExecuteDDL(ctx context.Context, dbName, ddlStatement string) (err error)
}

type IOperator interface {
	IConnector
	IDataExplorer
}

type GormDBTable struct {
	TableSchema string `db:"table_schema"`
	TableName   string `db:"table_name"`
	Comments    string `db:"comments"`
}

type GormTableColumn struct {
	TableSchema string `db:"table_schema"`
	TableName   string `db:"table_name"`
	ColumnName  string `db:"column_name"`
	Comments    string `db:"comments"`
	DataType    string `db:"data_type"`
}

type LogicDBInfo struct {
	SchemaName    string
	TableInfoList []*TableInfo
}
type TableInfo struct {
	TableName string // 列名
	Comment   string // 注释
}
type TableColInfo struct {
	TableName      string
	ColumnInfoList []*ColumnInfo // 列
}
type ColumnInfo struct {
	ColumnName string // 列名
	Comment    string // 注释
	DataType   string // 数据类型
}

// Pagination 分页结构体（该分页只适合数据量很少的情况）
type Pagination struct {
	Page      int64 `json:"page"`       // 当前页
	PageSize  int64 `json:"page_size"`  // 每页多少条记录
	PageCount int64 `json:"page_count"` // 一共多少页
	Total     int64 `json:"total"`      // 一共多少条记录
}

func (p *Pagination) SetPageCount() {
	p.PageCount = int64(math.Ceil(float64(p.Total) / float64(p.PageSize)))
	return
}

func (p *Pagination) GetOffset() (offset int64) {
	offset = (p.Page - 1) * p.PageSize
	return
}
