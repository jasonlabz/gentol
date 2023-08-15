package gormx

import (
	"errors"
	"fmt"
	"github.com/onlyzzg/oracle"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
	"sync"
)

var dbMap *sync.Map

func init() {
	dbMap = &sync.Map{}
}

type DBWrapper struct {
	DB     *gorm.DB
	Config *Config
}

func GetDBByConfig(config *Config) (*gorm.DB, error) {
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
	var dialect gorm.Dialector
	switch config.DBType {
	case DBTypeMySQL:
		dialect = mysql.Open(config.DSN)
	case DBTypeGreenplum:
		fallthrough
	case DBTypePostgres:
		dialect = postgres.Open(config.DSN)
	case DBTypeOracle:
		dialect = oracle.Open(config.DSN)
	case DBTypeSqlserver:
		dialect = sqlserver.Open(config.DSN)
	default:
		return nil, errors.New(fmt.Sprintf("unsupported dbType: %s", string(config.DBType)))
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
	var dialect gorm.Dialector
	switch config.DBType {
	case DBTypeMySQL:
		dialect = mysql.Open(config.DSN)
	case DBTypeGreenplum:
		fallthrough
	case DBTypePostgres:
		dialect = postgres.Open(config.DSN)
	case DBTypeOracle:
		dialect = oracle.Open(config.DSN)
	case DBTypeSqlserver:
		dialect = sqlserver.Open(config.DSN)
	default:
		return errors.New(fmt.Sprintf("unsupported dbType: %s", string(config.DBType)))
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
