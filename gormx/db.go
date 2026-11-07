package gormx

import (
	"errors"
	"fmt"
	"sync"

	oracle "github.com/godoes/gorm-oracle"
	dm "github.com/jasonlabz/gorm-dm-driver"
	"github.com/jasonlabz/sqlite"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
)

var dbMap *sync.Map

func init() {
	dbMap = &sync.Map{}
}

type DBWrapper struct {
	DB     *gorm.DB
	Config *Config
}

// newDialect 按 DBType 返回对应的 dialector。
// LoadDBInstance 与 InitConfig 共用同一份映射，避免两处 switch 各自漂移
// （历史上 InitConfig 就漏了 oracle / sqlite / dm，导致对应 operator 全部不可用）。
func newDialect(config *Config) (gorm.Dialector, error) {
	switch config.DBType {
	case DBTypeMySQL:
		return mysql.Open(config.DSN), nil
	case DBTypeGreenplum, DBTypePostgres:
		return postgres.Open(config.DSN), nil
	case DBTypeOracle:
		return oracle.Open(config.DSN), nil
	case DBTypeSQLite:
		return sqlite.Open(config.DSN), nil
	case DBTypeSqlserver:
		return sqlserver.Open(config.DSN), nil
	case DBTypeDM:
		return dm.Open(config.DSN), nil
	default:
		return nil, fmt.Errorf("unsupported dbType: %s", config.DBType)
	}
}

func LoadDBInstance(config *Config) (*gorm.DB, error) {
	if config == nil {
		return nil, errors.New("no db config")
	}
	if config.DBName == "" {
		return nil, errors.New("no db name")
	}
	if config.DSN == "" &&
		config.GenDSN() == "" {
		return nil, errors.New("no db dsn")
	}
	_dbWrapper, ok := dbMap.Load(config.DBName)
	if ok {
		return _dbWrapper.(*DBWrapper).DB, nil
	}
	dialect, err := newDialect(config)
	if err != nil {
		return nil, err
	}
	if config.MaxOpenConn == 0 {
		config.MaxOpenConn = defaultMaxOpenConn
	}
	if config.MaxIdleConn == 0 {
		config.MaxIdleConn = defaultMaxIdleConn
	}
	if config.ConnMaxLifeTime == 0 {
		config.ConnMaxLifeTime = defaultConnMaxLifeTime
	}
	db, err := gorm.Open(dialect)
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxOpenConns(config.MaxOpenConn)
	sqlDB.SetMaxIdleConns(config.MaxIdleConn)
	sqlDB.SetConnMaxLifetime(config.ConnMaxLifeTime)
	dbWrapper := &DBWrapper{
		DB:     db,
		Config: config,
	}
	dbMap.Store(config.DBName, dbWrapper)
	return db, nil
}

func InitConfig(config *Config) error {
	if config == nil {
		return errors.New("no db config")
	}
	if config.DBName == "" {
		return errors.New("no db name")
	}
	if config.DSN == "" &&
		config.GenDSN() == "" {
		return errors.New("no db dsn")
	}
	_, ok := dbMap.Load(config.DBName)
	if ok {
		return nil
	}
	dialect, err := newDialect(config)
	if err != nil {
		return err
	}
	db, err := gorm.Open(dialect)
	if err != nil {
		return err
	}
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	if config.MaxOpenConn == 0 {
		config.MaxOpenConn = defaultMaxOpenConn
	}
	if config.MaxIdleConn == 0 {
		config.MaxIdleConn = defaultMaxIdleConn
	}
	if config.ConnMaxLifeTime == 0 {
		config.ConnMaxLifeTime = defaultConnMaxLifeTime
	}
	sqlDB.SetMaxOpenConns(config.MaxOpenConn)
	sqlDB.SetMaxIdleConns(config.MaxIdleConn)
	sqlDB.SetConnMaxLifetime(config.ConnMaxLifeTime)
	dbWrapper := &DBWrapper{
		DB:     db,
		Config: config,
	}
	dbMap.Store(config.DBName, dbWrapper)
	return nil
}

func GetDBConfig(name string) (*Config, error) {
	db, ok := dbMap.Load(name)
	if !ok {
		return nil, errors.New("no db instance")
	}

	return db.(*DBWrapper).Config, nil
}

func GetDB(name string) (*gorm.DB, error) {
	db, ok := dbMap.Load(name)
	if !ok {
		return nil, errors.New("no db instance")
	}

	return db.(*DBWrapper).DB, nil
}

func Close(dbName string) error {
	if dbName == "" {
		return errors.New("empty db name")
	}
	v, ok := dbMap.LoadAndDelete(dbName)
	if !ok || v == nil {
		return nil
	}
	db, err := v.(*DBWrapper).DB.DB()
	if err != nil {
		return err
	}
	if db == nil {
		return nil
	}
	return db.Close()
}

func Ping(dbName string) error {
	db, err := GetDB(dbName)
	if err != nil {
		return err
	}
	if db == nil {
		return errors.New("db instance is nil")
	}
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	if sqlDB == nil {
		return errors.New("sql db is nil")
	}
	return sqlDB.Ping()
}
