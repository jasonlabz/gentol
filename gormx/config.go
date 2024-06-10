package gormx

import (
	"fmt"
	"time"
)

var (
	defaultMaxOpenConn     = 500              // 最大连接数
	defaultMaxIdleConn     = 5                // 最大空闲连接数
	defaultConnMaxLifeTime = 10 * time.Minute // 连接最大存活时间
)

// Config Database configuration
type Config struct {
	DBName          string        `json:"db_name"`  // 数据库名称（要求唯一）
	JDBCUrl         string        `json:"jdbc_url"` // jdbc_url
	DSN             string        `json:"dsn"`      // 数据库连接串
	DBType          DBType        `json:"db_type"`
	Host            string        `json:"host"`
	Port            int           `json:"port"`
	User            string        `json:"user"`
	Password        string        `json:"password"`
	Database        string        `json:"database"`
	MaxOpenConn     int           `json:"max_open_conn"`      // 最大连接数
	MaxIdleConn     int           `json:"max_idle_conn"`      // 最大空闲连接数
	ConnMaxLifeTime time.Duration `json:"conn_max_life_time"` // 连接最大存活时间
	SSLMode         string        `json:"ssl_mode"`
	TimeZone        string        `json:"time_zone"`
	Charset         string        `json:"charset"`
}

func (c *Config) GenDSN() (dsn string) {
	if c.DSN != "" {
		return c.DSN
	}

	switch c.DBType {
	case DBTypeSQLite:
		//dsn = c.DSN
	default:
		dbName := c.Database
		if dbName == "" {
			dbName = c.DBName
		}
		dsnTemplate, ok := DatabaseDsnMap[c.DBType]
		if !ok {
			return
		}
		c.DSN = fmt.Sprintf(dsnTemplate, c.User, c.Password, c.Host, c.Port, dbName)
	}
	dsn = c.DSN
	return
}

func (c *Config) GenJDBCUrl() (jdbcUrl string) {
	if c.JDBCUrl != "" {
		return c.JDBCUrl
	}

	dbName := c.Database
	if dbName == "" {
		dbName = c.DBName
	}
	jdbcTemplate, ok := JDBCUrlMap[c.DBType]
	if !ok {
		return
	}
	jdbcUrl = fmt.Sprintf(jdbcTemplate, c.Host, c.Port, dbName)

	c.JDBCUrl = jdbcUrl
	return
}
