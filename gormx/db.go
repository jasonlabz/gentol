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

var (
	dbMap    *sync.Map
	mockMode = false

	ErrDBConfigIsNil   = errors.New("db config is nil")
	ErrDBInstanceIsNil = errors.New("db instance is nil")
	ErrSqlDBIsNil      = errors.New("sql db is nil")
	ErrDBNameIsEmpty   = errors.New("empty db name")
)

func init() {
	dbMap = &sync.Map{}
}

// InitWithDB init database instance with db instance
func InitWithDB(dbName string, db *gorm.DB) error {
	if db == nil {
		return errors.New("no db")
	}
	if dbName == "" {
		return errors.New("no db name")
	}
	_, ok := dbMap.Load(dbName)
	if ok {
		return nil
	}
	// Store database
	dbMap.Store(dbName, db)
	return nil
}

// InitWithConfig init database instance with db configuration and dialect
func InitWithConfig(config *Config) error {
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
	_, ok := dbMap.Load(config.DBName)
	if ok {
		return nil
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
		config.MaxOpenConn = defaultConfig.MaxOpenConn
	}
	if config.MaxIdleConn == 0 {
		config.MaxIdleConn = defaultConfig.MaxIdleConn
	}
	if config.ConnMaxLifeTime == 0 {
		config.ConnMaxLifeTime = defaultConfig.ConnMaxLifeTime
	}
	sqlDB.SetMaxOpenConns(config.MaxOpenConn)
	sqlDB.SetMaxIdleConns(config.MaxIdleConn)
	sqlDB.SetConnMaxLifetime(config.ConnMaxLifeTime)
	// Store database
	dbMap.Store(config.DBName, db)
	return nil
}

func GetDB(name string) (*gorm.DB, error) {
	db, ok := dbMap.Load(name)
	if !ok {
		return nil, errors.New("no db instance")
	}
	return db.(*gorm.DB), nil
}

func GetDBWithPanic(name string) *gorm.DB {
	db, ok := dbMap.Load(name)
	if !ok {
		panic("no db instance")
	}
	return db.(*gorm.DB)
}

func Close(dbName string) error {
	if dbName == "" {
		return ErrDBNameIsEmpty
	}
	v, ok := dbMap.LoadAndDelete(dbName)
	if !ok || v == nil {
		return nil
	}
	db, err := v.(*gorm.DB).DB()
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
		return ErrDBInstanceIsNil
	}
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	if sqlDB == nil {
		return ErrSqlDBIsNil
	}
	return sqlDB.Ping()
}
