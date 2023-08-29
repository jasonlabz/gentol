package metadata

import (
	"github.com/jasonlabz/gentol/gormx"
	"strings"
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
	GoColumnName string
	GoColumnType string
	GoFieldName  string
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
		if columnInfo.IsPrimaryKey {
			m.PrimaryKeyList = append(m.PrimaryKeyList, &PrimaryKeyInfo{
				GoFieldName:  columnInfo.GoColumnName,
				GoColumnName: UnderscoreToLowerCamelCase(columnInfo.ColumnName),
				GoColumnType: columnInfo.GoColumnType,
			})
		}
	}
	result := map[string]any{
		"ModelModulePath":  m.ModelModulePath,
		"DaoModulePath":    m.DaoModulePath,
		"ModelPackageName": m.ModelPackageName,
		"DaoPackageName":   m.DaoPackageName,
		"ModelStructName":  m.ModelStructName,
		"ModelShortName":   ToLower(strings.Split(m.ModelStructName, "")[0]),
		"PrimaryKeyList":   m.PrimaryKeyList,
		"ColumnList":       m.ColumnList,
		"SchemaName":       m.SchemaName,
		"TableName":        m.TableName,
		"TitleTableName":   m.ModelStructName,
	}
	return result
}

const Dao = NotEditMark + `
package interfaces

import (
	"context"

	"{{.ModelModulePath}}"
)

type {{.ModelStructName}}Dao interface {
	// SelectOneByPrimaryKey 通过主键查询记录
	SelectOneByPrimaryKey(ctx context.Context, {{range .PrimaryKeyList}}{{.GoColumnName}} {{.GoColumnType}}, {{end}}selectFields ...model.{{.ModelStructName}}Field) (record *model.{{.ModelStructName}}, err error)
	
	// SelectRecordByCondition 通过指定条件查询记录
	SelectRecordByCondition(ctx context.Context, condition *model.Condition, selectFields ...model.{{.ModelStructName}}Field) (records []*model.{{.ModelStructName}}, err error)
	
	// SelectRecordASCByCondition 通过指定条件查询记录,升序排序
	SelectRecordASCByCondition(ctx context.Context, condition *model.Condition, clause model.OrderByClause,
		selectFields ...model.{{.ModelStructName}}Field) (records []*model.{{.ModelStructName}}, err error)
	
	// SelectRecordDESCByCondition 通过指定条件查询记录,降序排序
	SelectRecordDESCByCondition(ctx context.Context, condition *model.Condition, clause model.OrderByClause,
		selectFields ...model.{{.ModelStructName}}Field) (records []*model.{{.ModelStructName}}, err error)
	
	// SelectPageRecordByCondition 通过指定条件查询分页记录
	SelectPageRecordByCondition(ctx context.Context, condition *model.Condition, pageParam *model.Pagination,
		selectFields ...model.{{.ModelStructName}}Field) (records []*model.{{.ModelStructName}}, err error)
	
	// SelectPageRecordASCByCondition 通过指定条件查询分页记录,升序排序
	SelectPageRecordASCByCondition(ctx context.Context, condition *model.Condition, clause model.OrderByClause,
		pageParam *model.Pagination, selectFields ...model.{{.ModelStructName}}Field) (records []*model.{{.ModelStructName}}, err error)
	
	// SelectPageRecordDESCByCondition 通过指定条件查询分页记录,降序排序
	SelectPageRecordDESCByCondition(ctx context.Context, condition *model.Condition, clause model.OrderByClause,
		pageParam *model.Pagination, selectFields ...model.{{.ModelStructName}}Field) (records []*model.{{.ModelStructName}}, err error)
	
	// CountByCondition 通过指定条件查询记录数量
	CountByCondition(ctx context.Context, condition *model.Condition) (count int64, err error)
	
	// DeleteByCondition 通过指定条件删除记录，返回删除记录数量
	DeleteByCondition(ctx context.Context, condition *model.Condition) (affect int64, err error)
	
	// DeleteByPrimaryKey 通过主键删除记录，返回删除记录数量
	DeleteByPrimaryKey(ctx context.Context{{range .PrimaryKeyList}}, {{.GoColumnName}} {{.GoColumnType}}{{end}}) (affect int64, err error)
	
	// UpdateByCondition 更新指定条件下的记录
	UpdateByCondition(ctx context.Context, condition *model.Condition, updateField *model.UpdateField) (affect int64, err error)
	
	// UpdateByPrimaryKey 更新主键的记录
	UpdateByPrimaryKey(ctx context.Context, {{range .PrimaryKeyList}}{{.GoColumnName}} {{.GoColumnType}}, {{end}}updateField *model.UpdateField) (affect int64, err error)
	
	// Insert 插入记录
	Insert(ctx context.Context, record *model.{{.ModelStructName}}) (affect int64, err error)
	
	// BatchInsert 批量插入记录
	BatchInsert(ctx context.Context, records []*model.{{.ModelStructName}}) (affect int64, err error)
	
	// InsertOrUpdateOnDuplicateKey 插入记录，假如唯一键冲突则更新
	InsertOrUpdateOnDuplicateKey(ctx context.Context, record *model.{{.ModelStructName}}, updateFields ...model.{{.ModelStructName}}Field) (affect int64, err error)
	
	// BatchInsertOrUpdateOnDuplicateKey 批量插入记录，假如唯一键冲突则更新
	BatchInsertOrUpdateOnDuplicateKey(ctx context.Context, records []*model.{{.ModelStructName}}, updateFields ...model.{{.ModelStructName}}Field) (affect int64, err error)
}

`

const DaoImpl = NotEditMark + `
package {{.DaoPackageName}}

import (
	"context"
	"gorm.io/gorm/clause"
	"strings"

	"{{.DaoModulePath}}/interfaces"
	"{{.ModelModulePath}}"
)

var _ interfaces.{{.ModelStructName}}Dao = &{{.ModelStructName}}DaoImpl{}

type {{.ModelStructName}}DaoImpl struct{}

func ({{.ModelShortName}} {{.ModelStructName}}DaoImpl) SelectOneByPrimaryKey(ctx context.Context, {{range .PrimaryKeyList}}{{.GoColumnName}} {{.GoColumnType}}, {{end}}selectFields ...model.{{.ModelStructName}}Field) (record *model.{{.ModelStructName}}, err error) {
	tx := DB().WithContext(ctx).
		Table(model.TableName{{.ModelStructName}})
	if len(selectFields) > 0 {
		columns := make([]string, 0)
		for _, field := range selectFields {
			columns = append(columns, string(field))
		}
		tx = tx.Select(strings.Join(columns, ","))
	}
	err = tx.Where("record_id", recordID).First(&record).Error
	return
}

func ({{.ModelShortName}} {{.ModelStructName}}DaoImpl) SelectRecordByCondition(ctx context.Context, condition *model.Condition, selectFields ...model.{{.ModelStructName}}Field) (records []*model.{{.ModelStructName}}, err error) {
	tx := DB().WithContext(ctx).
		Table(model.TableName{{.ModelStructName}})
	if len(selectFields) > 0 {
		columns := make([]string, 0)
		for _, field := range selectFields {
			columns = append(columns, string(field))
		}
		tx = tx.Select(strings.Join(columns, ","))
	}
	if condition != nil && len(condition.MapCondition) > 0 {
		tx = tx.Where(condition.MapCondition)
	}
	if condition != nil && len(condition.StringCondition) > 0 {
		for _, strCondition := range condition.StringCondition {
			tx = tx.Where(strCondition)
		}
	}
	err = tx.Find(&records).Error
	return
}

func ({{.ModelShortName}} {{.ModelStructName}}DaoImpl) SelectRecordASCByCondition(ctx context.Context, condition *model.Condition, clause model.OrderByClause,
	selectFields ...model.{{.ModelStructName}}Field) (records []*model.{{.ModelStructName}}, err error) {
	tx := DB().WithContext(ctx).
		Table(model.TableName{{.ModelStructName}})
	if len(selectFields) > 0 {
		columns := make([]string, 0)
		for _, field := range selectFields {
			columns = append(columns, string(field))
		}
		tx = tx.Select(strings.Join(columns, ","))
	}
	if condition != nil && len(condition.MapCondition) > 0 {
		tx = tx.Where(condition.MapCondition)
	}
	if condition != nil && len(condition.StringCondition) > 0 {
		for _, strCondition := range condition.StringCondition {
			tx = tx.Where(strCondition)
		}
	}
	if len(clause) > 0 {
		columns := make([]string, 0)
		for _, field := range clause {
			columns = append(columns, field)
		}
		tx = tx.Order(strings.Join(columns, ",") + " asc")
	}
	err = tx.Find(&records).Error
	return
}

func ({{.ModelShortName}} {{.ModelStructName}}DaoImpl) SelectRecordDESCByCondition(ctx context.Context, condition *model.Condition, clause model.OrderByClause,
	selectFields ...model.{{.ModelStructName}}Field) (records []*model.{{.ModelStructName}}, err error) {
	tx := DB().WithContext(ctx).
		Table(model.TableName{{.ModelStructName}})
	if len(selectFields) > 0 {
		columns := make([]string, 0)
		for _, field := range selectFields {
			columns = append(columns, string(field))
		}
		tx = tx.Select(strings.Join(columns, ","))
	}
	if condition != nil && len(condition.MapCondition) > 0 {
		tx = tx.Where(condition.MapCondition)
	}
	if condition != nil && len(condition.StringCondition) > 0 {
		for _, strCondition := range condition.StringCondition {
			tx = tx.Where(strCondition)
		}
	}
	if len(clause) > 0 {
		columns := make([]string, 0)
		for _, field := range clause {
			columns = append(columns, field)
		}
		tx = tx.Order(strings.Join(columns, ",") + " desc")
	}
	err = tx.Find(&records).Error
	return
}

func ({{.ModelShortName}} {{.ModelStructName}}DaoImpl) SelectPageRecordByCondition(ctx context.Context, condition *model.Condition, pageParam *model.Pagination,
	selectFields ...model.{{.ModelStructName}}Field) (records []*model.{{.ModelStructName}}, err error) {
	tx := DB().WithContext(ctx).
		Table(model.TableName{{.ModelStructName}})
	if len(selectFields) > 0 {
		columns := make([]string, 0)
		for _, field := range selectFields {
			columns = append(columns, string(field))
		}
		tx = tx.Select(strings.Join(columns, ","))
	}
	if condition != nil && len(condition.MapCondition) > 0 {
		tx = tx.Where(condition.MapCondition)
	}
	if condition != nil && len(condition.StringCondition) > 0 {
		for _, strCondition := range condition.StringCondition {
			tx = tx.Where(strCondition)
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

func ({{.ModelShortName}} {{.ModelStructName}}DaoImpl) SelectPageRecordASCByCondition(ctx context.Context, condition *model.Condition, clause model.OrderByClause,
	pageParam *model.Pagination, selectFields ...model.{{.ModelStructName}}Field) (records []*model.{{.ModelStructName}}, err error) {
	tx := DB().WithContext(ctx).
		Table(model.TableName{{.ModelStructName}})
	if len(selectFields) > 0 {
		columns := make([]string, 0)
		for _, field := range selectFields {
			columns = append(columns, string(field))
		}
		tx = tx.Select(strings.Join(columns, ","))
	}
	if condition != nil && len(condition.MapCondition) > 0 {
		tx = tx.Where(condition.MapCondition)
	}
	if condition != nil && len(condition.StringCondition) > 0 {
		for _, strCondition := range condition.StringCondition {
			tx = tx.Where(strCondition)
		}
	}
	if len(clause) > 0 {
		columns := make([]string, 0)
		for _, field := range clause {
			columns = append(columns, field)
		}
		tx = tx.Order(strings.Join(columns, ",") + " asc")
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

func ({{.ModelShortName}} {{.ModelStructName}}DaoImpl) SelectPageRecordDESCByCondition(ctx context.Context, condition *model.Condition, clause model.OrderByClause,
	pageParam *model.Pagination, selectFields ...model.{{.ModelStructName}}Field) (records []*model.{{.ModelStructName}}, err error) {
	tx := DB().WithContext(ctx).
		Table(model.TableName{{.ModelStructName}})
	if len(selectFields) > 0 {
		columns := make([]string, 0)
		for _, field := range selectFields {
			columns = append(columns, string(field))
		}
		tx = tx.Select(strings.Join(columns, ","))
	}
	if condition != nil && len(condition.MapCondition) > 0 {
		tx = tx.Where(condition.MapCondition)
	}
	if condition != nil && len(condition.StringCondition) > 0 {
		for _, strCondition := range condition.StringCondition {
			tx = tx.Where(strCondition)
		}
	}
	if len(clause) > 0 {
		columns := make([]string, 0)
		for _, field := range clause {
			columns = append(columns, field)
		}
		tx = tx.Order(strings.Join(columns, ",") + " desc")
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

func ({{.ModelShortName}} {{.ModelStructName}}DaoImpl) CountByCondition(ctx context.Context, condition *model.Condition) (count int64, err error) {
	tx := DB().WithContext(ctx).
		Table(model.TableName{{.ModelStructName}})
	if condition != nil && len(condition.MapCondition) > 0 {
		tx = tx.Where(condition.MapCondition)
	}
	if condition != nil && len(condition.StringCondition) > 0 {
		for _, strCondition := range condition.StringCondition {
			tx = tx.Where(strCondition)
		}
	}
	err = tx.Count(&count).Error
	return
}

func ({{.ModelShortName}} {{.ModelStructName}}DaoImpl) DeleteByCondition(ctx context.Context, condition *model.Condition) (affect int64, err error) {
	tx := DB().WithContext(ctx)
	if condition != nil && len(condition.MapCondition) > 0 {
		tx = tx.Where(condition.MapCondition)
	}
	if condition != nil && len(condition.StringCondition) > 0 {
		for _, strCondition := range condition.StringCondition {
			tx = tx.Where(strCondition)
		}
	}
	tx = tx.Delete(&model.{{.ModelStructName}}{})
	affect = tx.RowsAffected
	err = tx.Error
	return
}

func ({{.ModelShortName}} {{.ModelStructName}}DaoImpl) DeleteByPrimaryKey(ctx context.Context{{range .PrimaryKeyList}}, {{.GoColumnName}} {{.GoColumnType}}{{end}}) (affect int64, err error) {
	tx := DB().WithContext(ctx).Delete(&model.{{.ModelStructName}}{
			{{range .PrimaryKeyList}}{{.GoFieldName}}: {{.GoColumnName}}, {{end}}
		})
	affect = tx.RowsAffected
	err = tx.Error
	return
}

func ({{.ModelShortName}} {{.ModelStructName}}DaoImpl) UpdateByCondition(ctx context.Context, condition *model.Condition, updateField *model.UpdateField) (affect int64, err error) {
	tx := DB().WithContext(ctx).
		Table(model.TableName{{.ModelStructName}})
	if condition != nil && len(condition.MapCondition) > 0 {
		tx = tx.Where(condition.MapCondition)
	}
	if condition != nil && len(condition.StringCondition) > 0 {
		for _, strCondition := range condition.StringCondition {
			tx = tx.Where(strCondition)
		}
	}
	tx = tx.Updates({{.ModelShortName}}pdateField)
	affect = tx.RowsAffected
	err = tx.Error
	return
}

func ({{.ModelShortName}} {{.ModelStructName}}DaoImpl) UpdateByPrimaryKey(ctx context.Context, {{range .PrimaryKeyList}}{{.GoColumnName}} {{.GoColumnType}}, {{end}}updateField *model.UpdateField) (affect int64, err error) {
	tx := DB().WithContext(ctx).
		Table(model.TableName{{.ModelStructName}}).
		Where("record_id = ?", recordID)
	tx = tx.Updates(updateField)
	affect = tx.RowsAffected
	err = tx.Error
	return
}

func ({{.ModelShortName}} {{.ModelStructName}}DaoImpl) Insert(ctx context.Context, record *model.{{.ModelStructName}}) (affect int64, err error) {
	tx := DB().WithContext(ctx).
		Table(model.TableName{{.ModelStructName}}).
		Create(&record)
	affect = tx.RowsAffected
	err = tx.Error
	return
}

func ({{.ModelShortName}} {{.ModelStructName}}DaoImpl) BatchInsert(ctx context.Context, records []*model.{{.ModelStructName}}) (affect int64, err error) {
	tx := DB().WithContext(ctx).
		Table(model.TableName{{.ModelStructName}}).
		Create(&records)
	affect = tx.RowsAffected
	err = tx.Error
	return
}

func ({{.ModelShortName}} {{.ModelStructName}}DaoImpl) InsertOrUpdateOnDuplicateKey(ctx context.Context, record *model.{{.ModelStructName}}, updateFields ...model.{{.ModelStructName}}Field) (affect int64, err error) {
	tx := DB().WithContext(ctx).
		Table(model.TableName{{.ModelStructName}}).
		Clauses(clause.OnConflict{
			UpdateAll: true,
		}).Create(&record)
	affect = tx.RowsAffected
	err = tx.Error
	return
}

func ({{.ModelShortName}} {{.ModelStructName}}DaoImpl) BatchInsertOrUpdateOnDuplicateKey(ctx context.Context, records []*model.{{.ModelStructName}}, updateFields ...model.{{.ModelStructName}}Field) (affect int64, err error) {
	tx := DB().WithContext(ctx).
		Table(model.TableName{{.ModelStructName}}).
		Clauses(clause.OnConflict{
			UpdateAll: true,
		}).Create(&records)
	affect = tx.RowsAffected
	err = tx.Error
	return
}



`
const Database = NotEditMark + `
package {{.DaoPackageName}}

import "gorm.io/gorm"

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
`
