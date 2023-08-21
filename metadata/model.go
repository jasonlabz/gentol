package metadata

import (
	"fmt"

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
	GoColumnName       string
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

		if columnInfo.Nullable {
			columnInfo.GoColumnType = func() string {
				if useSQLNullable {
					return columnInfo.SQLNullableType
				}
				return columnInfo.GureguNullableType
			}()
		}
		gormTag := fmt.Sprintf("gorm:\"%s%s%s%s%s\"", func() string {
			if columnInfo.IsPrimaryKey {
				return "primary_key;"
			}
			return ""
		}(), func() string {
			return fmt.Sprintf("column:%s;", columnInfo.ColumnName)
		}(), func() string {
			return fmt.Sprintf("type:%s;", columnInfo.DataBaseType)
		}(), func() string {
			if columnInfo.Length != 0 {
				return fmt.Sprintf("size:%d;", columnInfo.Length)
			}
			return ""
		}(), func() string {
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
		"ColumnList":       m.ColumnList,
		"SchemaName":       m.SchemaName,
		"TableName":        m.TableName,
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

// {{.ModelStructName}} struct is mapping to the {{.TableName}} table
type {{.ModelStructName}} struct {
    {{range .ColumnList}}
 
    {{.GoColumnName}} {{.GoColumnType}} ` + "`{{.Tags}}` " +
	"// Comment: {{if .Comment}}{{.Comment}}{{else}}no comment{{end}} " +
	`{{end}}
}

`

func init() {
	StoreTpl("model", Model)
}
