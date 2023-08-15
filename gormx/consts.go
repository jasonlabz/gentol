package gormx

type LogMode string

type DBType string

const (
	DBTypeOracle    DBType = "oracle"
	DBTypePostgres  DBType = "postgres"
	DBTypeMySQL     DBType = "mysql"
	DBTypeSqlserver DBType = "sqlserver"
	DBTypeGreenplum DBType = "greenplum"
)

// DatabaseDsnMap 关系型数据库类型  username、password、address、port、dbname
var DatabaseDsnMap = map[DBType]string{
	DBTypeOracle:    "%s/%s@%s:%d/%s",
	DBTypeMySQL:     "%s:%s@tcp(%s:%d)/%s?parseTime=True&loc=Local",
	DBTypePostgres:  "user=%s password=%s host=%s port=%d dbname=%s sslmode=disable TimeZone=Asia/Shanghai",
	DBTypeSqlserver: "user id=%s;password=%s;server=%s;port=%d;database=%s",
	DBTypeGreenplum: "user=%s password=%s host=%s port=%d dbname=%s sslmode=disable TimeZone=Asia/Shanghai",
}

// JDBCUrlMap 关系型数据库类型  username、password、address、port、dbname
var JDBCUrlMap = map[DBType]string{
	DBTypeOracle:    "jdbc:oracle:thin:@%s:%d/%s",
	DBTypeMySQL:     "jdbc:mysql://%s:%d/%s?parseTime=True&loc=Local",
	DBTypePostgres:  "jdbc:postgresql://%s:%d/%s",
	DBTypeSqlserver: "jdbc:sqlserver://%s:%d;DatabaseName=%s",
	DBTypeGreenplum: "jdbc:postgresql://%s:%d/%s",
}
