package main

import (
	"bytes"
	"fmt"
	"github.com/jasonlabz/gentol/configx"
	"go/format"
	"gorm.io/gorm"
	"io"
	"os"
	"path/filepath"
	"text/template"

	"github.com/jasonlabz/gentol/metadata"
)

// RenderingTemplate rendering a template with data
func RenderingTemplate(templateInfo *metadata.Template, dataGen metadata.IBaseData, outFilePath string, overwrite bool) (err error) {
	var file *os.File
	data := dataGen.GenRenderData()
	if !IsExist(outFilePath) && !overwrite {
		file, err = os.OpenFile(outFilePath, os.O_CREATE|os.O_RDWR, 0666)
		if err != nil {
			fmt.Printf("open file error %s\n", err.Error())
			return
		}
	} else {
		if overwrite {
			file, err = os.OpenFile(outFilePath, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0666)
			if err != nil {
				fmt.Printf("overwrite true: open file error %s\n", err.Error())
				return
			}
		} else {
			// skip
			return
		}
	}
	fileName := filepath.Base(outFilePath)

	tmpl, err := template.New(fileName).Option("missingkey=error").Parse(templateInfo.Content)
	if err != nil {
		return
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, data)
	if err != nil {
		return fmt.Errorf("error in rendering %s: %s", templateInfo.Name, err.Error())
	}

	fileContents, err := Format(templateInfo, buf.Bytes(), outFilePath)
	if err != nil {
		return fmt.Errorf("error writing %s - error: %v", outFilePath, err)
	}

	_, err = io.Copy(file, bytes.NewReader(fileContents))
	if err != nil {
		return fmt.Errorf("error writing %s - error: %v", outFilePath, err)
	}

	fmt.Printf("writing %s\n", outFilePath)

	return nil
}

func Format(templateInfo *metadata.Template, content []byte, outputFile string) ([]byte, error) {
	extension := filepath.Ext(outputFile)
	if extension == ".go" {
		formattedSource, err := format.Source([]byte(content))
		if err != nil {
			return nil, fmt.Errorf("error in formatting template: %s outputfile: %s source: %s", templateInfo.Name, outputFile, err.Error())
		}

		fileContents := NormalizeNewlines(formattedSource)
		fileContents = CRLFNewlines(formattedSource)
		return fileContents, nil
	}

	fileContents := NormalizeNewlines([]byte(content))
	fileContents = CRLFNewlines(fileContents)
	return fileContents, nil
}

// NormalizeNewlines normalizes \r\n (windows) and \r (mac)
// into \n (unix)
func NormalizeNewlines(d []byte) []byte {
	// replace CR LF \r\n (windows) with LF \n (unix)
	d = bytes.Replace(d, []byte{13, 10}, []byte{10}, -1)
	// replace CF \r (mac) with LF \n (unix)
	d = bytes.Replace(d, []byte{13}, []byte{10}, -1)
	return d
}

// CRLFNewlines transforms \n to \r\n (windows)
func CRLFNewlines(d []byte) []byte {
	// replace LF (unix) with CR LF \r\n (windows)
	d = bytes.Replace(d, []byte{10}, []byte{13, 10}, -1)
	return d
}

func WriteModel(dbInfo *configx.DBTableInfo, schemaName, tableName string, columnTypes []gorm.ColumnType) {
	modelData := &metadata.ModelMeta{
		ModelPackageName: func() string {
			if dbInfo.ModelPath == "" {
				dbInfo.ModelPath = "dal/db/model"
			}
			return metadata.ToLower(filepath.Base(dbInfo.ModelPath))
		}(),
		ModelStructName: metadata.UnderscoreToUpperCamelCase(tableName),
	}
	columnTempList := make([]*metadata.ColumnInfo, 0)
	getColumnInfo(columnTypes, &columnTempList)
	modelData.ColumnList = columnTempList
	modelData.DBType = dbInfo.DBType
	modelData.SchemaName = schemaName
	modelData.TableName = tableName
	modelData.ModelPath = dbInfo.ModelPath
	modelTpl, ok := metadata.LoadTpl("model")
	if !ok {
		fmt.Println("undefined template" + "model")
		return
	}
	exist := IsExist(modelData.ModelPath)
	if !exist {
		_ = os.MkdirAll(modelData.ModelPath, 0777)
	}
	ff, _ := filepath.Abs(filepath.Join(modelData.ModelPath, metadata.CamelCaseToUnderscore(modelData.TableName)+".go"))
	err := RenderingTemplate(modelTpl, modelData, ff, true)
	if err != nil {
		fmt.Println(err)
		return
	}

	hookFile := filepath.Join(modelData.ModelPath, metadata.CamelCaseToUnderscore(modelData.TableName)+"_hook.go")
	exist = IsExist(hookFile)
	if !exist && false {
		ff, _ = filepath.Abs(hookFile)
		modelHookTpl, ok := metadata.LoadTpl("model_hook")
		if !ok {
			fmt.Println("undefined template" + "model_hook")
			return
		}
		err = RenderingTemplate(modelHookTpl, modelData, ff, true)
		if err != nil {
			fmt.Println(err)
			return
		}
	}
	baseFile := filepath.Join(modelData.ModelPath, "base.go")
	exist = IsExist(baseFile)
	if !exist {
		ff, _ = filepath.Abs(baseFile)
		modelBaseTpl, ok := metadata.LoadTpl("model_base")
		if !ok {
			fmt.Println("undefined template" + "model_base")
			return
		}
		err = RenderingTemplate(modelBaseTpl, modelData, ff, true)
		if err != nil {
			panic(err)
		}
	}
	return
}

func WriteDao(dbInfo *configx.DBTableInfo, schemaName, tableName string, columnTypes []gorm.ColumnType) {
	daoData := &metadata.DaoMeta{
		ModelPackageName: metadata.ToLower(filepath.Base(dbInfo.ModelPath)),
		DaoPackageName:   metadata.ToLower(filepath.Base(dbInfo.DaoPath)),
		ModelModulePath:  dbInfo.ModelModule,
		DaoModulePath:    dbInfo.DaoModule,
		ModelStructName:  metadata.UnderscoreToUpperCamelCase(tableName),
	}
	columnTempList := make([]*metadata.ColumnInfo, 0)
	getColumnInfo(columnTypes, &columnTempList)
	daoData.ColumnList = columnTempList
	daoData.DBType = dbInfo.DBType
	daoData.SchemaName = schemaName
	daoData.TableName = tableName
	daoData.ModelPath = dbInfo.ModelPath
	daoData.DaoPath = dbInfo.DaoPath
	daoTpl, ok := metadata.LoadTpl("dao")
	if !ok {
		fmt.Println("undefined template" + "dao")
		return
	}
	daoInterfacePath := filepath.Join(daoData.DaoPath, "interfaces")
	interfaceExist := IsExist(daoInterfacePath)
	if !interfaceExist {
		_ = os.MkdirAll(daoInterfacePath, 0777)
	}

	ff, _ := filepath.Abs(filepath.Join(daoInterfacePath, metadata.CamelCaseToUnderscore(daoData.TableName)+"_dao.go"))
	err := RenderingTemplate(daoTpl, daoData, ff, true)
	if err != nil {
		fmt.Println(err)
		return
	}

	daoImplFile := filepath.Join(daoData.DaoPath, metadata.CamelCaseToUnderscore(daoData.TableName)+"_dao_impl.go")
	ff, _ = filepath.Abs(daoImplFile)
	daoImplTpl, ok := metadata.LoadTpl("dao_impl")
	if !ok {
		fmt.Println("undefined template" + "dao_impl")
		return
	}
	err = RenderingTemplate(daoImplTpl, daoData, ff, true)
	if err != nil {
		fmt.Println(err)
		return
	}
	baseFile := filepath.Join(daoData.DaoPath, "db.go")
	//baseFileExist := IsExist(baseFile)
	ff, _ = filepath.Abs(baseFile)
	daoBaseTpl, ok := metadata.LoadTpl("database")
	if !ok {
		fmt.Println("undefined template" + "database")
		return
	}
	err = RenderingTemplate(daoBaseTpl, daoData, ff, true)
	if err != nil {
		fmt.Println(err)
		return
	}
	daoCustomInterfacePath := filepath.Join(daoData.DaoPath, "custom")
	customExist := IsExist(daoCustomInterfacePath)
	if !customExist {
		_ = os.MkdirAll(daoCustomInterfacePath, 0777)
	}
	return
}

func getColumnInfo(columnTypes []gorm.ColumnType, columnInfoList *[]*metadata.ColumnInfo) {
	for _, columnType := range columnTypes {
		*columnInfoList = append(*columnInfoList, &metadata.ColumnInfo{
			ColumnName: columnType.Name(),
			ColumnType: func() string {
				columnTypeName, ok := columnType.ColumnType()
				if ok {
					return columnTypeName
				}
				return ""
			}(),
			DataBaseType: columnType.DatabaseTypeName(),
			IsPrimaryKey: func() bool {
				if prime, ok := columnType.PrimaryKey(); ok {
					return prime
				}
				return false
			}(),
			Unique: func() bool {
				if unique, ok := columnType.Unique(); ok {
					return unique
				}
				return false
			}(),
			AutoIncrement: func() bool {
				if increment, ok := columnType.AutoIncrement(); ok {
					return increment
				}
				return false
			}(),
			Length: func() int64 {
				if length, ok := columnType.Length(); ok {
					return length
				}
				return 0
			}(),
			Nullable: func() bool {
				null, ok := columnType.Nullable()
				if ok {
					return null
				}
				return false
			}(),
			Comment: func() string {
				comment, ok := columnType.Comment()
				if ok {
					return comment
				}
				return ""
			}(),
			DefaultValue: func() string {
				defaultVal, ok := columnType.DefaultValue()
				if ok {
					return defaultVal
				}
				return ""
			}(),
		})
	}
}
