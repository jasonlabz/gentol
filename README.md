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
go install github.com/jasonlabz/gentol@v1.0.7
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
tips: 当提供`--dsn`选项后，无需`--host --port --username --password`；
`--model --dao`为model和dao层的生成路径，当给定绝对路径时需要给定`--module="module_name""`,以便生成model的包路径。
--gen_hook参数可以生成对应model的hook文件
example: `gentol --db_type="postgres" --dsn="user=postgres password=XXXXX host=127.0.0.1 port=8432 dbname=dbName sslmode=disable TimeZone=Asia/Shanghai" --schema="public"`

- gentol工具在提供`db_type、dsn`参数情况下会生成当前数据库（当前模式）下所有表的model以及dao层代码，默认生成路径为`dal/db/dao,dal/model`,可以通过参数`--model \ --dao`修改， `--table="table1,table2"`可以指定表列表生成(不提供该参数时生成当前schema下所有table)。

3、生成示例
- model生成结果：
```go
package model

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/guregu/null"
	"github.com/satori/go.uuid"
)

var (
	_ = time.Second
	_ = sql.LevelDefault
	_ = null.Bool{}
	_ = uuid.UUID{}
)

const TableNameUser = "user"

type UserField string

// User struct is mapping to the user table
type User struct {
	UserID int32 `gorm:"primaryKey;autoIncrement;column:user_id;not null;type:int4;size:32;" json:"user_id"` // Comment: 用户ID

	Nickname string `gorm:"column:nickname;not null;type:varchar;size:255;" json:"nickname"` // Comment: 用户名

	Avatar string `gorm:"column:avatar;not null;type:varchar;size:255;" json:"avatar"` // Comment: 头像

	Password string `gorm:"column:password;not null;type:varchar;size:255;" json:"password"` // Comment: 用户密码 des/md5加密值

	Phone string `gorm:"unique;column:phone;not null;type:varchar;size:255;" json:"phone"` // Comment: 手机号 aes加密

	Gender int32 `gorm:"column:gender;not null;type:int4;size:32;default:9;" json:"gender"` // Comment: 性别 0|男、1|女、9|未知

	Status int32 `gorm:"column:status;not null;type:int4;size:32;default:0;" json:"status"` // Comment: 状态 0|正常、1|注销、2|冻结

	RegisterIP string `gorm:"column:register_ip;not null;type:varchar;size:255;" json:"register_ip"` // Comment: no comment

	RegisterTime time.Time `gorm:"column:register_time;not null;type:timestamp;size:64;default:CURRENT_TIMESTAMP;" json:"register_time"` // Comment: no comment

	LastLoginIP string `gorm:"column:last_login_ip;not null;type:varchar;size:255;" json:"last_login_ip"` // Comment: no comment

	LastLoginTime time.Time `gorm:"column:last_login_time;not null;type:timestamp;size:64;default:CURRENT_TIMESTAMP;" json:"last_login_time"` // Comment: no comment

	CreateTime time.Time `gorm:"column:create_time;not null;type:timestamp;size:64;default:CURRENT_TIMESTAMP;" json:"create_time"` // Comment: no comment

	UpdateTime time.Time `gorm:"column:update_time;not null;type:timestamp;size:64;default:CURRENT_TIMESTAMP;" json:"update_time"` // Comment: no comment
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

func (u *UserCondition) UserIDEqualTo(value int32) *UserCondition {
	return u.Where("user_id = %v", value)
}

func (u *UserCondition) UserIDNotEqualTo(value int32) *UserCondition {
	return u.Where("user_id <> %v", value)
}

func (u *UserCondition) UserIDGreaterThan(value int32) *UserCondition {
	return u.Where("user_id > %v", value)
}

func (u *UserCondition) UserIDGreaterThanOrEqualTo(value int32) *UserCondition {
	return u.Where("user_id >= %v", value)
}

func (u *UserCondition) UserIDLessThan(value int32) *UserCondition {
	return u.Where("user_id < %v", value)
}

func (u *UserCondition) UserIDLessThanOrEqualTo(value int32) *UserCondition {
	return u.Where("user_id <= %v", value)
}

func (u *UserCondition) UserIDBetween(startValue, endValue int32) *UserCondition {
	return u.Where("user_id between %v and %v", startValue, endValue)
}

func (u *UserCondition) UserIDNotBetween(startValue, endValue int32) *UserCondition {
	return u.Where("user_id not between %v and %v", startValue, endValue)
}

func (u *UserCondition) UserIDIn(inValues []int32) *UserCondition {
	return u.Where(TransInCondition("user_id in ", inValues))
}

func (u *UserCondition) UserIDNotIn(inValues []int32) *UserCondition {
	return u.Where(TransInCondition("user_id not in ", inValues))
}

func (u *UserCondition) NicknameIsLike(value string) *UserCondition {
	return u.Where("nickname like '%v'", value)
}

func (u *UserCondition) NicknameIsNull() *UserCondition {
	return u.Where("nickname is null")
}

func (u *UserCondition) NicknameIsNotNull() *UserCondition {
	return u.Where("nickname is not null")
}

func (u *UserCondition) NicknameEqualTo(value string) *UserCondition {
	return u.Where("nickname = '%v'", value)
}

func (u *UserCondition) NicknameNotEqualTo(value string) *UserCondition {
	return u.Where("nickname <> '%v'", value)
}

func (u *UserCondition) NicknameGreaterThan(value string) *UserCondition {
	return u.Where("nickname > '%v'", value)
}

func (u *UserCondition) NicknameGreaterThanOrEqualTo(value string) *UserCondition {
	return u.Where("nickname >= '%v'", value)
}

func (u *UserCondition) NicknameLessThan(value string) *UserCondition {
	return u.Where("nickname < '%v'", value)
}

func (u *UserCondition) NicknameLessThanOrEqualTo(value string) *UserCondition {
	return u.Where("nickname <= '%v'", value)
}

func (u *UserCondition) NicknameBetween(startValue, endValue string) *UserCondition {
	return u.Where("nickname between '%v' and '%v'", startValue, endValue)
}

func (u *UserCondition) NicknameNotBetween(startValue, endValue string) *UserCondition {
	return u.Where("nickname not between '%v' and '%v'", startValue, endValue)
}

func (u *UserCondition) NicknameIn(inValues []string) *UserCondition {
	return u.Where(TransInCondition("nickname in ", inValues))
}

func (u *UserCondition) NicknameNotIn(inValues []string) *UserCondition {
	return u.Where(TransInCondition("nickname not in ", inValues))
}

// ..... 省略中间部分
	
func (u *UserCondition) Where(query any, args ...any) *UserCondition {
	switch query.(type) {
	case map[string]any:
		mapCondition := query.(map[string]any)
		if u.MapCondition == nil {
			u.MapCondition = mapCondition
			break
		}
		for key, val := range mapCondition {
			u.MapCondition[key] = val
		}
	case string:
		condition := query.(string)
		u.StringCondition = append(u.StringCondition, fmt.Sprintf(condition, args...))
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
	u.HavingCondition = fmt.Sprintf(query, args...)
	return u
}

func (u *UserCondition) Joins(query string, args ...any) *UserCondition {
	u.JoinCondition = append(u.JoinCondition, fmt.Sprintf(query, args...))
	return u
}

func (u *UserCondition) Build() *Condition {
	return &u.Condition
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
	// 提供查询指定字段
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
