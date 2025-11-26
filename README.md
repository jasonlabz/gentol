# Golang代码生成工具：支持数据库model、dao层代码生成&&gin项目生成

## 概述

**gentol**工具旨在简化Golang中一些通用化的代码逻辑编写过程，提高开发效率。其中包含常用数据库的model和dao层的代码生成过程、gin项目初始化生成过程。对于初建项目时，它简单生成一套可用的gin程序代码，通过简单配置即可运行。对于数据库操作代码逻辑，它通过分析数据库结构，自动生成相应的model和dao层代码，让您能够快速创建数据库访问层。目前工具在不定期更新中

## 功能特点

- 支持初始化gin项目代码，简化项目搭建成本。
- 支持多种数据库类型，如达梦数据库、MySQL、PostgreSQL、Oracle、SqlServer、sqlite
- 支持以shell命令的方式使用，或者下载源码，通过配置table.yaml生成代码
- 根据数据库表结构自动生成相应的model和dao层代码
- 支持生成符合项目规范的代码，如命名规范、注释等

## 如何使用

#### 1. 下载并安装该工具。
```shell
go install github.com/jasonlabz/gentol@master
```
#### 2. 使用工具。
##### 用法一：生成gin项目
```shell
gentol new|init [project_name|module_name]

例如：gentol new projectA
     gentol new github.com/XXX/projectB
     
     
生成项目：
➜  awesomeProject $ gentol new github.com/jasonlabz/demo
writing demo/cmd/demo_program/main.go
writing demo/conf/application.yaml
writing demo/Makefile
writing demo/conf/log/service.yaml
writing demo/conf/servicer/demo.yaml
writing demo/bootstrap/bootstrap.go
writing demo/common/consts/constant.go
writing demo/common/ginx/response.go
writing demo/common/ginx/page.go
writing demo/common/helper/context.go
writing demo/docs/docs.go
writing demo/global/resource/resource.go
writing demo/server/controller/health_check.go
writing demo/server/middleware/logger.go
writing demo/server/middleware/context.go
writing demo/server/routers/router.go
writing demo/server/service/health_check.go
writing demo/server/service/health_check/health_check_impl.go
writing demo/server/service/health_check/dto/request.go
writing demo/server/service/health_check/dto/response.go
writing demo/go.mod
writing demo/main.go
writing demo/README.md

➜  awesomeProject $ cd demo
➜  demo $ ls
Makefile  README.md  bootstrap  cmd  common  conf  docs  global  go.mod  main.go  server
➜  demo $ ls -l
total 48
-rw-r--r-- 1 root root 1753 Nov 26 18:42 Makefile
-rw-r--r-- 1 root root 2533 Nov 26 18:42 README.md
drw-r--r-- 2 root root 4096 Nov 26 18:42 bootstrap
drw-r--r-- 3 root root 4096 Nov 26 18:42 cmd
drw-r--r-- 5 root root 4096 Nov 26 18:42 common
drw-r--r-- 5 root root 4096 Nov 26 18:42 conf
drw-r--r-- 2 root root 4096 Nov 26 18:42 docs
drw-r--r-- 3 root root 4096 Nov 26 18:42 global
-rw-r--r-- 1 root root 4948 Nov 26 18:42 go.mod
-rw-r--r-- 1 root root 3565 Nov 26 18:42 main.go
drw-r--r-- 6 root root 4096 Nov 26 18:42 server

➜  demo $ tree
.
|-- Makefile
|-- README.md
|-- bootstrap
|   `-- bootstrap.go
|-- cmd
|   `-- demo_program
|       `-- main.go
|-- common
|   |-- consts
|   |   `-- constant.go
|   |-- ginx
|   |   |-- page.go
|   |   `-- response.go
|   `-- helper
|       `-- context.go
|-- conf
|   |-- application.yaml
|   |-- log
|   |   `-- service.yaml
|   |-- schema
|   `-- servicer
|       `-- demo.yaml
|-- docs
|   `-- docs.go
|-- global
|   `-- resource
|       `-- resource.go
|-- go.mod
|-- main.go
`-- server
    |-- controller
    |   `-- health_check.go
    |-- middleware
    |   |-- context.go
    |   `-- logger.go
    |-- routers
    |   `-- router.go
    `-- service
        |-- health_check
        |   |-- dto
        |   |   |-- request.go
        |   |   `-- response.go
        |   `-- health_check_impl.go
        `-- health_check.go

21 directories, 23 files

```


##### 用法二：生成dao、model层代码
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
tips: 当提供`--dsn`选项后，无需`--host --port --username --password`；
`--model --dao`为model和dao层的生成路径；
`--gen_hook`参数可以生成对应model的hook文件；

示例: 

**postgresql**

`gentol --db_type="postgres" --dsn="user=postgres password=XXXXX host=127.0.0.1 port=8432 dbname=dbName sslmode=disable TimeZone=Asia/Shanghai" --schema="public"`

**mysql**

`gentol --db_type="mysql" --dsn="%s:%s@tcp(%s:%d)/%s?parseTime=True&loc=Local" --table="table1,table2" --only_model`

3、sqlite
`gentol --db_type="sqlite" --dsn="/path/dagine.db" --table="table1,table2" --dao=/etl/dal/db/dao --model=/etl/dal/db/model --gen_hook`

4、dm
`gentol --db_type="dm" --dsn="dm://%s:%s@%s:%d?schema=%s" --table="table1,table2" --gen_hook`

5、oracle
`gentol --db_type="oracle" --dsn="%s/%s@%s:%d/%s" --table="table1,table2" --gen_hook`

6、sqlserver
`gentol --db_type="sqlserver" --dsn="user id=%s;password=%s;server=%s;port=%d;database=%s;encrypt=disable" --table="table1,table2"  --gen_hook`

- gentol工具在提供`db_type、dsn`参数情况下会生成当前数据库（当前模式）下所有表的model以及dao层代码，默认生成路径为`dal/db/dao,dal/model`,
可以通过参数`--model \ --dao`修改， `--table="table1,table2"`可以指定表列表生成(不提供该参数时生成当前schema下所有table)。
`--use_sql_nullable`可以替换guregu

```go
const (
    DBTypeOracle    DBType = "oracle"
    DBTypePostgres  DBType = "postgres"
    DBTypeMySQL     DBType = "mysql"
    DBTypeSqlserver DBType = "sqlserver"
    DBTypeGreenplum DBType = "greenplum"
    DBTypeSQLite    DBType = "sqlite"
    DatabaseTypeDM  DBType = "dm"
)

// DatabaseDsnMap 关系型数据库类型  username、password、address、port、dbname
var DatabaseDsnMap = map[DBType]string{
    DBTypeSQLite:    "%s",
    DatabaseTypeDM:  "dm://%s:%s@%s:%d?schema=%s",
    DBTypeOracle:    "%s/%s@%s:%d/%s",
    DBTypeMySQL:     "%s:%s@tcp(%s:%d)/%s?parseTime=True&loc=Local&timeout=30s",
    DBTypePostgres:  "user=%s password=%s host=%s port=%d dbname=%s sslmode=disable TimeZone=Asia/Shanghai",
    DBTypeSqlserver: "user id=%s;password=%s;server=%s;port=%d;database=%s;encrypt=disable",
    DBTypeGreenplum: "user=%s password=%s host=%s port=%d dbname=%s sslmode=disable TimeZone=Asia/Shanghai",
}

```

3、生成示例
- model生成结果：
```go
// Code generated by jasonlabz/gentol. DO NOT EDIT.
// Code generated by jasonlabz/gentol. DO NOT EDIT.
// Code generated by jasonlabz/gentol. DO NOT EDIT.

package model

import (
	"database/sql"
	"time"

	"github.com/jasonlabz/null"
	"github.com/satori/go.uuid"
)

var (
	_ = time.Second
	_ = sql.LevelDefault
	_ = null.Bool{}
	_ = uuid.UUID{}
)

type UserField string

// User struct is mapping to the user table
type User struct {
	UserID int64 `gorm:"primaryKey;autoIncrement;column:user_id;not null;type:bigint(20)" json:"user_id"` // Comment: 用户ID

	Nickname string `gorm:"column:nickname;not null;type:varchar(255)" json:"nickname"` // Comment: 用户名

	Avatar string `gorm:"column:avatar;not null;type:varchar(255)" json:"avatar"` // Comment: 头像

	Password string `gorm:"column:password;not null;type:varchar(255)" json:"password"` // Comment: 用户密码 des/md5加密值

	Phone string `gorm:"unique;column:phone;not null;type:varchar(255);uniqueIndex:phone" json:"phone"` // Comment: 手机号 aes加密

	Gender int32 `gorm:"column:gender;not null;type:int(11);default:9" json:"gender"` // Comment: 性别 0|男、1|女、9|未知

	Status int32 `gorm:"column:status;not null;type:int(11);default:0" json:"status"` // Comment: 状态 0|正常、1|注销、2|冻结

	RegisterIP string `gorm:"column:register_ip;not null;type:varchar(255)" json:"register_ip"` // Comment: 注册IP

	RegisterTime time.Time `gorm:"column:register_time;not null;type:timestamp;default:current_timestamp()" json:"register_time"` // Comment: 注册时间

	LastLoginIP string `gorm:"column:last_login_ip;not null;type:varchar(255)" json:"last_login_ip"` // Comment: 最后登录IP

	LastLoginTime time.Time `gorm:"column:last_login_time;not null;type:timestamp;default:current_timestamp()" json:"last_login_time"` // Comment: 最后登录时间

	CreateTime time.Time `gorm:"column:create_time;not null;type:timestamp;default:current_timestamp()" json:"create_time"` // Comment: no comment

	UpdateTime time.Time `gorm:"column:update_time;not null;type:timestamp;default:current_timestamp()" json:"update_time"` // Comment: no comment
}

func (u *User) TableName() string {
	return "user"
}

type UserTableColumn struct {
	UserID        UserField
	Nickname      UserField
	Avatar        UserField
	Password      UserField
	Phone         UserField
	Gender        UserField
	Status        UserField
	RegisterIP    UserField
	RegisterTime  UserField
	LastLoginIP   UserField
	LastLoginTime UserField
	CreateTime    UserField
	UpdateTime    UserField
}

type UserCondition struct {
	Condition
}

func (u *UserCondition) ColumnInfo() UserTableColumn {
	return UserTableColumn{
		UserID:        "user_id",
		Nickname:      "nickname",
		Avatar:        "avatar",
		Password:      "password",
		Phone:         "phone",
		Gender:        "gender",
		Status:        "status",
		RegisterIP:    "register_ip",
		RegisterTime:  "register_time",
		LastLoginIP:   "last_login_ip",
		LastLoginTime: "last_login_time",
		CreateTime:    "create_time",
		UpdateTime:    "update_time",
	}
}

func (u *UserCondition) UserIDIsNull() *UserCondition {
	return u.Where("user_id is null")
}

func (u *UserCondition) UserIDIsNotNull() *UserCondition {
	return u.Where("user_id is not null")
}

func (u *UserCondition) UserIDEqualTo(value int64) *UserCondition {
	return u.Where("user_id = ?", value)
}

func (u *UserCondition) UserIDNotEqualTo(value int64) *UserCondition {
	return u.Where("user_id <> ?", value)
}

func (u *UserCondition) UserIDGreaterThan(value int64) *UserCondition {
	return u.Where("user_id > ?", value)
}

func (u *UserCondition) UserIDGreaterThanOrEqualTo(value int64) *UserCondition {
	return u.Where("user_id >= ?", value)
}

func (u *UserCondition) UserIDLessThan(value int64) *UserCondition {
	return u.Where("user_id < ?", value)
}

.... 省略其他字段方法


func (u *UserCondition) Where(query any, args ...any) *UserCondition {
	switch obj := query.(type) {
	case string:
		if len(args) > 0 {
			u.StringCondition = append(u.StringCondition, obj)
			u.Condition.Args = append(u.Condition.Args, args...)
		} else {
			u.StringCondition = append(u.StringCondition, obj)
		}
	case map[string]any:
		if u.MapCondition == nil {
			u.MapCondition = obj
		} else {
			for key, val := range obj {
				u.MapCondition[key] = val
			}
		}
	}
	return u
}

func (u *UserCondition) OrderBy(orderByClause ...string) *UserCondition {
	u.OrderByClause = append(u.OrderByClause, orderByClause...)
	return u
}

func (u *UserCondition) GroupBy(groupByClause string) *UserCondition {
	u.GroupByClause = groupByClause
	return u
}

func (u *UserCondition) Having(query string, args ...any) *UserCondition {
	u.HavingCondition = query
	u.HavingArgs = args
	return u
}

func (u *UserCondition) Joins(query string, args ...any) *UserCondition {
	u.JoinCondition = append(u.JoinCondition, query)
	u.JoinArgs = append(u.JoinArgs, args...)
	return u
}

func (u *UserCondition) Build() *Condition {
	return &u.Condition
}

```
- hook 文件
需要生成hook文件，只需要加上 --gen_hook参数
  BeforeSave方法可以做入库前加密操作； AfterFind方法可以做出库后解密操作
```go
// Code generated by jasonlabz/gentol. You may edit it.

package model

import (
	"github.com/jasonlabz/potato/cryptox/aes"
	"gorm.io/gorm"
)

// BeforeSave invoked before saving, return an error.
func (u *User) BeforeSave(tx *gorm.DB) (err error) {
	u.Password, err = aes.Encrypt(u.Password)
	return
}

// AfterSave invoked after saving, return an error if field is not populated.
func (u *User) AfterSave(tx *gorm.DB) (err error) {
	// TODO: something
	return
}

// BeforeCreate invoked before create, return an error.
func (u *User) BeforeCreate(tx *gorm.DB) (err error) {
	// TODO: something
	return
}

// AfterCreate invoked after create, return an error.
func (u *User) AfterCreate(tx *gorm.DB) (err error) {
	// TODO: something
	return
}

// BeforeUpdate invoked before update, return an error.
func (u *User) BeforeUpdate(tx *gorm.DB) (err error) {
	// TODO: something
	return
}

// AfterUpdate invoked after update, return an error.
func (u *User) AfterUpdate(tx *gorm.DB) (err error) {
	// TODO: something
	return
}

// BeforeDelete invoked before delete, return an error.
func (u *User) BeforeDelete(tx *gorm.DB) (err error) {
	// TODO: something
	return
}

// AfterDelete invoked after delete, return an error.
func (u *User) AfterDelete(tx *gorm.DB) (err error) {
	// TODO: something
	return
}

// AfterFind invoked after find, return an error.
func (u *User) AfterFind(tx *gorm.DB) (err error) {
	u.Password, err = aes.Decrypt(u.Password)
	return
}

```

- dao层：
```go
//  ...
type UserDao interface {
	// SelectAll 查询所有记录
	SelectAll(ctx context.Context, selectFields ...model.UserField) (records []*model.User, err error)

	// SelectOneByPrimaryKey 通过主键查询记录
	SelectOneByPrimaryKey(ctx context.Context, userID int32, selectFields ...model.UserField) (record *model.User, err error)

	// SelectRecordByCondition 通过指定条件查询记录
	SelectRecordByCondition(ctx context.Context, condition *model.Condition, selectFields ...model.UserField) (records []*model.User, err error)

	// SelectPageRecordByCondition 通过指定条件查询分页记录
	SelectPageRecordByCondition(ctx context.Context, condition *model.Condition, pageParam *model.Pagination,
		selectFields ...model.UserField) (records []*model.User, err error)

	// CountByCondition 通过指定条件查询记录数量
	CountByCondition(ctx context.Context, condition *model.Condition) (count int64, err error)

	// DeleteByCondition 通过指定条件删除记录，返回删除记录数量
	DeleteByCondition(ctx context.Context, condition *model.Condition) (affect int64, err error)

	// DeleteByPrimaryKey 通过主键删除记录，返回删除记录数量
	DeleteByPrimaryKey(ctx context.Context, userID int32) (affect int64, err error)

	// UpdateRecord 更新记录
	UpdateRecord(ctx context.Context, record *model.User) (affect int64, err error)

	// UpdateRecords 批量更新记录
	UpdateRecords(ctx context.Context, records []*model.User) (affect int64, err error)

	// UpdateByCondition 更新指定条件下的记录
	UpdateByCondition(ctx context.Context, condition *model.Condition, updateField *model.UpdateField) (affect int64, err error)

	// UpdateByPrimaryKey 更新主键的记录
	UpdateByPrimaryKey(ctx context.Context, userID int32, updateField *model.UpdateField) (affect int64, err error)

	// Insert 插入记录
	Insert(ctx context.Context, record *model.User) (affect int64, err error)

	// BatchInsert 批量插入记录
	BatchInsert(ctx context.Context, records []*model.User) (affect int64, err error)

	// InsertOrUpdateOnDuplicateKey 插入记录，假如唯一键冲突则更新
	InsertOrUpdateOnDuplicateKey(ctx context.Context, record *model.User) (affect int64, err error)

	// BatchInsertOrUpdateOnDuplicateKey 批量插入记录，假如唯一键冲突则更新
	BatchInsertOrUpdateOnDuplicateKey(ctx context.Context, records []*model.User) (affect int64, err error)
}

```
- 使用示例：
```go
	logger := log.CtxLogger(ctx)
	defer logger.Sync()
	user, err = s.userDao.SelectOneByPrimaryKey(ctx, userID)

	// 构建查询条件
	userCondition := &model.UserCondition{}
	userCondition.UserIDEqualTo(userID).GenderEqualTo(1)
	// 提供查询指定字段， 默认查询该表全部字段：s.userDao.SelectRecordByCondition(ctx, userCondition.Build()）
	column := userCondition.ColumnInfo()
	userList, err := s.userDao.SelectRecordByCondition(ctx, userCondition.Build(), column.UserID, column.Nickname, column.Password)
	if err != nil {
		logger.WithError(err).Errorf("get user error")
	}

```

## 注意事项

- 请确保您的开发环境已安装Golang开发环境，并已安装相应的数据库驱动程序。
- 该工具生成的代码仅供参考，您可能需要根据实际需求进行修改和调整。
- 对于复杂的数据库结构和业务逻辑，可能需要手动编写代码或使用其他工具。

- **godror运行报错解决方案**：修改交插编译参数 go env -w CGO_ENABLED=1



## 反馈和支持

如果您在使用该工具时遇到任何问题或建议，欢迎提出建议。
我们将尽快回复并提供必要的支持和解决方案。
