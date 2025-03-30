package metadata

import (
	"fmt"
	"github.com/jasonlabz/gentol/gormx"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"runtime"
	"strconv"
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
	case gormx.DBTypeDM:
		metaType = DmTrans(columnType)
	case gormx.DBTypeSQLite:
		metaType = SQLiteTrans(columnType)
	default:
		panic(fmt.Errorf("unsupported db_type: %s ", dbType))
	}
	return
}

func DmTrans(columnType string) (metaType MetaType) {
	columnType = strings.ToLower(columnType)
	switch columnType {
	case "bit", "bool", "boolean":
		metaType.GoType = "bool"
		metaType.SQLNullableType = "sql.NullBool"
		metaType.GureguNullableType = "null.Bool"
		metaType.ValueFormat = "%v"
	case "tinyint", "int1":
		metaType.GoType = "int8"
		metaType.SQLNullableType = "sql.NullInt32"
		metaType.GureguNullableType = "null.Int"
		metaType.ValueFormat = "%v"
	case "smallint", "int2":
		metaType.GoType = "int16"
		metaType.SQLNullableType = "sql.NullInt32"
		metaType.GureguNullableType = "null.Int"
		metaType.ValueFormat = "%v"
	case "int", "int4", "integer":
		metaType.GoType = "int32"
		metaType.SQLNullableType = "sql.NullInt32"
		metaType.GureguNullableType = "null.Int"
		metaType.ValueFormat = "%v"
	case "bigint", "int8":
		metaType.GoType = "int64"
		metaType.SQLNullableType = "sql.NullInt64"
		metaType.GureguNullableType = "null.Int"
		metaType.ValueFormat = "%v"
	case "real", "float", "float4":
		metaType.GoType = "float32"
		metaType.SQLNullableType = "sql.NullFloat32"
		metaType.GureguNullableType = "null.Float"
		metaType.ValueFormat = "%v"
	case "double", "float8", "number", "decimal", "numeric":
		metaType.GoType = "float64"
		metaType.SQLNullableType = "sql.NullFloat64"
		metaType.GureguNullableType = "null.Float"
		metaType.ValueFormat = "%v"
	case "char", "varchar", "varchar2", "text", "clob", "long":
		metaType.GoType = "string"
		metaType.SQLNullableType = "sql.NullString"
		metaType.GureguNullableType = "null.String"
		metaType.ValueFormat = "'%v'"
	case "date", "time", "timestamp", "datetime":
		metaType.GoType = "time.Time"
		metaType.SQLNullableType = "sql.NullTime"
		metaType.GureguNullableType = "null.Time"
		metaType.ValueFormat = "'%v'"
	case "blob", "raw", "byte", "binary":
		metaType.GoType = "[]byte"
		metaType.SQLNullableType = "sql.RawBytes"
		metaType.GureguNullableType = "null.Bytes"
		metaType.ValueFormat = "'%v'"
	default:
		if strings.Contains(columnType, "number") ||
			strings.Contains(columnType, "decimal") ||
			strings.Contains(columnType, "numeric") {
			metaType.GoType = "float64"
			metaType.SQLNullableType = "sql.NullFloat64"
			metaType.GureguNullableType = "null.Float"
			metaType.ValueFormat = "%v"
		} else if strings.Contains(columnType, "char") ||
			strings.Contains(columnType, "text") {
			metaType.GoType = "string"
			metaType.SQLNullableType = "sql.NullString"
			metaType.GureguNullableType = "null.String"
			metaType.ValueFormat = "'%v'"
		} else if strings.Contains(columnType, "time") ||
			strings.Contains(columnType, "date") {
			metaType.GoType = "time.Time"
			metaType.SQLNullableType = "sql.NullTime"
			metaType.GureguNullableType = "null.Time"
			metaType.ValueFormat = "'%v'"
		} else {
			fmt.Printf("unknown DM column type: %s, default to string\n", columnType)
			metaType.GoType = "string"
			metaType.SQLNullableType = "sql.NullString"
			metaType.GureguNullableType = "null.String"
			metaType.ValueFormat = "'%v'"
		}
	}
	return
}

func SQLiteTrans(columnType string) (metaType MetaType) {
	columnType = strings.ToLower(columnType)
	switch columnType {
	case "boolean", "bool":
		metaType.GoType = "bool"
		metaType.SQLNullableType = "sql.NullBool"
		metaType.GureguNullableType = "null.Bool"
		metaType.ValueFormat = "%v"
	case "tinyint", "int1":
		metaType.GoType = "int8"
		metaType.SQLNullableType = "sql.NullInt32"
		metaType.GureguNullableType = "null.Int"
		metaType.ValueFormat = "%v"
	case "smallint", "int2":
		metaType.GoType = "int16"
		metaType.SQLNullableType = "sql.NullInt32"
		metaType.GureguNullableType = "null.Int"
		metaType.ValueFormat = "%v"
	case "integer", "int", "int4":
		metaType.GoType = "int32"
		metaType.SQLNullableType = "sql.NullInt32"
		metaType.GureguNullableType = "null.Int"
		metaType.ValueFormat = "%v"
	case "bigint", "int8":
		metaType.GoType = "int64"
		metaType.SQLNullableType = "sql.NullInt64"
		metaType.GureguNullableType = "null.Int"
		metaType.ValueFormat = "%v"
	case "real", "float":
		metaType.GoType = "float64"
		metaType.SQLNullableType = "sql.NullFloat64"
		metaType.GureguNullableType = "null.Float"
		metaType.ValueFormat = "%v"
	case "text", "varchar", "nchar", "clob":
		metaType.GoType = "string"
		metaType.SQLNullableType = "sql.NullString"
		metaType.GureguNullableType = "null.String"
		metaType.ValueFormat = "'%v'"
	case "blob", "binary":
		metaType.GoType = "[]byte"
		metaType.SQLNullableType = "sql.RawBytes"
		metaType.GureguNullableType = "null.Bytes"
		metaType.ValueFormat = "'%v'"
	case "date", "datetime", "timestamp":
		metaType.GoType = "time.Time"
		metaType.SQLNullableType = "sql.NullTime"
		metaType.GureguNullableType = "null.Time"
		metaType.ValueFormat = "'%v'"
	default:
		if strings.Contains(columnType, "int") {
			metaType.GoType = "int64"
			metaType.SQLNullableType = "sql.NullInt64"
			metaType.GureguNullableType = "null.Int"
			metaType.ValueFormat = "%v"
		} else if strings.Contains(columnType, "real") ||
			strings.Contains(columnType, "float") ||
			strings.Contains(columnType, "double") {
			metaType.GoType = "float64"
			metaType.SQLNullableType = "sql.NullFloat64"
			metaType.GureguNullableType = "null.Float"
			metaType.ValueFormat = "%v"
		} else if strings.Contains(columnType, "char") ||
			strings.Contains(columnType, "text") {
			metaType.GoType = "string"
			metaType.SQLNullableType = "sql.NullString"
			metaType.GureguNullableType = "null.String"
			metaType.ValueFormat = "'%v'"
		} else {
			fmt.Printf("unknown SQLite column type: %s, default to string\n", columnType)
			metaType.GoType = "string"
			metaType.SQLNullableType = "sql.NullString"
			metaType.GureguNullableType = "null.String"
			metaType.ValueFormat = "'%v'"
		}
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
		metaType.ValueFormat = "%v"
	case "smallint", "int2":
		metaType.GoType = "int16"
		metaType.SQLNullableType = "sql.NullInt32"
		metaType.GureguNullableType = "null.Int"
		metaType.ValueFormat = "%v"
	case "integer", "int", "int4", "serial":
		metaType.GoType = "int32"
		metaType.SQLNullableType = "sql.NullInt32"
		metaType.GureguNullableType = "null.Int"
		metaType.ValueFormat = "%v"
	case "int8", "bigint", "bigserial":
		metaType.GoType = "int64"
		metaType.SQLNullableType = "sql.NullInt64"
		metaType.GureguNullableType = "null.Int"
		metaType.ValueFormat = "%v"
	case "real", "float4":
		metaType.GoType = "float32"
		metaType.SQLNullableType = "sql.NullFloat32"
		metaType.GureguNullableType = "null.Float"
		metaType.ValueFormat = "%v"
	case "double precision", "float8", "numeric", "money", "decimal":
		metaType.GoType = "float64"
		metaType.SQLNullableType = "sql.NullFloat64"
		metaType.GureguNullableType = "null.Float"
		metaType.ValueFormat = "%v"
	case "bytea":
		metaType.GoType = "[]byte"
		metaType.SQLNullableType = "sql.RawBytes"
		metaType.GureguNullableType = "null.Bytes"
		metaType.ValueFormat = "'%v'"
	case "char", "varchar", "character", "text", "json", "xml", "jsonb":
		metaType.GoType = "string"
		metaType.SQLNullableType = "sql.NullString"
		metaType.GureguNullableType = "null.String"
		metaType.ValueFormat = "'%v'"
	case "date", "time", "timetz", "timestamp", "timestamptz":
		metaType.GoType = "time.Time"
		metaType.GureguNullableType = "null.Time"
		metaType.SQLNullableType = "sql.NullTime"
		metaType.ValueFormat = "'%v'"
	default:
		if strings.Contains(columnType, "numeric") ||
			strings.Contains(columnType, "decimal") {
			metaType.GoType = "float64"
			metaType.SQLNullableType = "sql.NullFloat64"
			metaType.GureguNullableType = "null.Float"
			metaType.ValueFormat = "%v"
		} else if strings.Contains(columnType, "character") {
			metaType.GoType = "string"
			metaType.SQLNullableType = "sql.NullString"
			metaType.GureguNullableType = "null.String"
			metaType.ValueFormat = "'%v'"
		} else if strings.Contains(columnType, "timestamp") ||
			strings.Contains(columnType, "time") {
			metaType.GoType = "time.Time"
			metaType.GureguNullableType = "null.Time"
			metaType.SQLNullableType = "sql.NullTime"
			metaType.ValueFormat = "'%v'"
		} else {
			fmt.Printf("unknow column type : %s, replace it with \"string\"\n", columnType)
			metaType.GoType = "string"
			metaType.SQLNullableType = "sql.NullString"
			metaType.GureguNullableType = "null.String"
			metaType.ValueFormat = "'%v'"
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
		metaType.ValueFormat = "%v"
	case "int1":
		metaType.GoType = "int8"
		metaType.SQLNullableType = "sql.NullInt32"
		metaType.GureguNullableType = "null.Int"
		metaType.ValueFormat = "%v"
	case "smallint", "int2", "tinyint":
		metaType.GoType = "int16"
		metaType.SQLNullableType = "sql.NullInt32"
		metaType.GureguNullableType = "null.Int"
		metaType.ValueFormat = "%v"
	case "mediumint", "middleint", "serial", "integer", "int", "int3", "int4":
		metaType.GoType = "int32"
		metaType.SQLNullableType = "sql.NullInt32"
		metaType.GureguNullableType = "null.Int"
		metaType.ValueFormat = "%v"
	case "int8", "bigint":
		metaType.GoType = "int64"
		metaType.SQLNullableType = "sql.NullInt64"
		metaType.GureguNullableType = "null.Int"
		metaType.ValueFormat = "%v"
	case "float", "float4":
		metaType.GoType = "float32"
		metaType.SQLNullableType = "sql.NullFloat32"
		metaType.GureguNullableType = "null.Float"
		metaType.ValueFormat = "%v"
	case "float8", "numeric", "double", "decimal":
		metaType.GoType = "float64"
		metaType.SQLNullableType = "sql.NullFloat64"
		metaType.GureguNullableType = "null.Float"
		metaType.ValueFormat = "%v"
	case "set", "enum", "json", "tinytext", "mediumtext", "longtext",
		"char", "nchar", "varchar", "character", "text":
		metaType.GoType = "string"
		metaType.SQLNullableType = "sql.NullString"
		metaType.GureguNullableType = "null.String"
		metaType.ValueFormat = "'%v'"
	case "binary", "varbinary", "blob":
		metaType.GoType = "[]byte"
		metaType.SQLNullableType = "sql.RawBytes"
		metaType.GureguNullableType = "null.Bytes"
		metaType.ValueFormat = "'%v'"
	case "year", "date", "time", "timestamp", "datetime":
		metaType.GoType = "time.Time"
		metaType.GureguNullableType = "null.Time"
		metaType.SQLNullableType = "sql.NullTime"
		metaType.ValueFormat = "'%v'"
	default:
		if strings.Contains(columnType, "numeric") ||
			strings.Contains(columnType, "decimal") {
			metaType.GoType = "float64"
			metaType.SQLNullableType = "sql.NullFloat64"
			metaType.GureguNullableType = "null.Float"
			metaType.ValueFormat = "%v"
		} else if strings.Contains(columnType, "char") {
			metaType.GoType = "string"
			metaType.SQLNullableType = "sql.NullString"
			metaType.GureguNullableType = "null.String"
			metaType.ValueFormat = "'%v'"
		} else if strings.Contains(columnType, "timestamp") ||
			strings.Contains(columnType, "time") {
			metaType.GoType = "time.Time"
			metaType.GureguNullableType = "null.Time"
			metaType.SQLNullableType = "sql.NullTime"
			metaType.ValueFormat = "'%v'"
		} else {
			fmt.Printf("unknow column type : %s, replace it with \"string\"\n", columnType)
			metaType.GoType = "string"
			metaType.SQLNullableType = "sql.NullString"
			metaType.GureguNullableType = "null.String"
			metaType.ValueFormat = "'%v'"
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
		metaType.ValueFormat = "%v"
	case "tinyint", "smallint":
		metaType.GoType = "int16"
		metaType.SQLNullableType = "sql.NullInt32"
		metaType.GureguNullableType = "null.Int"
		metaType.ValueFormat = "%v"
	case "int":
		metaType.GoType = "int32"
		metaType.SQLNullableType = "sql.NullInt32"
		metaType.GureguNullableType = "null.Int"
		metaType.ValueFormat = "%v"
	case "bigint":
		metaType.GoType = "int64"
		metaType.SQLNullableType = "sql.NullInt64"
		metaType.GureguNullableType = "null.Int"
		metaType.ValueFormat = "%v"
	case "smallmoney":
		metaType.GoType = "float32"
		metaType.SQLNullableType = "sql.NullFloat32"
		metaType.GureguNullableType = "null.Float"
		metaType.ValueFormat = "%v"
	case "money", "real", "float", "numeric", "decimal":
		metaType.GoType = "float64"
		metaType.SQLNullableType = "sql.NullFloat64"
		metaType.GureguNullableType = "null.Float"
		metaType.ValueFormat = "%v"
	case "ntext", "text", "xml", "table", "char", "varchar", "nchar", "nvarchar", "varbinary", "binary", "image":
		metaType.GoType = "string"
		metaType.SQLNullableType = "sql.NullString"
		metaType.GureguNullableType = "null.String"
		metaType.ValueFormat = "'%v'"
	case "datetime", "datetime2", "smalldatetime", "date", "time", "datetimeoffset", "timestamp":
		metaType.GoType = "time.Time"
		metaType.GureguNullableType = "null.Time"
		metaType.SQLNullableType = "sql.NullTime"
		metaType.ValueFormat = "'%v'"
	default:
		if strings.Contains(columnType, "numeric") ||
			strings.Contains(columnType, "decimal") ||
			strings.Contains(columnType, "money") {
			metaType.GoType = "float64"
			metaType.SQLNullableType = "sql.NullFloat64"
			metaType.GureguNullableType = "null.Float"
			metaType.ValueFormat = "%v"
		} else {
			fmt.Printf("unknow column type : %s, replace it with \"string\"\n", columnType)
			metaType.GoType = "string"
			metaType.SQLNullableType = "sql.NullString"
			metaType.GureguNullableType = "null.String"
			metaType.ValueFormat = "'%v'"
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
		metaType.ValueFormat = "%v"
	case "smallint", "int2":
		metaType.GoType = "int16"
		metaType.SQLNullableType = "sql.NullInt32"
		metaType.GureguNullableType = "null.Int"
		metaType.ValueFormat = "%v"
	case "integer":
		metaType.GoType = "int32"
		metaType.SQLNullableType = "sql.NullInt32"
		metaType.GureguNullableType = "null.Int"
		metaType.ValueFormat = "%v"
	case "real", "binary float":
		metaType.GoType = "float32"
		metaType.SQLNullableType = "sql.NullFloat32"
		metaType.GureguNullableType = "null.Float"
		metaType.ValueFormat = "%v"
	case "float", "binary double", "number", "decimal":
		metaType.GoType = "float64"
		metaType.SQLNullableType = "sql.NullFloat64"
		metaType.GureguNullableType = "null.Float"
		metaType.ValueFormat = "%v"
	case "char", "long", "nchar", "varchar", "varchar2", "nvarchar2", "rowid", "nrowid",
		"clob", "nclob", "blob", "raw", "long raw", "bfile":
		metaType.GoType = "string"
		metaType.SQLNullableType = "sql.NullString"
		metaType.GureguNullableType = "null.String"
		metaType.ValueFormat = "'%v'"
	case "date", "timestamp":
		metaType.GoType = "time.Time"
		metaType.GureguNullableType = "null.Time"
		metaType.SQLNullableType = "sql.NullTime"
		metaType.ValueFormat = "'%v'"
	default:
		if strings.Contains(columnType, "number") ||
			strings.Contains(columnType, "decimal") {
			metaType.GoType = "float64"
			metaType.SQLNullableType = "sql.NullFloat64"
			metaType.GureguNullableType = "null.Float"
			metaType.ValueFormat = "%v"
		} else if strings.Contains(columnType, "character") {
			metaType.GoType = "string"
			metaType.SQLNullableType = "sql.NullString"
			metaType.GureguNullableType = "null.String"
			metaType.ValueFormat = "'%v'"
		} else if strings.Contains(columnType, "timestamp") ||
			strings.Contains(columnType, "time") ||
			strings.Contains(columnType, "interval") {
			metaType.GoType = "time.Time"
			metaType.GureguNullableType = "null.Time"
			metaType.SQLNullableType = "sql.NullTime"
			metaType.ValueFormat = "'%v'"
		} else {
			fmt.Printf("unknow column type : %s, replace it with \"string\"\n", columnType)
			metaType.GoType = "string"
			metaType.SQLNullableType = "sql.NullString"
			metaType.GureguNullableType = "null.String"
			metaType.ValueFormat = "'%v'"
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

// LowerCamelCaseToUpperCamelCase 小写驼峰单词转为大写驼峰单词
func LowerCamelCaseToUpperCamelCase(s string) string {
	if len(s) == 0 {
		return s
	}

	// 检查是否匹配特殊单词
	for word := range lowerAbbreviationMap {
		if strings.HasPrefix(s, word) {
			// 如果匹配特殊单词，直接返回小写形式
			return strings.ToUpper(word) + s[len(word):]
		}
	}

	// 将第一个字符转换为大写
	return string(unicode.ToUpper(rune(s[0]))) + s[1:]
}

// UpperCamelCaseToLowerCamelCase 大写驼峰单词转为小写驼峰单词
func UpperCamelCaseToLowerCamelCase(s string) string {
	if len(s) == 0 {
		return s
	}

	// 检查是否匹配特殊单词
	for word := range abbreviationMap {
		if strings.HasPrefix(s, word) {
			// 如果匹配特殊单词，直接返回小写形式
			return strings.ToLower(word) + s[len(word):]
		}
	}

	// 将第一个字符转换为小写
	return string(unicode.ToLower(rune(s[0]))) + s[1:]
}

// UnderscoreToLowerCamelCase 下划线单词转为小写驼峰单词
func UnderscoreToLowerCamelCase(s string) string {
	for key := range abbreviationMap {
		lowKey := strings.ToLower(key)
		if strings.HasPrefix(s, key) || strings.HasPrefix(s, lowKey) {
			return lowKey + s[len(key):]
		}
	}
	if ToUpper(s) == s {
		return ToLower(s)
	}
	s = UnderscoreToUpperCamelCase(s)
	return string(unicode.ToLower(rune(s[0]))) + s[1:]
}

// CamelCaseToUnderscore 驼峰单词转下划线单词
func CamelCaseToUnderscore(s string) string {
	var output []rune
	for i, r := range s {
		// 如果当前字符是大写字母
		if unicode.IsUpper(r) {
			// 如果不是第一个字符，并且前一个字符不是大写字母，或者下一个字符是小写字母
			if i > 0 && (!unicode.IsUpper(rune(s[i-1])) || (i+1 < len(s) && unicode.IsLower(rune(s[i+1])))) {
				output = append(output, '_')
			}
			output = append(output, unicode.ToLower(r))
		} else {
			output = append(output, r)
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

// WalkDir 获取指定目录及所有子目录下的所有文件，可以匹配后缀过滤。
func WalkDir(dirPth, suffix string) (files []string, err error) {
	files = make([]string, 0)
	suffix = strings.ToUpper(suffix) //忽略后缀匹配的大小写

	err = filepath.Walk(dirPth, func(filename string, fi os.FileInfo, err error) error { //遍历目录

		if fi.IsDir() { // 忽略目录
			return nil
		}
		if suffix == "" {
			files = append(files, fi.Name())
			return nil
		}
		if suffix != "" && strings.HasSuffix(strings.ToUpper(fi.Name()), suffix) {
			files = append(files, fi.Name())
		}

		return nil
	})

	return files, err
}

// GetFuncNamePath 获取函数所在模块路劲
func GetFuncNamePath(fn interface{}) string {
	value := reflect.ValueOf(fn)
	ptr := value.Pointer()
	ffp := runtime.FuncForPC(ptr)
	return ffp.Name()
}

func SupportGenericType() bool {
	versionInfo := runtime.Version()
	reg := regexp.MustCompile(`(\d+\.\d+\.*\d*)`)
	if reg == nil {
		return false
	}
	versionStr := reg.FindString(versionInfo)
	versionSlice := strings.Split(versionStr, ".")
	if len(versionSlice) >= 2 {
		version, _ := strconv.Atoi(versionSlice[1])
		return version >= 18
	}
	return false
}
