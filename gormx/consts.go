package gormx

type LogMode string

type DBType string

const (
	DBTypeOracle    DBType = "oracle"
	DBTypePostgres  DBType = "postgres"
	DBTypeMySQL     DBType = "mysql"
	DBTypeSqlserver DBType = "sqlserver"
	DBTypeGreenplum DBType = "greenplum"
	DBTypeSQLite    DBType = "sqlite"
)

// DatabaseDsnMap 关系型数据库类型  username、password、address、port、dbname
var DatabaseDsnMap = map[DBType]string{
	DBTypeSQLite:    "%s",
	DBTypeOracle:    "%s/%s@%s:%d/%s",
	DBTypeMySQL:     "%s:%s@tcp(%s:%d)/%s?parseTime=True&loc=Local&timeout=30s",
	DBTypePostgres:  "user=%s password=%s host=%s port=%d dbname=%s sslmode=disable TimeZone=Asia/Shanghai",
	DBTypeSqlserver: "user id=%s;password=%s;server=%s;port=%d;database=%s;encrypt=disable",
	DBTypeGreenplum: "user=%s password=%s host=%s port=%d dbname=%s sslmode=disable TimeZone=Asia/Shanghai",
}

// JDBCUrlMap 关系型数据库类型  username、password、address、port、dbname
var JDBCUrlMap = map[DBType]string{
	DBTypeOracle:    "jdbc:oracle:thin:@%s:%d/%s",
	DBTypeMySQL:     "jdbc:mysql://%s:%d/%s?charset=utf8mb4&parseTime=True&loc=Local&timeout=30s",
	DBTypePostgres:  "jdbc:postgresql://%s:%d/%s",
	DBTypeSqlserver: "jdbc:sqlserver://%s:%d;DatabaseName=%s;encrypt=disable",
	DBTypeGreenplum: "jdbc:postgresql://%s:%d/%s",
}
