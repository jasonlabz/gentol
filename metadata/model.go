package metadata

import "github.com/jasonlabz/gentol/gormx"

type ModelMeta struct {
	BaseConfig
	ModelPackageName string
	ModelStructName  string
	ImportPkgList    []string
	ColumnList       []*ColumnInfo
}

type ColumnInfo struct {
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
	for _, columnInfo := range m.ColumnList {
		metaType := GetMetaType(gormx.DBType(m.DBType), columnInfo.DataBaseType)
		columnInfo.GoColumnType = metaType.GoType
		columnInfo.GureguNullableType = metaType.GureguNullableType
		columnInfo.SQLNullableType = metaType.SQLNullableType
	}
	result := map[string]any{
		"ModelPackageName": m.ModelPackageName,
		"ModelStructName":  m.ModelStructName,
		"ColumnList":       m.ColumnList,
		"SchemaName":       m.SchemaName,
		"TableName":        m.TableName,
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
    {{range .FieldList}}
    {{if .MultilineComment -}}
	/*
{{.ColumnComment}}
    */
	{{end -}}
    {{.GoColumnName}} {{.GoColumnType}} ` + "`{{.Tags}}` " +
	"{{if not .MultilineComment}}{{if .ColumnComment}}// {{.ColumnComment}}{{end}}{{end}}" +
	`{{end}}
}

`
