# gentol - Golang 代码生成工具

支持数据库 model/dao 层代码生成 && Gin 项目脚手架生成

## 概述

**gentol** 旨在简化 Golang 通用代码编写，提高开发效率。提供两大核心能力：

- **Gin 项目脚手架**：从模板仓库克隆项目骨架，自动替换模块路径和项目名，一键生成可运行的 Gin 项目
- **数据库代码生成**：连接数据库自动分析表结构，生成 GORM Model、DAO 接口/实现、Condition Builder 等代码

## 安装

```shell
go install github.com/jasonlabz/gentol@master
```

---

## 一、Gin 项目生成

### 1.1 基本用法

```shell
gentol new|init <project_name|module_name>
```

默认从内置模板仓库克隆项目到内存，替换模块路径和项目名后写入磁盘。

```shell
# 短项目名
gentol new myapp

# 完整模块路径
gentol new github.com/myorg/myapp
```

### 1.2 指定模板源

| 参数 | 说明 |
|------|------|
| `--template_repo=<git_url>` | 从指定 Git 仓库克隆模板 |
| `--template_dir=<local_path>` | 从本地目录加载模板 |

```shell
# 从指定 Git 仓库克隆模板
gentol new github.com/myorg/myapp --template_repo=https://github.com/xxx/my-template.git

# 从本地目录加载模板（开发调试用）
gentol new github.com/myorg/myapp --template_dir=/path/to/template
```

### 1.3 替换规则

gentol 会自动读取模板项目 `go.mod` 中的 `module` 路径，执行以下替换：

| 上下文 | 替换方式 | 示例 |
|--------|----------|------|
| Go 文件 import 语句 | 完整模块路径 | `template/bootstrap` → `github.com/myorg/myapp/bootstrap` |
| go.mod module 行 | 完整模块路径 | `module template` → `module github.com/myorg/myapp` |
| Go 文件非 import 行 | 项目短名称 | 注释、字符串中的项目名 |
| Makefile / YAML / 其他文件 | 项目短名称 | `TARGETNAME = myapp`、`name: myapp` |
| 文件名 / 目录名 | 项目短名称 | `demo_program/` → `myapp_program/`、`demo.yaml` → `myapp.yaml` |

### 1.4 更新项目

```shell
# 在项目目录内执行
gentol update

# 在外层指定项目名
gentol update myapp
gentol update github.com/myorg/myapp
```

### 1.5 添加 Service / Manager

```shell
# 添加 service（在项目目录内执行）
gentol add user
gentol add user_service    # 同上

# 添加 manager（可调用多个 service，位于 controller 和 service 之间）
gentol add user_manager
```

生成的文件结构：

```
server/service/
├── user_service.go               # 接口定义 + sync.Once 单例 Getter
└── user/
    ├── user_service_impl.go      # 实现
    └── body/
        ├── request.go
        ├── response.go
        ├── vo.go
        └── dto.go

server/manager/
├── user_manager.go               # 接口定义 + sync.Once 单例 Getter
└── user/
    ├── user_manager_impl.go      # 实现
    └── body/
        ├── request.go
        ├── response.go
        ├── vo.go
        └── dto.go
```

### 1.6 内存化流程

项目生成采用内存化处理，不产生临时目录残留：

```
加载模板到内存（git clone / 本地读取）
  → 内存中替换模块路径 + 项目名 + 文件路径
  → 一次性写入磁盘目标目录
  → 执行 go mod tidy
```

### 1.7 模板项目维护

只需维护一个标准的 Gin 项目作为模板，push 到 Git 仓库即可。模板项目的唯一约定：

- `go.mod` 中的 `module` 行定义了模块路径，gentol 自动读取并替换
- 建议模板使用完整模块路径（如 `github.com/jasonlabz/gentol-template`），这样替换逻辑最精确

---

## 二、数据库代码生成

### 2.1 基本用法

不指定子命令时进入 DB 代码生成模式：

```shell
gentol --db_type=<type> --dsn=<connection_string> [options...]
```

也可以通过 YAML 配置文件（`conf/table.yaml` 或 `./table.yaml`）指定生成参数。

### 2.2 参数说明

| 参数 | 简写 | 默认值 | 说明 |
|------|------|--------|------|
| `--db_type` | | `postgres` | 数据库类型：mysql, postgres, sqlserver, oracle, greenplum, sqlite, dm |
| `--dsn` | | | 数据库连接字符串，提供后无需 host/port/username/password |
| `--host` | `-h` | | 数据库主机 |
| `--port` | `-p` | | 数据库端口 |
| `--username` | `-U` | | 数据库用户名 |
| `--password` | `-P` | | 数据库密码 |
| `--database` | `-d` | | 数据库名 |
| `--schema` | `-s` | | Schema 名（PostgreSQL/Oracle 等） |
| `--table` | `-t` | | 表名列表（逗号分隔），不提供则生成当前 schema 下所有表 |
| `--model` | | `dal/db/model` | Model 层输出路径 |
| `--dao` | | `dal/db/dao` | DAO 层输出路径 |
| `--service` | | `server/service` | Service 层输出路径 |
| `--module` | `-m` | | Go module 名（用于 import 路径） |
| `--json_format` | | `snake` | JSON tag 命名格式：snake / upper_camel / lower_camel |
| `--protobuf_format` | | `snake` | Protobuf tag 命名格式 |
| `--only_model` | | | 仅生成 Model，不生成 DAO |
| `--gen_hook` | | | 生成 GORM Hook 文件 |
| `--use_sql_nullable` | | | 使用 sql.Null 类型替代 guregu/null |
| `--proto` | | | 添加 Protobuf 注解 |
| `--rungofmt` | | | 生成后执行 gofmt |

### 2.3 各数据库连接示例

**PostgreSQL**

```shell
gentol --db_type="postgres" \
  --dsn="user=postgres password=XXXXX host=127.0.0.1 port=5432 dbname=mydb sslmode=disable" \
  --schema="public"
```

**MySQL**

```shell
gentol --db_type="mysql" \
  --dsn="root:password@tcp(127.0.0.1:3306)/mydb?parseTime=True&loc=Local" \
  --table="users,orders"
```

**SQLite**

```shell
gentol --db_type="sqlite" \
  --dsn="/path/to/database.db" \
  --table="users" \
  --gen_hook
```

**Oracle**

```shell
gentol --db_type="oracle" \
  --dsn="username/password@host:1521/service_name" \
  --table="USERS" \
  --gen_hook
```

**SQL Server**

```shell
gentol --db_type="sqlserver" \
  --dsn="user id=sa;password=XXX;server=127.0.0.1;port=1433;database=mydb;encrypt=disable" \
  --table="users"
```

**DM（达梦）**

```shell
gentol --db_type="dm" \
  --dsn="dm://username:password@host:5236?schema=SCHEMA_NAME" \
  --table="USERS"
```

**DSN 格式参考**

```go
var DatabaseDsnMap = map[DBType]string{
    DBTypeSQLite:    "%s",
    DatabaseTypeDM:  "dm://%s:%s@%s:%d?schema=%s",
    DBTypeOracle:    "%s/%s@%s:%d/%s",
    DBTypeMySQL:     "%s:%s@tcp(%s:%d)/%s?parseTime=True&loc=Local&timeout=30s",
    DBTypePostgres:  "user=%s password=%s host=%s port=%d dbname=%s sslmode=disable",
    DBTypeSqlserver: "user id=%s;password=%s;server=%s;port=%d;database=%s;encrypt=disable",
    DBTypeGreenplum: "user=%s password=%s host=%s port=%d dbname=%s sslmode=disable",
}
```

### 2.4 生成结果

每张表生成以下文件：

| 文件 | 路径 | 覆盖策略 |
|------|------|----------|
| `{table}.go` | `dal/db/model/` | 始终覆盖 |
| `{table}_hook.go` | `dal/db/model/` | 仅首次生成（可手动编辑） |
| `base.go` | `dal/db/model/` | 始终覆盖 |
| `{table}_dao.go` | `dal/db/dao/` | 始终覆盖 |
| `{table}_dao_ext.go` | `dal/db/dao/` | 仅首次生成（可手动扩展） |
| `{table}_dao_impl.go` | `dal/db/dao/impl/` | 始终覆盖 |
| `{table}_dao_ext_impl.go` | `dal/db/dao/impl/` | 仅首次生成（可手动扩展） |
| `db.go` | `dal/db/dao/` | 始终覆盖 |

**Model 生成示例**

```go
// Code generated by jasonlabz/gentol. DO NOT EDIT.

package model

type UserField string

type User struct {
    UserID   int64  `gorm:"primaryKey;autoIncrement;column:user_id;not null;type:bigint" json:"user_id"`
    Nickname string `gorm:"column:nickname;not null;type:varchar(255)" json:"nickname"`
    Phone    string `gorm:"unique;column:phone;not null;type:varchar(255)" json:"phone"`
    // ...
}

type UserCondition struct { Condition }

// 流式查询构建
func (u *UserCondition) UserIDEqualTo(value int64) *UserCondition {
    return u.Where("user_id = ?", value)
}

func (u *UserCondition) NicknamePrefixLike(value string) *UserCondition {
    return u.Where("nickname like ?", value+"%")
}
```

**DAO 接口生成示例**

```go
type UserDao interface {
    UserDaoExt  // 用户自定义扩展接口
    SelectAll(ctx context.Context, selectFields ...model.UserField) ([]*model.User, error)
    SelectOneByPrimaryKey(ctx context.Context, id int64, selectFields ...model.UserField) (*model.User, error)
    SelectRecordByCondition(ctx context.Context, condition *model.Condition, selectFields ...model.UserField) ([]*model.User, error)
    SelectPageRecordByCondition(ctx context.Context, condition *model.Condition, page *model.Pagination, selectFields ...model.UserField) ([]*model.User, error)
    CountByCondition(ctx context.Context, condition *model.Condition) (int64, error)
    Insert(ctx context.Context, record *model.User) (int64, error)
    BatchInsert(ctx context.Context, records []*model.User) (int64, error)
    UpdateRecord(ctx context.Context, record *model.User) (int64, error)
    UpdateByCondition(ctx context.Context, condition *model.Condition, updateField *model.UpdateField) (int64, error)
    DeleteByPrimaryKey(ctx context.Context, id int64) (int64, error)
    DeleteByCondition(ctx context.Context, condition *model.Condition) (int64, error)
    UpsertRecord(ctx context.Context, record *model.User) (int64, error)
    InsertOrUpdateOnDuplicateKey(ctx context.Context, record *model.User) (int64, error)
    // ...
}
```

**使用示例**

```go
// 按主键查询
user, err := userDao.SelectOneByPrimaryKey(ctx, userID)

// 条件查询
cond := &model.UserCondition{}
cond.UserIDEqualTo(userID).GenderEqualTo(1)

// 指定查询字段
col := cond.ColumnInfo()
users, err := userDao.SelectRecordByCondition(ctx, cond.Build(), col.UserID, col.Nickname)

// 分页查询
users, err := userDao.SelectPageRecordByCondition(ctx, cond.Build(), pagination)
```

---

## 注意事项

- 请确保已安装 Golang 开发环境及相应数据库驱动
- 生成的代码仅供参考，可能需要根据实际需求修改
- `_ext.go` 和 `_hook.go` 文件仅首次生成，后续不会被覆盖，可安全编辑
- Oracle 驱动运行报错时：`go env -w CGO_ENABLED=1`

## 反馈和支持

如有问题或建议，欢迎提 Issue 或 PR。
