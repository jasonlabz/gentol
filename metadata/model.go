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
	Index              int
	UpperTableName     string
	TitleTableName     string
	GoColumnName       string
	GoUpperColumnName  string
	GoColumnType       string // string
	GoColumnOriginType string // string
	ValueFormat        string // string
	Tags               string
	ModelPackageName   string
	ModelStructName    string
	ModelShortName     string
	SchemaName         string
	TableQuota         bool
	TableName          string

	ColumnQuota        bool
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
		columnInfo.GoColumnOriginType = metaType.GoType
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

		columnInfo.ColumnQuota = func() bool {
			if m.DBType == string(gormx.DBTypePostgres) ||
				m.DBType == string(gormx.DBTypeGreenplum) ||
				m.DBType == string(gormx.DBTypeDM) {
				if ToLower(columnInfo.ColumnName) == columnInfo.ColumnName {
					return false
				}
				return true
			}
			return false
		}()
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
				if strings.Contains(columnInfo.DefaultValue, "\"") {
					return ""
				}
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
		"SchemaQuota": func() bool {
			if m.DBType == string(gormx.DBTypePostgres) ||
				m.DBType == string(gormx.DBTypeGreenplum) ||
				m.DBType == string(gormx.DBTypeDM) {
				if ToLower(m.SchemaName) == m.SchemaName {
					return false
				}
				return true
			}
			return false
		}(),
		"TableQuota": func() bool {
			if m.DBType == string(gormx.DBTypePostgres) ||
				m.DBType == string(gormx.DBTypeGreenplum) ||
				m.DBType == string(gormx.DBTypeDM) {
				if ToLower(m.TableName) == m.TableName {
					return false
				}
				return true
			}
			return false
		}(),
	}
	return result
}

// Model used as a variable because it cannot load template file after packed, params still can pass file
const Model = NotEditMark + `
package {{.ModelPackageName}}

import (
	"database/sql"
	"time"

	"github.com/jasonlabz/null"
	"github.com/satori/go.uuid"
	{{range .ImportPkgList}}{{.}} ` + "\n" + `{{end}}
)

var (
	_ = time.Second
	_ = sql.LevelDefault
	_ = null.Bool{}
	_ = uuid.UUID{}
)

type {{.TitleTableName}}Field string

// {{.ModelStructName}} struct is mapping to the {{.TableName}} table
type {{.ModelStructName}} struct {
    {{range .ColumnList}}
 
    {{if eq .GoColumnName "TableName" }}{{.GoColumnName}}_{{ else }}{{.GoColumnName}}{{ end }} {{.GoColumnType}} ` + "`{{.Tags}}` " +
	"// Comment: {{if .Comment}}{{.Comment}}{{else}}no comment{{end}} " +
	`{{end}}
}

func ({{.ModelShortName}} *{{.ModelStructName}}) TableName() string {
{{if .TableName -}}
	{{if eq .DBType "postgres" -}}
		{{if and .SchemaName (ne .SchemaName "public") -}}
	return "{{if .SchemaQuota -}}\"{{.SchemaName}}\"{{- else}}{{.SchemaName}}{{- end}}.{{if .TableQuota -}}\"{{.TableName}}\"{{- else}}{{.TableName}}{{- end}}"
		{{- else -}}
	return "{{if .TableQuota -}}\"{{.TableName}}\"{{- else}}{{.TableName}}{{- end}}"
		{{ end }}
	{{- else if eq .DBType "oracle" -}}
 		{{if .SchemaName -}}
	return "{{.SchemaName}}.{{.TableName}}"
		{{- else -}}
	return "{{.TableName}}"
		{{ end }}
 	{{- else if eq .DBType "dm" -}}
 		{{if .SchemaName -}}
	return "{{if .SchemaQuota -}}\"{{.SchemaName}}\"{{- else}}{{.SchemaName}}{{- end}}.{{if .TableQuota -}}\"{{.TableName}}\"{{- else}}{{.TableName}}{{- end}}"
		{{- else -}}
	return "{{if .TableQuota -}}\"{{.TableName}}\"{{- else}}{{.TableName}}{{- end}}"
		{{ end }}
 	{{- else if eq .DBType "sqlserver" -}}
 		{{if ne .SchemaName "dbo" -}}
	return "{{.SchemaName}}.{{.TableName}}"
		{{- else -}}
	return "{{.TableName}}"
		{{end}}
	{{- else}}
	return "{{.TableName}}"	
	{{- end}}
{{- end}}
}

type {{.ModelStructName}}TableColumn struct {
	{{range .ColumnList}}
	{{- .GoColumnName}} {{.TitleTableName}}Field
	{{end}}
}

type {{.ModelStructName}}Condition struct {
	Condition
}

func ({{.ModelShortName}} *{{.ModelStructName}}Condition) ColumnInfo() {{.ModelStructName}}TableColumn {
	return {{.ModelStructName}}TableColumn{
		{{range .ColumnList}}
		{{- .GoColumnName}}: "{{.ColumnName}}",
		{{end}}		
	}
}

{{range .ColumnList}}
{{if eq .GoColumnOriginType "string"}}
func ({{.ModelShortName}} *{{.ModelStructName}}Condition) {{.GoColumnName}}PrefixLike(value string) *{{.ModelStructName}}Condition {
	return {{.ModelShortName}}.Where("{{if .ColumnQuota -}}\"{{.ColumnName}}\"{{- else}}{{.ColumnName}}{{- end}} like ?", value+"%")
}

func ({{.ModelShortName}} *{{.ModelStructName}}Condition) {{.GoColumnName}}IsLike(value string) *{{.ModelStructName}}Condition {
	return {{.ModelShortName}}.Where("{{if .ColumnQuota -}}\"{{.ColumnName}}\"{{- else}}{{.ColumnName}}{{- end}} like ?", "%"+value+"%")
}
{{end}}
func ({{.ModelShortName}} *{{.ModelStructName}}Condition) {{.GoColumnName}}IsNull() *{{.ModelStructName}}Condition {
	return {{.ModelShortName}}.Where("{{if .ColumnQuota -}}\"{{.ColumnName}}\"{{- else}}{{.ColumnName}}{{- end}} is null")
}

func ({{.ModelShortName}} *{{.ModelStructName}}Condition) {{.GoColumnName}}IsNotNull() *{{.ModelStructName}}Condition {
	return {{.ModelShortName}}.Where("{{if .ColumnQuota -}}\"{{.ColumnName}}\"{{- else}}{{.ColumnName}}{{- end}} is not null")
}

func ({{.ModelShortName}} *{{.ModelStructName}}Condition) {{.GoColumnName}}EqualTo(value {{.GoColumnOriginType}}) *{{.ModelStructName}}Condition {
	return {{.ModelShortName}}.Where("{{if .ColumnQuota -}}\"{{.ColumnName}}\"{{- else}}{{.ColumnName}}{{- end}} = ?", value)
}

func ({{.ModelShortName}} *{{.ModelStructName}}Condition) {{.GoColumnName}}NotEqualTo(value {{.GoColumnOriginType}}) *{{.ModelStructName}}Condition {
	return {{.ModelShortName}}.Where("{{if .ColumnQuota -}}\"{{.ColumnName}}\"{{- else}}{{.ColumnName}}{{- end}} <> ?", value)
}

func ({{.ModelShortName}} *{{.ModelStructName}}Condition) {{.GoColumnName}}GreaterThan(value {{.GoColumnOriginType}}) *{{.ModelStructName}}Condition {
	return {{.ModelShortName}}.Where("{{if .ColumnQuota -}}\"{{.ColumnName}}\"{{- else}}{{.ColumnName}}{{- end}} > ?", value)
}

func ({{.ModelShortName}} *{{.ModelStructName}}Condition) {{.GoColumnName}}GreaterThanOrEqualTo(value {{.GoColumnOriginType}}) *{{.ModelStructName}}Condition {
	return {{.ModelShortName}}.Where("{{if .ColumnQuota -}}\"{{.ColumnName}}\"{{- else}}{{.ColumnName}}{{- end}} >= ?", value)
}

func ({{.ModelShortName}} *{{.ModelStructName}}Condition) {{.GoColumnName}}LessThan(value {{.GoColumnOriginType}}) *{{.ModelStructName}}Condition {
	return {{.ModelShortName}}.Where("{{if .ColumnQuota -}}\"{{.ColumnName}}\"{{- else}}{{.ColumnName}}{{- end}} < ?", value)
}

func ({{.ModelShortName}} *{{.ModelStructName}}Condition) {{.GoColumnName}}LessThanOrEqualTo(value {{.GoColumnOriginType}}) *{{.ModelStructName}}Condition {
	return {{.ModelShortName}}.Where("{{if .ColumnQuota -}}\"{{.ColumnName}}\"{{- else}}{{.ColumnName}}{{- end}} <= ?", value)
}

func ({{.ModelShortName}} *{{.ModelStructName}}Condition) {{.GoColumnName}}Between(startValue, endValue  {{.GoColumnOriginType}}) *{{.ModelStructName}}Condition {
	return {{.ModelShortName}}.Where("{{if .ColumnQuota -}}\"{{.ColumnName}}\"{{- else}}{{.ColumnName}}{{- end}} between ? and ?", startValue, endValue)
}

func ({{.ModelShortName}} *{{.ModelStructName}}Condition) {{.GoColumnName}}NotBetween(startValue, endValue  {{.GoColumnOriginType}}) *{{.ModelStructName}}Condition {
	return {{.ModelShortName}}.Where("{{if .ColumnQuota -}}\"{{.ColumnName}}\"{{- else}}{{.ColumnName}}{{- end}} not between ? and ?", startValue, endValue)
}

func ({{.ModelShortName}} *{{.ModelStructName}}Condition) {{.GoColumnName}}In(inValues []{{.GoColumnOriginType}}) *{{.ModelStructName}}Condition {
	return {{.ModelShortName}}.Where(TransInCondition("{{if .ColumnQuota -}}\"{{.ColumnName}}\"{{- else}}{{.ColumnName}}{{- end}} in (?)", inValues))
}

func ({{.ModelShortName}} *{{.ModelStructName}}Condition) {{.GoColumnName}}NotIn(inValues []{{.GoColumnOriginType}}) *{{.ModelStructName}}Condition {
	return {{.ModelShortName}}.Where(TransInCondition("{{if .ColumnQuota -}}\"{{.ColumnName}}\"{{- else}}{{.ColumnName}}{{- end}} not in (?)", inValues))
}
{{end}}

func ({{.ModelShortName}} *{{.ModelStructName}}Condition) Where(query any, args ...any) *{{.ModelStructName}}Condition {
	switch v := query.(type) {
	case string:
		if len(args) > 0 {
			{{.ModelShortName}}.StringCondition = append({{.ModelShortName}}.StringCondition, v)
			{{.ModelShortName}}.Condition.Args = append({{.ModelShortName}}.Condition.Args, args...)
		} else {
			{{.ModelShortName}}.StringCondition = append({{.ModelShortName}}.StringCondition, v)
		}
	case map[string]any:
		if {{.ModelShortName}}.MapCondition == nil {
			{{.ModelShortName}}.MapCondition = v
		} else {
			for key, val := range v {
				{{.ModelShortName}}.MapCondition[key] = val
			}
		}
	}
	return {{.ModelShortName}}
}

func ({{.ModelShortName}} *{{.ModelStructName}}Condition) OrderBy(orderByClause ...string) *{{.ModelStructName}}Condition {
	{{.ModelShortName}}.OrderByClause = append({{.ModelShortName}}.OrderByClause, orderByClause...)
	return {{.ModelShortName}}
}

func ({{.ModelShortName}} *{{.ModelStructName}}Condition) GroupBy(groupByClause string) *{{.ModelStructName}}Condition {
	{{.ModelShortName}}.GroupByClause = groupByClause
	return {{.ModelShortName}}
}

func ({{.ModelShortName}} *{{.ModelStructName}}Condition) Having(query string, args ...any) *{{.ModelStructName}}Condition {
	{{.ModelShortName}}.HavingCondition = query
	{{.ModelShortName}}.HavingArgs = args
	return {{.ModelShortName}}
}

func ({{.ModelShortName}} *{{.ModelStructName}}Condition) Joins(query string, args ...any) *{{.ModelStructName}}Condition {
	{{.ModelShortName}}.JoinCondition = append({{.ModelShortName}}.JoinCondition, fmt.Sprintf(query, args...))
	{{.ModelShortName}}.JoinCondition = append({{.ModelShortName}}.JoinCondition, query)
	{{.ModelShortName}}.JoinArgs = append({{.ModelShortName}}.JoinArgs, args...)
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
func ({{.ModelShortName}} *{{.ModelStructName}}) BeforeSave(tx *gorm.DB) (err error) {
	// TODO: something
	return 
}

// AfterSave invoked after saving, return an error if field is not populated.
func ({{.ModelShortName}} *{{.ModelStructName}}) AfterSave(tx *gorm.DB) (err error) {
	// TODO: something
	return 
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
func ({{.ModelShortName}} *{{.ModelStructName}}) BeforeUpdate(tx *gorm.DB) (err error) {
	// TODO: something
	return 
}

// AfterUpdate invoked after update, return an error.
func ({{.ModelShortName}} *{{.ModelStructName}}) AfterUpdate(tx *gorm.DB) (err error) {
	// TODO: something
	return
}

// BeforeDelete invoked before delete, return an error.
func ({{.ModelShortName}} *{{.ModelStructName}}) BeforeDelete(tx *gorm.DB) (err error) {
	// TODO: something
	return 
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

import (
	"fmt"
	"math"
	"strings"
)

type ConditionBuilder interface {
	Build() *Condition
}

type Condition struct {
	JoinCondition   []string
	JoinArgs        []any
	MapCondition    map[string]any
	StringCondition []string
	Args            []any
	GroupByClause   string
	HavingCondition string
	HavingArgs      []any
	OrderByClause   []string
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
