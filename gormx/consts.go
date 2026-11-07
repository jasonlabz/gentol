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
	DBTypeDM        DBType = "dm"
)

// DatabaseDsnMap 关系型数据库类型  username、password、address、port、dbname
// 注意：Oracle 走 sijms/go-ora，DSN 必须是 oracle:// 开头的 URL 形式（老 godror 的
// user/pass@host:port/service 不再被识别）。
// SQL Server 固定带 odbc: 前缀：不带前缀时 msdsn 按 ADO 规则以 ; 无脑切分、不认花括号，
// 值里含 ; 会被截断；带前缀才走支持 {} 转义的状态机。
var DatabaseDsnMap = map[DBType]string{
	DBTypeSQLite:    "%s",
	DBTypeDM:        "dm://%s:%s@%s:%d?schema=%s",
	DBTypeOracle:    "oracle://%s:%s@%s:%d/%s",
	DBTypeMySQL:     "%s:%s@tcp(%s:%d)/%s?parseTime=True&loc=Local&timeout=30s",
	DBTypePostgres:  "user=%s password=%s host=%s port=%d dbname=%s sslmode=disable TimeZone=Asia/Shanghai",
	DBTypeSqlserver: "odbc:user id=%s;password=%s;server=%s;port=%d;database=%s;encrypt=disable",
	DBTypeGreenplum: "user=%s password=%s host=%s port=%d dbname=%s sslmode=disable TimeZone=Asia/Shanghai",
}
