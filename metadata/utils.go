package metadata

import (
	"fmt"
	"github.com/jasonlabz/gentol/gormx"
	"os"
	"strings"
	"unicode"
)

func GetMetaType(dbType gormx.DBType, columnType string) (metaType MetaType) {
	switch dbType {
	case gormx.DBTypeGreenplum:
		fallthrough
	case gormx.DBTypePostgres:
		metaType = PostgresTrans(columnType)
	case gormx.DBTypeMySQL:
		metaType = MySQLTrans(columnType)
	case gormx.DBTypeSqlserver:
		metaType = SQLServerTrans(columnType)
	case gormx.DBTypeOracle:
		metaType = OracleTrans(columnType)
	default:
		panic(fmt.Errorf("unsupported db_type: %s ", dbType))
	}
	return
}

func PostgresTrans(columnType string) (metaType MetaType) {
	columnType = strings.ToLower(columnType)
	switch columnType {
	case "bool", "boolean", "bit":
		metaType.GoType = "bool"
		metaType.SQLNullableType = "sql.NullBool"
		metaType.GureguNullableType = "null.Int"
	case "smallint", "int2":
		metaType.GoType = "int16"
		metaType.SQLNullableType = "sql.NullInt32"
		metaType.GureguNullableType = "null.Int"
	case "integer", "int", "int4", "serial":
		metaType.GoType = "int32"
		metaType.SQLNullableType = "sql.NullInt32"
		metaType.GureguNullableType = "null.Int"
	case "int8", "bigint", "bigserial":
		metaType.GoType = "int64"
		metaType.SQLNullableType = "sql.NullInt64"
		metaType.GureguNullableType = "null.Int"
	case "real", "float4":
		metaType.GoType = "float32"
		metaType.SQLNullableType = "sql.NullFloat32"
		metaType.GureguNullableType = "null.Float"
	case "double precision", "float8", "numeric", "money", "decimal":
		metaType.GoType = "float64"
		metaType.SQLNullableType = "sql.NullFloat64"
		metaType.GureguNullableType = "null.Float"
	case "bytea", "char", "varchar", "character", "text", "json", "xml", "jsonb":
		metaType.GoType = "string"
		metaType.SQLNullableType = "sql.NullString"
		metaType.GureguNullableType = "null.String"
	case "date", "time", "timetz", "timestamp", "timestamptz":
		metaType.GoType = "time.Time"
		metaType.GureguNullableType = "null.Time"
		metaType.SQLNullableType = "time.Time"
	default:
		if strings.Contains(columnType, "numeric") ||
			strings.Contains(columnType, "decimal") {
			metaType.GoType = "float64"
			metaType.SQLNullableType = "sql.NullFloat64"
			metaType.GureguNullableType = "null.Float"
		} else if strings.Contains(columnType, "character") {
			metaType.GoType = "string"
			metaType.SQLNullableType = "sql.NullString"
			metaType.GureguNullableType = "null.String"
		} else if strings.Contains(columnType, "timestamp") ||
			strings.Contains(columnType, "time") {
			metaType.GoType = "time.Time"
			metaType.GureguNullableType = "null.Time"
			metaType.SQLNullableType = "time.Time"
		} else {
			fmt.Printf("unknow column type : %s, replace it with \"string\"", columnType)
			metaType.GoType = "string"
			metaType.SQLNullableType = "sql.NullString"
			metaType.GureguNullableType = "null.String"
		}
	}
	return
}

func MySQLTrans(columnType string) (metaType MetaType) {
	columnType = strings.ToLower(columnType)
	switch columnType {
	case "bool", "boolean", "bit":
		metaType.GoType = "bool"
		metaType.SQLNullableType = "sql.NullBool"
		metaType.GureguNullableType = "null.Int"
	case "int1":
		metaType.GoType = "int8"
		metaType.SQLNullableType = "sql.NullInt32"
		metaType.GureguNullableType = "null.Int"
	case "smallint", "int2", "tinyint":
		metaType.GoType = "int16"
		metaType.SQLNullableType = "sql.NullInt32"
		metaType.GureguNullableType = "null.Int"
	case "mediumint", "middleint", "serial", "integer", "int", "int3", "int4":
		metaType.GoType = "int32"
		metaType.SQLNullableType = "sql.NullInt32"
		metaType.GureguNullableType = "null.Int"
	case "int8", "bigint":
		metaType.GoType = "int64"
		metaType.SQLNullableType = "sql.NullInt64"
		metaType.GureguNullableType = "null.Int"
	case "float", "float4":
		metaType.GoType = "float32"
		metaType.SQLNullableType = "sql.NullFloat32"
		metaType.GureguNullableType = "null.Float"
	case "float8", "numeric", "double", "decimal":
		metaType.GoType = "float64"
		metaType.SQLNullableType = "sql.NullFloat64"
		metaType.GureguNullableType = "null.Float"
	case "set", "enum", "json", "binary", "varbinary", "tinytext", "mediumtext", "longtext",
		"char", "nchar", "varchar", "character", "text", "blob":
		metaType.GoType = "string"
		metaType.SQLNullableType = "sql.NullString"
		metaType.GureguNullableType = "null.String"
	case "year", "date", "time", "timestamp", "datetime":
		metaType.GoType = "time.Time"
		metaType.GureguNullableType = "null.Time"
		metaType.SQLNullableType = "time.Time"
	default:
		if strings.Contains(columnType, "numeric") ||
			strings.Contains(columnType, "decimal") {
			metaType.GoType = "float64"
			metaType.SQLNullableType = "sql.NullFloat64"
			metaType.GureguNullableType = "null.Float"
		} else if strings.Contains(columnType, "char") {
			metaType.GoType = "string"
			metaType.SQLNullableType = "sql.NullString"
			metaType.GureguNullableType = "null.String"
		} else if strings.Contains(columnType, "timestamp") ||
			strings.Contains(columnType, "time") {
			metaType.GoType = "time.Time"
			metaType.GureguNullableType = "null.Time"
			metaType.SQLNullableType = "time.Time"
		} else {
			fmt.Printf("unknow column type : %s, replace it with \"string\"", columnType)
			metaType.GoType = "string"
			metaType.SQLNullableType = "sql.NullString"
			metaType.GureguNullableType = "null.String"
		}
	}
	return
}

func SQLServerTrans(columnType string) (metaType MetaType) {
	columnType = strings.ToLower(columnType)
	switch columnType {
	case "bool", "boolean", "bit":
		metaType.GoType = "bool"
		metaType.SQLNullableType = "sql.NullBool"
		metaType.GureguNullableType = "null.Int"
	case "tinyint", "smallint":
		metaType.GoType = "int16"
		metaType.SQLNullableType = "sql.NullInt32"
		metaType.GureguNullableType = "null.Int"
	case "int":
		metaType.GoType = "int32"
		metaType.SQLNullableType = "sql.NullInt32"
		metaType.GureguNullableType = "null.Int"
	case "bigint":
		metaType.GoType = "int64"
		metaType.SQLNullableType = "sql.NullInt64"
		metaType.GureguNullableType = "null.Int"
	case "smallmoney":
		metaType.GoType = "float32"
		metaType.SQLNullableType = "sql.NullFloat32"
		metaType.GureguNullableType = "null.Float"
	case "money", "real", "float", "numeric", "decimal":
		metaType.GoType = "float64"
		metaType.SQLNullableType = "sql.NullFloat64"
		metaType.GureguNullableType = "null.Float"
	case "ntext", "text", "xml", "table", "char", "varchar", "nchar", "nvarchar", "varbinary", "binary", "image":
		metaType.GoType = "string"
		metaType.SQLNullableType = "sql.NullString"
		metaType.GureguNullableType = "null.String"
	case "datetime", "datetime2", "smalldatetime", "date", "time", "datetimeoffset", "timestamp":
		metaType.GoType = "time.Time"
		metaType.GureguNullableType = "null.Time"
		metaType.SQLNullableType = "time.Time"
	default:
		if strings.Contains(columnType, "numeric") ||
			strings.Contains(columnType, "decimal") ||
			strings.Contains(columnType, "money") {
			metaType.GoType = "float64"
			metaType.SQLNullableType = "sql.NullFloat64"
			metaType.GureguNullableType = "null.Float"
		} else {
			fmt.Printf("unknow column type : %s, replace it with \"string\"", columnType)
			metaType.GoType = "string"
			metaType.SQLNullableType = "sql.NullString"
			metaType.GureguNullableType = "null.String"
		}
	}
	return
}

func OracleTrans(columnType string) (metaType MetaType) {
	columnType = strings.ToLower(columnType)
	switch columnType {
	case "bool", "boolean", "bit":
		metaType.GoType = "bool"
		metaType.SQLNullableType = "sql.NullBool"
		metaType.GureguNullableType = "null.Int"
	case "smallint", "int2":
		metaType.GoType = "int16"
		metaType.SQLNullableType = "sql.NullInt32"
		metaType.GureguNullableType = "null.Int"
	case "integer":
		metaType.GoType = "int32"
		metaType.SQLNullableType = "sql.NullInt32"
		metaType.GureguNullableType = "null.Int"
	case "real", "binary float":
		metaType.GoType = "float32"
		metaType.SQLNullableType = "sql.NullFloat32"
		metaType.GureguNullableType = "null.Float"
	case "float", "binary double", "number", "decimal":
		metaType.GoType = "float64"
		metaType.SQLNullableType = "sql.NullFloat64"
		metaType.GureguNullableType = "null.Float"
	case "char", "long", "nchar", "varchar", "varchar2", "nvarchar2", "rowid", "nrowid",
		"clob", "nclob", "blob", "raw", "long raw", "bfile":
		metaType.GoType = "string"
		metaType.SQLNullableType = "sql.NullString"
		metaType.GureguNullableType = "null.String"
	case "date", "timestamp":
		metaType.GoType = "time.Time"
		metaType.GureguNullableType = "null.Time"
		metaType.SQLNullableType = "time.Time"
	default:
		if strings.Contains(columnType, "number") ||
			strings.Contains(columnType, "decimal") {
			metaType.GoType = "float64"
			metaType.SQLNullableType = "sql.NullFloat64"
			metaType.GureguNullableType = "null.Float"
		} else if strings.Contains(columnType, "character") {
			metaType.GoType = "string"
			metaType.SQLNullableType = "sql.NullString"
			metaType.GureguNullableType = "null.String"
		} else if strings.Contains(columnType, "timestamp") ||
			strings.Contains(columnType, "time") ||
			strings.Contains(columnType, "interval") {
			metaType.GoType = "time.Time"
			metaType.GureguNullableType = "null.Time"
			metaType.SQLNullableType = "time.Time"
		} else {
			fmt.Printf("unknow column type : %s, replace it with \"string\"", columnType)
			metaType.GoType = "string"
			metaType.SQLNullableType = "sql.NullString"
			metaType.GureguNullableType = "null.String"
		}
	}
	return
}

// ToUpper 单词全部转化为大写
func ToUpper(s string) string {
	return strings.ToUpper(s)
}

// ToLower 单词全部转化为小写
func ToLower(s string) string {
	return strings.ToLower(s)
}

// UnderscoreToUpperCamelCase 下划线单词转为大写驼峰单词
func UnderscoreToUpperCamelCase(s string) string {
	splitList := strings.Split(s, "_")
	for index, item := range splitList {
		_, ok := abbreviationMap[ToUpper(item)]
		if ok {
			splitList[index] = ToUpper(item)
		} else {
			splitList[index] = strings.Title(item)
		}
	}
	s = strings.Join(splitList, "")
	return s
}

// UnderscoreToLowerCamelCase 下划线单词转为小写驼峰单词
func UnderscoreToLowerCamelCase(s string) string {
	s = UnderscoreToUpperCamelCase(s)
	return string(unicode.ToLower(rune(s[0]))) + s[1:]
}

// CamelCaseToUnderscore 驼峰单词转下划线单词
func CamelCaseToUnderscore(s string) string {
	var output []rune
	var next int
	for i, r := range s {
		if i == 0 {
			output = append(output, unicode.ToLower(r))
		} else {
			if i > next && unicode.IsUpper(r) {
				next = i + 1
				output = append(output, '_')
			}

			output = append(output, unicode.ToLower(r))
		}
	}
	return string(output)
}

// ListDir 获取指定目录下文件
func ListDir(dirPth string, suffix string) (files []string, err error) {
	files = make([]string, 0)

	dir, err := os.ReadDir(dirPth)
	if err != nil {
		return nil, err
	}

	PthSep := string(os.PathSeparator)
	suffix = strings.ToUpper(suffix) //忽略后缀匹配的大小写

	for _, fi := range dir {
		if fi.IsDir() { // 忽略目录
			continue
		}
		if suffix == "" {
			files = append(files, dirPth+PthSep+fi.Name())
			continue
		}
		if strings.HasSuffix(strings.ToUpper(fi.Name()), suffix) { //匹配文件
			files = append(files, dirPth+PthSep+fi.Name())
		}
	}

	return files, nil
}
