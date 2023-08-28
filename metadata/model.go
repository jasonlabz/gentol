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
	Tags               string
	ColumnName         string
	SQLNullableType    string
	GureguNullableType string
	ColumnType         string // varchar(64)
	DataBaseType       string // varchar
	Length             int64  // 64
	IsPrimaryKey       bool
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

		if columnInfo.Nullable {
			columnInfo.GoColumnType = func() string {
				if useSQLNullable {
					return columnInfo.SQLNullableType
				}
				return columnInfo.GureguNullableType
			}()
		}
		gormTag := fmt.Sprintf("gorm:\"%s%s%s%s%s%s\"",
			func() string {
				if columnInfo.IsPrimaryKey {
					return "primary_key;"
				}
				return ""
			}(),
			func() string {
				if columnInfo.AutoIncrement {
					return "auto_increment;"
				}
				return ""
			}(),
			func() string {
				return fmt.Sprintf("column:%s;", columnInfo.ColumnName)
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
	{{if .SchemaName -}}
const TableName{{.ModelStructName}} = "{{.SchemaName}}.{{.TableName}}"
	{{- else}}
const TableName{{.ModelStructName}} = "{{.TableName}}"	
	{{- end}}
{{- end}}

type {{.TitleTableName}}Field string

const (
	{{range .ColumnList}}
	{{.UpperTableName}}_{{.GoUpperColumnName}}  {{.TitleTableName}}Field = "{{.ColumnName}}"
	{{end}}
)

// {{.ModelStructName}} struct is mapping to the {{.TableName}} table
type {{.ModelStructName}} struct {
    {{range .ColumnList}}
 
    {{.GoColumnName}} {{.GoColumnType}} ` + "`{{.Tags}}` " +
	"// Comment: {{if .Comment}}{{.Comment}}{{else}}no comment{{end}} " +
	`{{end}}
}

`

// ModelHook hook file (no overwrite if file is existed), provide func BeforeCreate、AfterUpdate、BeforeDelete etc.
const ModelHook = NotEditMark + `
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

type Condition struct {
	MapCondition    map[string]any
	StringCondition []string
}

type UpdateField map[string]any

type OrderByClause []string

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

func init() {
	StoreTpl("model", Model)
	StoreTpl("model_base", ModelBase)
	StoreTpl("model_hook", ModelHook)
}
