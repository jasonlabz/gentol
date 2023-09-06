package metadata

import (
	"fmt"
	"strings"

	"github.com/jasonlabz/gentol/gormx"
)

type ModelMeta struct {
	BaseConfig
	ModelPackageName string
	ModelStructName  string
	ImportPkgList    []string
	ColumnList       []*ColumnInfo
}

type ColumnInfo struct {
	Index             int
	UpperTableName    string
	TitleTableName    string
	GoColumnName      string
	GoUpperColumnName string
	GoColumnType      string // string
	ValueFormat       string // string
	Tags              string
	ModelPackageName  string
	ModelStructName   string
	ModelShortName    string
	SchemaName        string
	TableName         string

	ColumnName         string
	SQLNullableType    string
	GureguNullableType string
	ColumnType         string // varchar(64)
	DataBaseType       string // varchar
	Length             int64  // 64
	IsPrimaryKey       bool
	Unique             bool
	AutoIncrement      bool
	Nullable           bool
	Comment            string
	DefaultValue       string
}

func (m *ModelMeta) GenRenderData() map[string]any {
	if m == nil {
		return map[string]any{}
	}
	useSQLNullable := m.UseSQLNullable
	for index, columnInfo := range m.ColumnList {
		columnInfo.Index = index + 1
		metaType := GetMetaType(gormx.DBType(m.DBType), columnInfo.DataBaseType)
		columnInfo.GoColumnType = metaType.GoType
		columnInfo.GureguNullableType = metaType.GureguNullableType
		columnInfo.SQLNullableType = metaType.SQLNullableType

		columnInfo.GoColumnName = UnderscoreToUpperCamelCase(columnInfo.ColumnName)
		columnInfo.TitleTableName = m.ModelStructName
		columnInfo.GoUpperColumnName = ToUpper(columnInfo.ColumnName)
		columnInfo.UpperTableName = ToUpper(m.ModelStructName)
		columnInfo.ModelPackageName = m.ModelPackageName
		columnInfo.ModelStructName = m.ModelStructName
		columnInfo.ModelShortName = ToLower(strings.Split(m.ModelStructName, "")[0])
		columnInfo.TableName = m.TableName
		columnInfo.SchemaName = m.SchemaName

		if columnInfo.Nullable {
			columnInfo.GoColumnType = func() string {
				if useSQLNullable {
					return columnInfo.SQLNullableType
				}
				return columnInfo.GureguNullableType
			}()
		}
		columnInfo.ValueFormat = metaType.ValueFormat
		gormTag := fmt.Sprintf("gorm:\"%s%s%s%s%s%s\"",
			func() string {
				var tag string
				if columnInfo.IsPrimaryKey {
					tag = tag + "primaryKey;"
				}
				if columnInfo.Unique {
					tag = tag + "unique;"
				}
				return tag
			}(),
			func() string {
				if columnInfo.AutoIncrement {
					return "autoIncrement;"
				}
				return ""
			}(),
			func() string {
				var tag = fmt.Sprintf("column:%s;", columnInfo.ColumnName)
				if !columnInfo.Nullable {
					tag = tag + "not null;"
				}
				return tag
			}(),
			func() string {
				return fmt.Sprintf("type:%s;", columnInfo.DataBaseType)
			}(),
			func() string {
				if columnInfo.Length != 0 {
					return fmt.Sprintf("size:%d;", columnInfo.Length)
				}
				return ""
			}(),
			func() string {
				if strings.Contains(columnInfo.DefaultValue, "::") {
					return fmt.Sprintf("default:%s;", columnInfo.DefaultValue[:strings.Index(columnInfo.DefaultValue, "::")])
				}
				if columnInfo.DefaultValue != "" {
					return fmt.Sprintf("default:%s;", columnInfo.DefaultValue)
				}
				return ""
			}())

		jsonTag := fmt.Sprintf("json:\"%s\"", func() string {
			switch m.JsonFormat {
			case "snake":
				return CamelCaseToUnderscore(columnInfo.ColumnName)
			case "upper_camel":
				return UnderscoreToUpperCamelCase(columnInfo.ColumnName)
			case "lower_camel":
				return UnderscoreToLowerCamelCase(columnInfo.ColumnName)
			default:
				return CamelCaseToUnderscore(columnInfo.ColumnName)
			}
		}())

		columnInfo.Tags = fmt.Sprintf("%s %s", gormTag, jsonTag)
	}
	result := map[string]any{
		"DBType":           m.DBType,
		"ModelPackageName": m.ModelPackageName,
		"ModelStructName":  m.ModelStructName,
		"ModelShortName":   ToLower(strings.Split(m.ModelStructName, "")[0]),
		"ColumnList":       m.ColumnList,
		"SchemaName":       m.SchemaName,
		"TableName":        m.TableName,
		"TitleTableName":   m.ModelStructName,
		"ImportPkgList":    []string{},
	}
	return result
}

// Model used as a variable because it cannot load template file after packed, params still can pass file
const Model = NotEditMark + `
package {{.ModelPackageName}}

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/guregu/null"
	"github.com/satori/go.uuid"
	{{range .ImportPkgList}}{{.}} ` + "\n" + `{{end}}
)

var (
	_ = time.Second
	_ = sql.LevelDefault
	_ = null.Bool{}
	_ = uuid.UUID{}
)

{{if .TableName -}}
	{{if and  .SchemaName (eq .DBType "postgres") -}}
const TableName{{.ModelStructName}} = "\"{{.SchemaName}}\".\"{{.TableName}}\""
	{{- else}}
const TableName{{.ModelStructName}} = "\"{{.TableName}}\""	
	{{- end}}
{{- end}}

type {{.TitleTableName}}Field string

// {{.ModelStructName}} struct is mapping to the {{.TableName}} table
type {{.ModelStructName}} struct {
    {{range .ColumnList}}
 
    {{.GoColumnName}} {{.GoColumnType}} ` + "`{{.Tags}}` " +
	"// Comment: {{if .Comment}}{{.Comment}}{{else}}no comment{{end}} " +
	`{{end}}
}

type {{.ModelStructName}}TableColumn struct {
	{{range .ColumnList}}
	{{- .GoColumnName}} {{.TitleTableName}}Field
	{{end}}
}

func ({{.ModelShortName}} *{{.ModelStructName}}) TableName() string {
	return "{{.TableName}}"
}

func ({{.ModelShortName}} *{{.ModelStructName}}) GetColumnInfo() {{.ModelStructName}}TableColumn {
	return {{.ModelStructName}}TableColumn{
		{{range .ColumnList}}
		{{- .GoColumnName}}: "{{.ColumnName}}",
		{{end}}		
	}
}

type {{.ModelStructName}}Condition struct {
	Condition
}

func ({{.ModelShortName}} *{{.ModelStructName}}Condition) AddStrCondition(condition string, args ...any) *{{.ModelStructName}}Condition {
	if len(args)> 0 {
		{{.ModelShortName}}.StringCondition = append({{.ModelShortName}}.StringCondition, fmt.Sprintf(condition, args))
		return {{.ModelShortName}}
	}
	{{.ModelShortName}}.StringCondition = append({{.ModelShortName}}.StringCondition, condition)
	return {{.ModelShortName}}
}

{{range .ColumnList}}
{{if eq .GoColumnType "string"}}
func ({{.ModelShortName}} *{{.ModelStructName}}Condition) {{.GoColumnName}}IsLike(value string) *{{.ModelStructName}}Condition {
	return {{.ModelShortName}}.AddStrCondition("\"{{.ColumnName}}\" like %q", value)
}
{{end}}
func ({{.ModelShortName}} *{{.ModelStructName}}Condition) {{.GoColumnName}}IsNull() *{{.ModelStructName}}Condition {
	return {{.ModelShortName}}.AddStrCondition("\"{{.ColumnName}}\" is null")
}

func ({{.ModelShortName}} *{{.ModelStructName}}Condition) {{.GoColumnName}}IsNotNull() *{{.ModelStructName}}Condition {
	return {{.ModelShortName}}.AddStrCondition("\"{{.ColumnName}}\" is not null")
}

func ({{.ModelShortName}} *{{.ModelStructName}}Condition) {{.GoColumnName}}EqualTo(value {{.GoColumnType}}) *{{.ModelStructName}}Condition {
	return {{.ModelShortName}}.AddStrCondition("\"{{.ColumnName}}\" = {{.ValueFormat}}", value)
}

func ({{.ModelShortName}} *{{.ModelStructName}}Condition) {{.GoColumnName}}NotEqualTo(value {{.GoColumnType}}) *{{.ModelStructName}}Condition {
	return {{.ModelShortName}}.AddStrCondition("\"{{.ColumnName}}\" <> {{.ValueFormat}}", value)
}

func ({{.ModelShortName}} *{{.ModelStructName}}Condition) {{.GoColumnName}}GreaterThan(value {{.GoColumnType}}) *{{.ModelStructName}}Condition {
	return {{.ModelShortName}}.AddStrCondition("\"{{.ColumnName}}\" > {{.ValueFormat}}", value)
}

func ({{.ModelShortName}} *{{.ModelStructName}}Condition) {{.GoColumnName}}GreaterThanOrEqualTo(value {{.GoColumnType}}) *{{.ModelStructName}}Condition {
	return {{.ModelShortName}}.AddStrCondition("\"{{.ColumnName}}\" >= {{.ValueFormat}}", value)
}

func ({{.ModelShortName}} *{{.ModelStructName}}Condition) {{.GoColumnName}}LessThan(value {{.GoColumnType}}) *{{.ModelStructName}}Condition {
	return {{.ModelShortName}}.AddStrCondition("\"{{.ColumnName}}\" < {{.ValueFormat}}", value)
}

func ({{.ModelShortName}} *{{.ModelStructName}}Condition) {{.GoColumnName}}LessThanOrEqualTo(value {{.GoColumnType}}) *{{.ModelStructName}}Condition {
	return {{.ModelShortName}}.AddStrCondition("\"{{.ColumnName}}\" <= {{.ValueFormat}}", value)
}

func ({{.ModelShortName}} *{{.ModelStructName}}Condition) {{.GoColumnName}}Between(startValue, endValue  {{.GoColumnType}}) *{{.ModelStructName}}Condition {
	return {{.ModelShortName}}.AddStrCondition("\"{{.ColumnName}}\" between {{.ValueFormat}} and {{.ValueFormat}}", startValue, endValue)
}

func ({{.ModelShortName}} *{{.ModelStructName}}Condition) {{.GoColumnName}}NotBetween(startValue, endValue  {{.GoColumnType}}) *{{.ModelStructName}}Condition {
	return {{.ModelShortName}}.AddStrCondition("\"{{.ColumnName}}\" not between {{.ValueFormat}} and {{.ValueFormat}}", startValue, endValue)
}

func ({{.ModelShortName}} *{{.ModelStructName}}Condition) {{.GoColumnName}}In(inValues []{{.GoColumnType}}) *{{.ModelStructName}}Condition {
	if len(inValues) == 0 {
		return {{.ModelShortName}}.AddStrCondition("\"{{.ColumnName}}\" in ()")
	}
	bytes, _ := json.Marshal(inValues)
	return {{.ModelShortName}}.AddStrCondition("\"{{.ColumnName}}\" in %s", "(" + strings.TrimRight(strings.TrimLeft(string(bytes), "["), "]") + ")")
}

func ({{.ModelShortName}} *{{.ModelStructName}}Condition) {{.GoColumnName}}NotIn(inValues []{{.GoColumnType}}) *{{.ModelStructName}}Condition {
	if len(inValues) == 0 {
		return {{.ModelShortName}}.AddStrCondition("\"{{.ColumnName}}\" not in ()")
	}
	bytes, _ := json.Marshal(inValues)
	return {{.ModelShortName}}.AddStrCondition("\"{{.ColumnName}}\" not in %s", "(" + strings.TrimRight(strings.TrimLeft(string(bytes), "["), "]") + ")")
}
{{end}}	

func ({{.ModelShortName}} *{{.ModelStructName}}Condition) AddMapCondition(mapCondition map[string]any) *{{.ModelStructName}}Condition {
	if len( {{.ModelShortName}}.MapCondition) == 0 {
		{{.ModelShortName}}.MapCondition = mapCondition
	} else  {
		for key, val := range {{.ModelShortName}}.MapCondition {
			{{.ModelShortName}}.MapCondition[key] = val
		}
	}
	return {{.ModelShortName}}
}

func ({{.ModelShortName}} *{{.ModelStructName}}Condition) AddOrderByClause(orderByClause ...string) *{{.ModelStructName}}Condition {
	{{.ModelShortName}}.OrderByClause = append({{.ModelShortName}}.OrderByClause, orderByClause...)
	return {{.ModelShortName}}
}

func ({{.ModelShortName}} *{{.ModelStructName}}Condition) Build() *Condition {
	return &{{.ModelShortName}}.Condition
}

`

// ModelHook hook file (no overwrite if file is existed), provide func BeforeCreate、AfterUpdate、BeforeDelete etc.
const ModelHook = EditMark + `
package {{.ModelPackageName}}

import (
	"gorm.io/gorm"
)

// BeforeSave invoked before saving, return an error.
func ({{.ModelShortName}} *{{.ModelStructName}}) BeforeSave() error {
	// TODO: something
	return nil
}

// AfterSave invoked after saving, return an error if field is not populated.
func ({{.ModelShortName}} *{{.ModelStructName}}) AfterSave() error {
	// TODO: something
	return nil
}

// BeforeCreate invoked before create, return an error.
func ({{.ModelShortName}} *{{.ModelStructName}}) BeforeCreate(tx *gorm.DB) (err error) {
	// TODO: something
	return
}

// AfterCreate invoked after create, return an error.
func ({{.ModelShortName}} *{{.ModelStructName}}) AfterCreate(tx *gorm.DB) (err error) {
	// TODO: something
	return
}

// BeforeUpdate invoked before update, return an error.
func ({{.ModelShortName}} *{{.ModelStructName}}) BeforeUpdate() error {
	// TODO: something
	return nil
}

// AfterUpdate invoked after update, return an error.
func ({{.ModelShortName}} *{{.ModelStructName}}) AfterUpdate(tx *gorm.DB) (err error) {
	// TODO: something
	return
}

// BeforeDelete invoked before delete, return an error.
func ({{.ModelShortName}} *{{.ModelStructName}}) BeforeDelete() error {
	// TODO: something
	return nil
}

// AfterDelete invoked after delete, return an error.
func ({{.ModelShortName}} *{{.ModelStructName}}) AfterDelete(tx *gorm.DB) (err error) {
	// TODO: something
	return
}

// AfterFind invoked after find, return an error.
func ({{.ModelShortName}} *{{.ModelStructName}}) AfterFind(tx *gorm.DB) (err error) {
	// TODO: something
	return
}

`

// ModelBase base file (no overwrite if file is existed)
const ModelBase = NotEditMark + `
package {{.ModelPackageName}}

import "math"

type ConditionBuilder interface {
	Build() *Condition
}

type Condition struct {
	MapCondition    map[string]any
	StringCondition []string
	OrderByClause []string
}

type UpdateField map[string]any

// Pagination 分页结构体（该分页只适合数据量很少的情况）
type Pagination struct {
	Page      int64 ` + "`json:\"page\"`       // 当前页\n" +
	"PageSize  int64 " + "`json:\"page_size\"`  // 每页多少条记录\n" +
	"PageCount int64 " + "`json:\"page_count\"` // 一共多少页\n" +
	"Total     int64 " + "`json:\"total\"`      // 一共多少条记录" + `
}

func (p *Pagination) CalculatePageCount() {
	if p.Page == 0 || p.PageSize == 0 {
		panic("error pagination param")
	}
	p.PageCount = int64(math.Ceil(float64(p.Total) / float64(p.PageSize)))
	return
}

func (p *Pagination) CalculateOffset() (offset int64) {
	if p.Page == 0 || p.PageSize == 0 {
		panic("error pagination param")
	}
	offset = (p.Page - 1) * p.PageSize
	return
}

`
