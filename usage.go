package main

import "fmt"

// topLevelUsage 顶层帮助信息：列出所有子命令及简要说明
const topLevelUsage = `gentol - GORM model/dao code generator and project scaffolding tool

Usage:
  gentol                          generate model/dao code from a DB table (default mode)
  gentol new <module_path>        create a new project from template
  gentol init <module_path>       alias of "new"
  gentol update [module_path]     update an existing project from template
  gentol ddl <sql_file> [flags]   validate and execute DDL statements from a SQL file
  gentol help | -h | --help       show this help message

Run 'gentol <command> -h' for details on a specific command.
Run 'gentol -h' (no command) to see flags for the default db-generation mode.
`

// newUsage new/init 子命令帮助信息
const newUsage = `Usage: gentol new <module_path> [flags]
       gentol init <module_path> [flags]

Create a new project from a template repository or local directory.

Arguments:
  module_path              go module path for the new project, e.g. github.com/you/myapp

Flags:
      --template_repo=value  git repository URL to use as template (default: built-in template)
      --template_dir=value   local directory to use as template instead of cloning a repo
`

// updateUsage update 子命令帮助信息
const updateUsage = `Usage: gentol update [module_path] [flags]

Update an existing project from a template, overwriting same-named files
while keeping files that only exist in the project.

Arguments:
  module_path               optional; defaults to the module in ./go.mod of the current directory

Flags:
      --template_repo=value  git repository URL to use as template (default: built-in template)
      --template_dir=value   local directory to use as template instead of cloning a repo
`

// ddlUsage ddl 子命令帮助信息
const ddlUsage = `Usage: gentol ddl <sql_file> [flags]

Validate that a SQL file contains only DDL statements (CREATE, ALTER, DROP,
TRUNCATE, RENAME, COMMENT), then execute them against the target database.

Arguments:
  sql_file                  path to a .sql file containing DDL statements

Flags:
      --db_type=value        database type such as [mysql, sqlserver, postgres, oracle, greenplum etc.]
      --dsn=value             database connection string; if set, host/port/username/password/database are ignored
      --host=value            db host, if there is a dsn, ignore it
      --port=value            db port, if there is a dsn, ignore it
      --username=value        db username, if there is a dsn, ignore it
      --password=value        db password, if there is a dsn, ignore it
      --database=value        database name
      --schema=value          schema name
`

// isHelpFlag 判断参数是否为帮助标志
func isHelpFlag(arg string) bool {
	return arg == "-h" || arg == "--help" || arg == "help"
}

// hasHelpFlag 检查参数列表中是否包含帮助标志
func hasHelpFlag(args []string) bool {
	for _, arg := range args {
		if isHelpFlag(arg) {
			return true
		}
	}
	return false
}

func printTopLevelUsage() {
	fmt.Print(topLevelUsage)
}

func printSubUsage(usage string) {
	fmt.Print(usage)
}

// printSubcommandHint 在默认模式 usage 前提示子命令的存在
func printSubcommandHint() {
	fmt.Println("Tip: run 'gentol -h' to see all subcommands (new, init, update, ddl).")
}
