# Golang常用数据库的model、dao层低代码生成工具

## 概述

**gentol**工具旨在简化Golang中常用数据库的model和dao层的代码生成过程，提高开发效率。它通过分析数据库结构，自动生成相应的model和dao层代码，让您能够快速创建数据库访问层。目前工具在不定期更新中

## 功能特点

- 支持多种数据库类型，如MySQL、PostgreSQL、Oracle、SqlServer等
- 支持以shell命令的方式使用，或者下载源码，通过配置table.yaml生成代码
- 根据数据库表结构自动生成相应的model和dao层代码
- 可自定义代码生成模板，满足不同项目需求
- 支持生成符合项目规范的代码，如命名规范、注释等

## 如何使用

1. 下载并安装该工具。
```shell
go install github.com/jasonlabz/gentol@latest
```
2. 使用工具。
```shell
gentol [--dao value] [-d value] [--db_type value] [--dsn value] [--gogoproto value] [--grpc value] [-h value] [--json_format value] [--model value] [--only_model] [--out value] [-P value] [-p value] [--proto] [--protobuf_format value] [--rungofmt] [-s value] [--service value] [-t value] [--template_dir value] [--use_sql_nullable] [-U value] [parameters ...]
     --dao=value    name to set for dao package [dal/db/dao]
 -d, --database=value
                    database to for db table {check}
     --db_type=value
                    database type such as [mysql, sqlserver, postgres, oracle,
                    greenplum etc. ] [mysql] {check}
                    json name format [snake | upper_camel | lower_camel] [snake]
     --model=value  name to set for model package [dal/db/model]
     --only_model   overwrite existing files (default)
     --out=value    output dir [.]
 -P, --password=value
                    db password, if there is a dsn, ignore it
 -p, --port=value   db port, if there is a dsn, ignore it
     --rungofmt     run gofmt on output dir
 -s, --schema=value
                    schema to for db table [public]
     --service=value
                    name to set for service package [server/service]
 -t, --table=value  table name to build struct from [user] {check}
     --template_dir=value
                    Template Dir [./template]
     --use_sql_nullable
                    use sql.Null if use_sql_nullable true, default use guregu
 -U, --username=value
                    db username, if there is a dsn, ignore it
                    
 
```
tips: 当提供`--dsn`选项后，无需`--host --port --username --password`；`--model --dao`为model和dao层的生成路径，当给定绝对路径时需要给定`--module`,以便生成model的包路径。

example: `gentol --db_type="postgres" --dsn="user=postgres password=halojeff host=1.117.232.208 port=8432 dbname=lg_server sslmode=disable TimeZone=Asia/Shanghai" --schema="public"`

- gentol工具在提供`db_type、dsn`参数情况下会生成当前数据库（当前模式）下所有表的model以及dao层代码，默认生成路径为`dal/db/dao,dal/model`,可以通过参数`--model \ --dao`修改， `--table="table1,table2"`可以指定表列表生成。
## 注意事项

- 请确保您的开发环境已安装Golang开发环境，并已安装相应的数据库驱动程序。
- 该工具生成的代码仅供参考，您可能需要根据实际需求进行修改和调整。
- 对于复杂的数据库结构和业务逻辑，可能需要手动编写代码或使用其他工具。

- **godror运行报错解决方案**：修改交插编译参数 go env -w CGO_ENABLED=1



## 反馈和支持

如果您在使用该工具时遇到任何问题或建议，欢迎提出建议。
我们将尽快回复并提供必要的支持和解决方案。
