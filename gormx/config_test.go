package gormx

import (
	"net/url"
	"testing"
)

// TestGenDSNTemplates 锁住「模板形状」与「Sprintf 参数顺序」的一致性。
// DatabaseDsnMap 是位置模板，参数按 user、password、host、port、dbname 固定顺序传入，
// 改模板时若增减了占位符而没同步改 GenDSN，这里会直接报错，
// 而不是在运行期拼出一个带 %!(EXTRA ...) 的坏字符串。
func TestGenDSNTemplates(t *testing.T) {
	cfg := func(dbType DBType) *Config {
		return &Config{
			DBName:   "demo",
			DBType:   dbType,
			Host:     "192.168.3.30",
			Port:     1521,
			User:     "lucas",
			Password: "openthedoor",
			Database: "XE",
		}
	}

	cases := []struct {
		dbType DBType
		want   string
	}{
		{DBTypeOracle, "oracle://lucas:openthedoor@192.168.3.30:1521/XE"},
		{DBTypeDM, "dm://lucas:openthedoor@192.168.3.30:1521?schema=XE"},
		{DBTypeMySQL, "lucas:openthedoor@tcp(192.168.3.30:1521)/XE?parseTime=True&loc=Local&timeout=30s"},
		{DBTypePostgres, "user=lucas password=openthedoor host=192.168.3.30 port=1521 dbname=XE sslmode=disable TimeZone=Asia/Shanghai"},
		{DBTypeGreenplum, "user=lucas password=openthedoor host=192.168.3.30 port=1521 dbname=XE sslmode=disable TimeZone=Asia/Shanghai"},
		{DBTypeSqlserver, "odbc:user id=lucas;password=openthedoor;server=192.168.3.30;port=1521;database=XE;encrypt=disable"},
	}

	for _, c := range cases {
		t.Run(string(c.dbType), func(t *testing.T) {
			if got := cfg(c.dbType).GenDSN(); got != c.want {
				t.Fatalf("GenDSN() =\n  %q\nwant:\n  %q", got, c.want)
			}
		})
	}

	// database 为空时回落到 db_name，避免拼出空的 service / dbname
	t.Run("database 回落 db_name", func(t *testing.T) {
		c := cfg(DBTypeOracle)
		c.Database = ""
		if got, want := c.GenDSN(), "oracle://lucas:openthedoor@192.168.3.30:1521/demo"; got != want {
			t.Fatalf("GenDSN() = %q, want %q", got, want)
		}
	})

	// 显式配了 dsn 时原样返回，不参与拼装
	t.Run("显式 DSN 优先", func(t *testing.T) {
		c := cfg(DBTypeOracle)
		c.DSN = "oracle://u:p@h:1521/svc"
		if got := c.GenDSN(); got != c.DSN {
			t.Fatalf("GenDSN() = %q, want %q", got, c.DSN)
		}
	})
}

// TestOracleDSNParseable 用 url 解析回读 Oracle DSN，确认 go-ora 需要的
// oracle://user:pass@host:port/service 形式各段都能取回原值。
func TestOracleDSNParseable(t *testing.T) {
	c := &Config{
		DBName:   "demo",
		DBType:   DBTypeOracle,
		Host:     "192.168.3.30",
		Port:     1521,
		User:     "lucas",
		Password: "openthedoor",
		Database: "XE",
	}
	u, err := url.Parse(c.GenDSN())
	if err != nil {
		t.Fatalf("url.Parse 失败: %v", err)
	}
	pass, _ := u.User.Password()
	if u.Scheme != "oracle" || u.User.Username() != "lucas" || pass != "openthedoor" ||
		u.Hostname() != "192.168.3.30" || u.Port() != "1521" || u.Path != "/XE" {
		t.Fatalf("DSN 各段还原不正确: scheme=%s user=%s pass=%s host=%s port=%s path=%s",
			u.Scheme, u.User.Username(), pass, u.Hostname(), u.Port(), u.Path)
	}
}
