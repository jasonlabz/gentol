package metadata

import (
	"strings"

	"github.com/jasonlabz/gentol/gormx"
)

type DaoMeta struct {
	BaseConfig
	ModelModulePath  string
	ModelPackageName string
	ModelStructName  string
	DaoModulePath    string
	DaoPackageName   string
	PrimaryKeyList   []*PrimaryKeyInfo
	ColumnList       []*ColumnInfo
}

type PrimaryKeyInfo struct {
	GoColumnName       string
	GoColumnType       string
	GoColumnOriginType string
	GoFieldName        string
}

func (m *DaoMeta) GenRenderData() map[string]any {
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
		columnInfo.ValueFormat = metaType.ValueFormat
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
		if columnInfo.IsPrimaryKey && len(m.PrimaryKeyList) == 0 {
			m.PrimaryKeyList = append(m.PrimaryKeyList, &PrimaryKeyInfo{
				GoFieldName:        columnInfo.ColumnName,
				GoColumnName:       UnderscoreToLowerCamelCase(columnInfo.ColumnName),
				GoColumnType:       columnInfo.GoColumnType,
				GoColumnOriginType: columnInfo.GoColumnOriginType,
			})
		}
	}
	result := map[string]any{
		"ModelModulePath":     m.ModelModulePath,
		"DaoModulePath":       m.DaoModulePath,
		"ModelPackageName":    m.ModelPackageName,
		"DaoPackageName":      m.DaoPackageName,
		"ModelStructName":     m.ModelStructName,
		"ModelLowerCamelName": UnderscoreToLowerCamelCase(m.TableName),
		"ModelShortName":      ToLower(strings.Split(m.ModelStructName, "")[0]),
		"PrimaryKeyList":      m.PrimaryKeyList,
		"ColumnList":          m.ColumnList,
		"SchemaName":          m.SchemaName,
		"TableName":           m.TableName,
		"TitleTableName":      m.ModelStructName,
	}
	return result
}

const Dao = NotEditMark + `
package {{.DaoPackageName}}

import (
	"context"

	"{{.ModelModulePath}}"
)

type {{.ModelStructName}}Dao interface {
	// 可编辑自定义dao层逻辑
	{{.ModelStructName}}DaoExt

	// SelectByRawSQL 自定义SQL查询，满足连表查询场景
	SelectByRawSQL(ctx context.Context, rawSQL string, result any) (err error)

	// SelectAll 查询所有记录
	SelectAll(ctx context.Context, selectFields ...{{.ModelPackageName}}.{{.ModelStructName}}Field) (records []*{{.ModelPackageName}}.{{.ModelStructName}}, err error)
	
	// SelectOneByPrimaryKey 通过主键查询记录
	SelectOneByPrimaryKey(ctx context.Context, {{range .PrimaryKeyList}}{{.GoColumnName}} {{.GoColumnOriginType}}, {{end}}selectFields ...{{.ModelPackageName}}.{{.ModelStructName}}Field) (record *{{.ModelPackageName}}.{{.ModelStructName}}, err error)
	
	// SelectRecordByCondition 通过指定条件查询记录
	SelectRecordByCondition(ctx context.Context, condition *{{.ModelPackageName}}.Condition, selectFields ...{{.ModelPackageName}}.{{.ModelStructName}}Field) (records []*{{.ModelPackageName}}.{{.ModelStructName}}, err error)

	// SelectPageRecordByCondition 通过指定条件查询分页记录
	SelectPageRecordByCondition(ctx context.Context, condition *{{.ModelPackageName}}.Condition, pageParam *{{.ModelPackageName}}.Pagination,
		selectFields ...{{.ModelPackageName}}.{{.ModelStructName}}Field) (records []*{{.ModelPackageName}}.{{.ModelStructName}}, err error)
	
	// CountByCondition 通过指定条件查询记录数量
	CountByCondition(ctx context.Context, condition *{{.ModelPackageName}}.Condition) (count int64, err error)
	
	// DeleteByCondition 通过指定条件删除记录，返回删除记录数量
	DeleteByCondition(ctx context.Context, condition *{{.ModelPackageName}}.Condition) (affect int64, err error)
	
	// DeleteByPrimaryKey 通过主键删除记录，返回删除记录数量
	DeleteByPrimaryKey(ctx context.Context{{range .PrimaryKeyList}}, {{.GoColumnName}} {{.GoColumnOriginType}}{{end}}) (affect int64, err error)

	// UpdateRecord 更新记录
	UpdateRecord(ctx context.Context, record *{{.ModelPackageName}}.{{.ModelStructName}}) (affect int64, err error)

	// UpdateRecords 批量更新记录
	UpdateRecords(ctx context.Context, records []*{{.ModelPackageName}}.{{.ModelStructName}}) (affect int64, err error)

	// UpdateByCondition 更新指定条件下的记录
	UpdateByCondition(ctx context.Context, condition *{{.ModelPackageName}}.Condition, updateField {{.ModelPackageName}}.UpdateField) (affect int64, err error)
	
	// UpdateByPrimaryKey 更新主键的记录
	UpdateByPrimaryKey(ctx context.Context, {{range .PrimaryKeyList}}{{.GoColumnName}} {{.GoColumnOriginType}}, {{end}}updateField {{.ModelPackageName}}.UpdateField) (affect int64, err error)
	
	// Insert 插入记录
	Insert(ctx context.Context, record *{{.ModelPackageName}}.{{.ModelStructName}}) (affect int64, err error)
	
	// BatchInsert 批量插入记录
	BatchInsert(ctx context.Context, records []*{{.ModelPackageName}}.{{.ModelStructName}}) (affect int64, err error)
	
	// InsertOrUpdateOnDuplicateKey 插入记录，假如唯一键冲突则更新
	InsertOrUpdateOnDuplicateKey(ctx context.Context, record *{{.ModelPackageName}}.{{.ModelStructName}},
	uniqueKeys ...{{.ModelPackageName}}.{{.ModelStructName}}Field) (affect int64, err error)
	
	// BatchInsertOrUpdateOnDuplicateKey 批量插入记录，假如唯一键冲突则更新
	BatchInsertOrUpdateOnDuplicateKey(ctx context.Context, records []*{{.ModelPackageName}}.{{.ModelStructName}},
	uniqueKeys ...{{.ModelPackageName}}.{{.ModelStructName}}Field) (affect int64, err error)
}
`

const DaoExt = `
package dao

type {{.ModelStructName}}DaoExt interface {
}
`

const DaoImpl = NotEditMark + `
package impl

import (
	"context"
	"strings"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"{{.DaoModulePath}}"
	"{{.ModelModulePath}}"
)

var {{.ModelLowerCamelName}}Dao {{.DaoPackageName}}.{{.ModelStructName}}Dao = &{{.ModelLowerCamelName}}DaoImpl{}

func Get{{.ModelStructName}}Dao() {{.DaoPackageName}}.{{.ModelStructName}}Dao {
	return {{.ModelLowerCamelName}}Dao
}

type {{.ModelLowerCamelName}}DaoImpl struct{}

func ({{.ModelShortName}} {{.ModelLowerCamelName}}DaoImpl) tx(ctx context.Context) *gorm.DB {
	tx, ok := ctx.Value("transactionDB").(*gorm.DB)
	if ok {
		return tx
	}
	return {{.DaoPackageName}}.DB()
}

func ({{.ModelShortName}} {{.ModelLowerCamelName}}DaoImpl) SelectByRawSQL(ctx context.Context, rawSQL string, result any) (err error) {
	err = {{.ModelShortName}}.tx(ctx).WithContext(ctx).
		Raw(rawSQL).Scan(result).Error
	return
}

func ({{.ModelShortName}} {{.ModelLowerCamelName}}DaoImpl) SelectAll(ctx context.Context, selectFields ...{{.ModelPackageName}}.{{.ModelStructName}}Field) (records []*{{.ModelPackageName}}.{{.ModelStructName}}, err error) {
	tx := {{.ModelShortName}}.tx(ctx).WithContext(ctx).
		Table({{.ModelPackageName}}.TableName{{.ModelStructName}})
	if len(selectFields) > 0 {
		columns := make([]string, 0)
		for _, field := range selectFields {
			columns = append(columns, string(field))
		}
		tx = tx.Select(strings.Join(columns, ","))
	}
	err = tx.Find(&records).Error
	return
}

func ({{.ModelShortName}} {{.ModelLowerCamelName}}DaoImpl) SelectOneByPrimaryKey(ctx context.Context, {{range .PrimaryKeyList}}{{.GoColumnName}} {{.GoColumnOriginType}}, {{end}}selectFields ...{{.ModelPackageName}}.{{.ModelStructName}}Field) (record *{{.ModelPackageName}}.{{.ModelStructName}}, err error) {
	tx := {{.ModelShortName}}.tx(ctx).WithContext(ctx).
		Table({{.ModelPackageName}}.TableName{{.ModelStructName}})
	if len(selectFields) > 0 {
		columns := make([]string, 0)
		for _, field := range selectFields {
			columns = append(columns, string(field))
		}
		tx = tx.Select(strings.Join(columns, ","))
	}
	whereCondition := map[string]any{
 		{{ range .PrimaryKeyList -}}
		"{{- .GoFieldName -}}": {{- .GoColumnName }},
		{{ end }}
	}
	err = tx.Where(whereCondition).First(&record).Error
	return
}

func ({{.ModelShortName}} {{.ModelLowerCamelName}}DaoImpl) SelectRecordByCondition(ctx context.Context, condition *{{.ModelPackageName}}.Condition, selectFields ...{{.ModelPackageName}}.{{.ModelStructName}}Field) (records []*{{.ModelPackageName}}.{{.ModelStructName}}, err error) {
	if condition == nil {
		return {{.ModelShortName}}.SelectAll(ctx, selectFields...)
	}
	tx := {{.ModelShortName}}.tx(ctx).WithContext(ctx).
		Table({{.ModelPackageName}}.TableName{{.ModelStructName}})
	if len(selectFields) > 0 {
		columns := make([]string, 0)
		for _, field := range selectFields {
			columns = append(columns, string(field))
		}
		tx = tx.Select(strings.Join(columns, ","))
	}
	for _, strCondition := range condition.StringCondition {
		tx = tx.Where(strCondition)
	}
	if len(condition.MapCondition) > 0 {
		tx = tx.Where(condition.MapCondition)
	}
	for _, order := range condition.OrderByClause {
		tx = tx.Order(order)
	}
	err = tx.Find(&records).Error
	return
}

func ({{.ModelShortName}} {{.ModelLowerCamelName}}DaoImpl) SelectPageRecordByCondition(ctx context.Context, condition *{{.ModelPackageName}}.Condition, pageParam *{{.ModelPackageName}}.Pagination,
	selectFields ...{{.ModelPackageName}}.{{.ModelStructName}}Field) (records []*{{.ModelPackageName}}.{{.ModelStructName}}, err error) {
	tx := {{.ModelShortName}}.tx(ctx).WithContext(ctx).
		Table({{.ModelPackageName}}.TableName{{.ModelStructName}})
	if len(selectFields) > 0 {
		columns := make([]string, 0)
		for _, field := range selectFields {
			columns = append(columns, string(field))
		}
		tx = tx.Select(strings.Join(columns, ","))
	}

	if condition != nil {
		for _, strCondition := range condition.StringCondition {
			tx = tx.Where(strCondition)
		}
		if len(condition.MapCondition) > 0 {
			tx = tx.Where(condition.MapCondition)
		}
		for _, order := range condition.OrderByClause {
			tx = tx.Order(order)
		}
	}
	var count int64
	if pageParam != nil {
		tx = tx.Count(&count).Offset(int(pageParam.CalculateOffset())).Limit(int(pageParam.PageSize))
	}
	err = tx.Find(&records).Error
	if pageParam != nil {
		pageParam.Total = count
		pageParam.CalculatePageCount()
	}
	return
}

func ({{.ModelShortName}} {{.ModelLowerCamelName}}DaoImpl) CountByCondition(ctx context.Context, condition *{{.ModelPackageName}}.Condition) (count int64, err error) {
	tx := {{.ModelShortName}}.tx(ctx).WithContext(ctx).
		Table({{.ModelPackageName}}.TableName{{.ModelStructName}})
	if condition != nil {
		for _, strCondition := range condition.StringCondition {
			tx = tx.Where(strCondition)
		}
		if len(condition.MapCondition) > 0 {
			tx = tx.Where(condition.MapCondition)
		}
	}
	err = tx.Count(&count).Error
	return
}

func ({{.ModelShortName}} {{.ModelLowerCamelName}}DaoImpl) DeleteByCondition(ctx context.Context, condition *{{.ModelPackageName}}.Condition) (affect int64, err error) {
	tx := {{.ModelShortName}}.tx(ctx).WithContext(ctx)
	if condition != nil {
		for _, strCondition := range condition.StringCondition {
			tx = tx.Where(strCondition)
		}
		if len(condition.MapCondition) > 0 {
			tx = tx.Where(condition.MapCondition)
		}
	}
	tx = tx.Delete(&{{.ModelPackageName}}.{{.ModelStructName}}{})
	affect = tx.RowsAffected
	err = tx.Error
	return
}

func ({{.ModelShortName}} {{.ModelLowerCamelName}}DaoImpl) DeleteByPrimaryKey(ctx context.Context{{range .PrimaryKeyList}}, {{.GoColumnName}} {{.GoColumnOriginType}}{{end}}) (affect int64, err error) {
	whereCondition := map[string]any{
 		{{ range .PrimaryKeyList -}}
		"{{- .GoFieldName -}}": {{- .GoColumnName }},
		{{ end }}
	}	
	tx := {{.ModelShortName}}.tx(ctx).WithContext(ctx).Where(whereCondition).Delete(&{{.ModelPackageName}}.{{.ModelStructName}}{})
	affect = tx.RowsAffected
	err = tx.Error
	return
}

func ({{.ModelShortName}} {{.ModelLowerCamelName}}DaoImpl) UpdateRecord(ctx context.Context, record *{{.ModelPackageName}}.{{.ModelStructName}}) (affect int64, err error) {
	tx := {{.ModelShortName}}.tx(ctx).WithContext(ctx).
		Table({{.ModelPackageName}}.TableName{{.ModelStructName}}).
		Save(record)
	affect = tx.RowsAffected
	err = tx.Error
	return
}

func ({{.ModelShortName}} {{.ModelLowerCamelName}}DaoImpl) UpdateRecords(ctx context.Context, records []*{{.ModelPackageName}}.{{.ModelStructName}}) (affect int64, err error) {
	tx := {{.ModelShortName}}.tx(ctx).WithContext(ctx).
		Table({{.ModelPackageName}}.TableName{{.ModelStructName}}).
		Save(records)
	affect = tx.RowsAffected
	err = tx.Error
	return
}

func ({{.ModelShortName}} {{.ModelLowerCamelName}}DaoImpl) UpdateByCondition(ctx context.Context, condition *{{.ModelPackageName}}.Condition, updateField {{.ModelPackageName}}.UpdateField) (affect int64, err error) {
	tx := {{.ModelShortName}}.tx(ctx).WithContext(ctx).
		Table({{.ModelPackageName}}.TableName{{.ModelStructName}})
		if condition != nil {
		for _, strCondition := range condition.StringCondition {
			tx = tx.Where(strCondition)
		}
		if len(condition.MapCondition) > 0 {
			tx = tx.Where(condition.MapCondition)
		}
	}
	tx = tx.Updates(map[string]any(updateField))
	affect = tx.RowsAffected
	err = tx.Error
	return
}

func ({{.ModelShortName}} {{.ModelLowerCamelName}}DaoImpl) UpdateByPrimaryKey(ctx context.Context, {{range .PrimaryKeyList}}{{.GoColumnName}} {{.GoColumnOriginType}}, {{end}}updateField {{.ModelPackageName}}.UpdateField) (affect int64, err error) {
	whereCondition := map[string]any{
 		{{ range .PrimaryKeyList -}}
		"{{- .GoFieldName -}}": {{- .GoColumnName }},
		{{ end }}
	}
	tx := {{.ModelShortName}}.tx(ctx).WithContext(ctx).
		Table({{.ModelPackageName}}.TableName{{.ModelStructName}}).
		Where(whereCondition)
	tx = tx.Updates(map[string]any(updateField))
	affect = tx.RowsAffected
	err = tx.Error
	return
}

func ({{.ModelShortName}} {{.ModelLowerCamelName}}DaoImpl) Insert(ctx context.Context, record *{{.ModelPackageName}}.{{.ModelStructName}}) (affect int64, err error) {
	tx := {{.ModelShortName}}.tx(ctx).WithContext(ctx).
		Table({{.ModelPackageName}}.TableName{{.ModelStructName}}).
		Create(&record)
	affect = tx.RowsAffected
	err = tx.Error
	return
}

func ({{.ModelShortName}} {{.ModelLowerCamelName}}DaoImpl) BatchInsert(ctx context.Context, records []*{{.ModelPackageName}}.{{.ModelStructName}}) (affect int64, err error) {
	tx := {{.ModelShortName}}.tx(ctx).WithContext(ctx).
		Table({{.ModelPackageName}}.TableName{{.ModelStructName}}).
		Create(&records)
	affect = tx.RowsAffected
	err = tx.Error
	return
}

func ({{.ModelShortName}} {{.ModelLowerCamelName}}DaoImpl) InsertOrUpdateOnDuplicateKey(ctx context.Context, record *{{.ModelPackageName}}.{{.ModelStructName}},
	uniqueKeys ...{{.ModelPackageName}}.{{.ModelStructName}}Field) (affect int64, err error) {
	columns := make([]clause.Column, 0)
	for _, field := range uniqueKeys {
		columns = append(columns, clause.Column{
			Name: string(field),
		})
	}
	tx := {{.ModelShortName}}.tx(ctx).WithContext(ctx).
		Table({{.ModelPackageName}}.TableName{{.ModelStructName}}).
		Clauses(clause.OnConflict{
			Columns:   columns,
			UpdateAll: true,
		}).Create(&record)
	affect = tx.RowsAffected
	err = tx.Error
	return
}

func ({{.ModelShortName}} {{.ModelLowerCamelName}}DaoImpl) BatchInsertOrUpdateOnDuplicateKey(ctx context.Context, records []*{{.ModelPackageName}}.{{.ModelStructName}},
	uniqueKeys ...{{.ModelPackageName}}.{{.ModelStructName}}Field) (affect int64, err error) {
	columns := make([]clause.Column, 0)
	for _, field := range uniqueKeys {
		columns = append(columns, clause.Column{
			Name: string(field),
		})
	}
	tx := {{.ModelShortName}}.tx(ctx).WithContext(ctx).
		Table({{.ModelPackageName}}.TableName{{.ModelStructName}}).
		Clauses(clause.OnConflict{
			Columns:   columns,
			UpdateAll: true,
		}).Create(&records)
	affect = tx.RowsAffected
	err = tx.Error
	return
}



`

const DaoExtImpl = `
package impl

// CustomMethod 自定义方法, 该文件不会被覆盖
// func ({{.ModelShortName}} {{.ModelLowerCamelName}}DaoImpl) CustomMethod(ctx context.Context, rawSQL string, result any) (err error) {
//		return
// }
`

const Database = NotEditMark + `
package dao

import (
	"context"

	"gorm.io/gorm"
)

var gormDB *gorm.DB

func SetGormDB(db *gorm.DB) {
	if db == nil {
		panic("db connection is nil")
	}
	gormDB = db
	return
}

func DB() *gorm.DB {
	if gormDB == nil {
		panic("db connection is nil")
	}
	return gormDB
}

func RunTransaction(ctx context.Context, f func(ctx context.Context) error) error {
	// 使用 Transaction 方法并绑定上下文
	return DB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		transCtx := context.WithValue(ctx, "transactionDB", tx)
		// 在事务中创建用户
		if err := f(transCtx); err != nil {
			return err // 返回错误，事务会回滚
		}
		return nil // 返回 nil，事务会提交
	})
}
`
