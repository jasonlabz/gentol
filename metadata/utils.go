package metadata

import (
	"fmt"
	"github.com/jasonlabz/gentol/gormx"
	"strings"
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
