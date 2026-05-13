package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/jasonlabz/gentol/datasource"
	"github.com/jasonlabz/gentol/gormx"
)

var (
	ddlKeywords = map[string]bool{
		"CREATE": true, "ALTER": true, "DROP": true,
		"TRUNCATE": true, "RENAME": true, "COMMENT": true,
	}
	dmlKeywords = map[string]bool{
		"INSERT": true, "UPDATE": true, "DELETE": true,
		"MERGE": true, "REPLACE": true, "CALL": true,
		"EXEC": true, "EXECUTE": true, "SELECT": true,
		"GRANT": true, "REVOKE": true, "LOAD": true,
	}
	commentPattern     = regexp.MustCompile(`/\*[\s\S]*?\*/`)
	lineCommentPattern = regexp.MustCompile(`--[^\n]*`)
	whitespaceOnly     = regexp.MustCompile(`^\s*$`)
)

const ddlDBName = "_ddl_db_"

func processDDL() {
	if len(os.Args) < 3 || strings.HasPrefix(os.Args[2], "--") {
		log.Fatal("用法: gentol ddl <sql文件路径> [--db_type=...] [--dsn=...] [--host=...] [--port=...] [--username=...] [--password=...] [--database=...]")
	}

	sqlFilePath := os.Args[2]
	dbConfig := parseDDLArgs()

	content, err := os.ReadFile(sqlFilePath)
	if err != nil {
		log.Fatalf("读取SQL文件失败: %v", err)
	}

	statements, err := validateDDL(string(content))
	if err != nil {
		log.Fatalf("DDL校验失败:\n%v", err)
	}

	log.Printf("校验通过，共 %d 条DDL语句", len(statements))

	db, err := gormx.LoadDBInstance(dbConfig)
	if err != nil {
		log.Fatalf("数据库连接失败: %v", err)
	}
	_ = db

	ds, err := datasource.GetDS(dbConfig.DBType)
	if err != nil {
		log.Fatalf("不支持的数据库类型 %s: %v", dbConfig.DBType, err)
	}

	ctx := context.Background()
	for i, stmt := range statements {
		log.Printf("[%d/%d] 执行中...", i+1, len(statements))
		if err := ds.ExecuteDDL(ctx, dbConfig.DBName, "", "", stmt); err != nil {
			log.Fatalf("执行失败 [%d/%d]: %v\nSQL: %s", i+1, len(statements), err, stmt)
		}
	}

	log.Printf("成功执行 %d 条DDL语句", len(statements))
}

// validateDDL checks that the SQL content contains only DDL statements.
// Returns the parsed statements and any validation error.
func validateDDL(sqlContent string) ([]string, error) {
	clean := stripComments(sqlContent)
	raw := strings.Split(clean, ";")
	statements := make([]string, 0, len(raw))

	for i, stmt := range raw {
		trimmed := strings.TrimSpace(stmt)
		if whitespaceOnly.MatchString(trimmed) {
			continue
		}

		keyword := extractKeyword(trimmed)
		upperKW := strings.ToUpper(keyword)

		if keyword == "" {
			continue
		}

		if dmlKeywords[upperKW] {
			return nil, fmt.Errorf("第 %d 条语句包含禁止的DML操作 [%s]，仅允许DDL语句", i+1, upperKW)
		}

		if !ddlKeywords[upperKW] {
			return nil, fmt.Errorf("第 %d 条语句以未知关键词 [%s] 开头，仅允许: CREATE, ALTER, DROP, TRUNCATE, RENAME, COMMENT", i+1, upperKW)
		}

		statements = append(statements, trimmed)
	}

	if len(statements) == 0 {
		return nil, fmt.Errorf("SQL文件中未找到有效的DDL语句")
	}

	return statements, nil
}

// stripComments removes SQL comments (single-line -- and multi-line /* */) from content.
func stripComments(content string) string {
	content = commentPattern.ReplaceAllString(content, "")
	content = lineCommentPattern.ReplaceAllString(content, "")
	return content
}

// extractKeyword returns the first SQL keyword from a statement (case-insensitive).
func extractKeyword(statement string) string {
	fields := strings.Fields(statement)
	if len(fields) == 0 {
		return ""
	}
	return fields[0]
}

// parseDDLArgs parses database connection flags from command-line arguments.
func parseDDLArgs() *gormx.Config {
	config := &gormx.Config{
		DBName: ddlDBName,
	}

	args := os.Args[3:]
	for i := range len(args) {
		arg := args[i]
		if strings.HasPrefix(arg, "--") {
			key, val := splitFlag(arg)
			switch key {
			case "db_type":
				config.DBType = gormx.DBType(val)
			case "dsn":
				config.DSN = val
			case "host":
				config.Host = val
			case "port":
				fmt.Sscanf(val, "%d", &config.Port)
			case "username":
				config.User = val
			case "password":
				config.Password = val
			case "database":
				config.Database = val
			}
		}
	}

	if config.DBType == "" {
		log.Fatal("缺少必要参数: --db_type")
	}

	if config.DSN == "" {
		if config.Host == "" || config.Port == 0 || config.User == "" || config.Database == "" {
			log.Fatal("缺少数据库连接参数，请提供 --dsn 或 (--host, --port, --username, --password, --database)")
		}
		config.GenDSN()
	}

	return config
}

// splitFlag splits "--key=value" or "--key" "value" style flags.
func splitFlag(arg string) (key, value string) {
	arg = strings.TrimPrefix(arg, "--")
	parts := strings.SplitN(arg, "=", 2)
	key = parts[0]
	if len(parts) == 2 {
		value = parts[1]
	}
	return
}
