package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"gorm.io/gorm"

	"github.com/jasonlabz/gentol/configx"
	"github.com/jasonlabz/gentol/datasource"
	"github.com/jasonlabz/gentol/gormx"
)

func main() {
	if len(os.Args) < 2 {
		process()
		return
	}

	switch os.Args[1] {
	case "init", "new":
		projectName := getProjectName()
		if !isValidProjectName(projectName) {
			log.Fatal("项目名称无效，只允许字母、数字、斜杠、下划线和连字符")
		}
		handleNewProject(projectName)
	case "update":
		updateProject(getProjectName())
	default:
		process()
	}
}

// getProjectName 获取并验证项目名称
func getProjectName() string {
	projectName := ""
	if len(os.Args) > 2 {
		projectName = os.Args[2]
	}
	return projectName
}

// isValidProjectName 验证项目名称格式
func isValidProjectName(name string) bool {
	if len(name) == 0 {
		return false
	}
	match, _ := regexp.MatchString("^[/.a-zA-Z0-9_-]+$", name)
	return match
}

func process() {
	argHandler()
	handleDB()
}

func handleDB() {
	tableConfigs := configx.TableConfigs

	for _, dbInfo := range tableConfigs.Configs {
		processDatabaseConfig(dbInfo, tableConfigs.GoModule)
	}
}

// processDatabaseConfig 处理单个数据库配置
func processDatabaseConfig(dbInfo *configx.DBTableInfo, goModule string) {
	findModule, modelModule, daoModule := resolveModulePaths(dbInfo, goModule)

	if !findModule {
		modelModule, daoModule = findModuleByTraversal(dbInfo)
	}

	dbInfo.ModelModule = modelModule
	dbInfo.DaoModule = daoModule

	processDatabaseTables(dbInfo)
}

// resolveModulePaths 解析模块路径
func resolveModulePaths(dbInfo *configx.DBTableInfo, goModule string) (bool, string, string) {
	modPath, relativeModelPath, relativeDaoPath := getModulePathInfo(dbInfo, goModule)

	if !IsExist(modPath) {
		return false, "{{TODO:fix your module path}}/" + filepath.Base(relativeModelPath),
			"{{TODO:fix your module path}}/" + filepath.Base(relativeDaoPath)
	}

	return extractModuleFromFile(modPath, relativeModelPath, relativeDaoPath)
}

// getModulePathInfo 获取模块路径信息
func getModulePathInfo(dbInfo *configx.DBTableInfo, goModule string) (string, string, string) {
	modPath, _ := filepath.Abs(fmt.Sprintf(".%sgo.mod", string(os.PathSeparator)))
	relativeModelPath := dbInfo.ModelPath
	relativeDaoPath := dbInfo.DaoPath

	if goModule == "" {
		return modPath, relativeModelPath, relativeDaoPath
	}

	// 处理模型路径
	if filepath.IsAbs(dbInfo.ModelPath) {
		if modulePath := extractModulePath(dbInfo.ModelPath, goModule); modulePath != "" {
			modPath = modulePath
			relativeModelPath = normalizePath(dbInfo.ModelPath[strings.LastIndex(dbInfo.ModelPath, goModule)+len(goModule)+1:])
		}
	}

	// 处理DAO路径
	if filepath.IsAbs(dbInfo.DaoPath) {
		if modulePath := extractModulePath(dbInfo.DaoPath, goModule); modulePath != "" {
			modPath = modulePath
			relativeDaoPath = normalizePath(dbInfo.DaoPath[strings.LastIndex(dbInfo.DaoPath, goModule)+len(goModule)+1:])
		}
	}

	return modPath, relativeModelPath, relativeDaoPath
}

// extractModulePath 提取模块路径
func extractModulePath(absPath, goModule string) string {
	lastIndex := strings.LastIndex(absPath, goModule)
	if lastIndex == -1 {
		return ""
	}
	return filepath.Join(absPath[:lastIndex+len(goModule)], "go.mod")
}

// normalizePath 规范化路径
func normalizePath(path string) string {
	return strings.ReplaceAll(path, string(os.PathSeparator), "/")
}

// extractModuleFromFile 从go.mod文件提取模块信息
func extractModuleFromFile(modPath, relativeModelPath, relativeDaoPath string) (bool, string, string) {
	file, err := os.Open(modPath)
	if err != nil {
		return false, "{{TODO:fix your module path}}/" + filepath.Base(relativeModelPath),
			"{{TODO:fix your module path}}/" + filepath.Base(relativeDaoPath)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "module ") {
			moduleName := strings.TrimPrefix(line, "module ")

			// 跳过gentol自身的模块
			if strings.Contains(moduleName, "github.com/jasonlabz/gentol") {
				break
			}

			modelModule := buildModulePath(moduleName, relativeModelPath)
			daoModule := buildModulePath(moduleName, relativeDaoPath)
			return true, modelModule, daoModule
		}
	}

	return false, "{{TODO:fix your module path}}/" + filepath.Base(relativeModelPath),
		"{{TODO:fix your module path}}/" + filepath.Base(relativeDaoPath)
}

// buildModulePath 构建完整的模块路径
func buildModulePath(moduleName, relativePath string) string {
	normalizedPath := strings.TrimLeft(normalizePath(relativePath), "/")
	return fmt.Sprintf("%s/%s", moduleName, normalizedPath)
}

// findModuleByTraversal 通过遍历目录查找模块
func findModuleByTraversal(dbInfo *configx.DBTableInfo) (string, string) {
	modelAbsPath := getAbsolutePath(dbInfo.ModelPath)
	daoAbsPath := getAbsolutePath(dbInfo.DaoPath)

	searchPath := modelAbsPath
	for searchPath != "" {
		modPath := filepath.Join(searchPath, "go.mod")

		if IsExist(modPath) {
			return extractModulesFromFoundPath(modPath, modelAbsPath, daoAbsPath, searchPath)
		}

		searchPath = getParentPath(searchPath)
	}

	return "{{TODO:fix your module path}}/" + filepath.Base(modelAbsPath),
		"{{TODO:fix your module path}}/" + filepath.Base(daoAbsPath)
}

// getAbsolutePath 获取绝对路径
func getAbsolutePath(path string) string {
	if filepath.IsAbs(path) {
		return path
	}

	// 清理路径前缀
	cleanPath := strings.TrimPrefix(path, "."+string(os.PathSeparator))
	cleanPath = strings.TrimPrefix(cleanPath, "./")

	abs, err := filepath.Abs(cleanPath)
	if err != nil {
		return path
	}
	return abs
}

// getParentPath 获取父路径
func getParentPath(path string) string {
	lastIndex := strings.LastIndex(path, string(os.PathSeparator))
	if lastIndex == -1 {
		return ""
	}
	return path[:lastIndex]
}

// extractModulesFromFoundPath 从找到的路径提取模块信息
func extractModulesFromFoundPath(modPath, modelAbsPath, daoAbsPath, basePath string) (string, string) {
	file, err := os.Open(modPath)
	if err != nil {
		return "{{TODO:fix your module path}}/" + filepath.Base(modelAbsPath),
			"{{TODO:fix your module path}}/" + filepath.Base(daoAbsPath)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "module ") {
			moduleName := strings.TrimPrefix(line, "module ")

			relativeModelPath := getRelativePath(modelAbsPath, basePath)
			relativeDaoPath := getRelativePath(daoAbsPath, basePath)

			modelModule := buildModulePath(moduleName, relativeModelPath)
			daoModule := buildModulePath(moduleName, relativeDaoPath)

			return modelModule, daoModule
		}
	}

	return "{{TODO:fix your module path}}/" + filepath.Base(modelAbsPath),
		"{{TODO:fix your module path}}/" + filepath.Base(daoAbsPath)
}

// getRelativePath 获取相对路径
func getRelativePath(absPath, basePath string) string {
	return strings.ReplaceAll(strings.TrimPrefix(absPath, basePath), string(os.PathSeparator), "/")
}

// processDatabaseTables 处理数据库表
func processDatabaseTables(dbInfo *configx.DBTableInfo) {
	db := createDBConnection(dbInfo)
	ds := getDataSource(gormx.DBType(dbInfo.DBType))

	tableMap := buildTableMap(dbInfo, ds, db)
	processTables(dbInfo, db, tableMap)
}

// createDBConnection 创建数据库连接
func createDBConnection(dbInfo *configx.DBTableInfo) *gorm.DB {
	dbConfig := &gormx.Config{
		DBName: dbInfo.DBName,
		DSN:    dbInfo.DSN,
		DBType: gormx.DBType(dbInfo.DBType),
	}

	db, err := gormx.LoadDBInstance(dbConfig)
	if err != nil {
		panic(err)
	}
	return db
}

// getDataSource 获取数据源
func getDataSource(dbType gormx.DBType) *datasource.DS {
	ds, err := datasource.GetDS(dbType)
	if err != nil {
		panic(err)
	}
	return ds
}

// buildTableMap 构建表映射
func buildTableMap(dbInfo *configx.DBTableInfo, ds *datasource.DS, db *gorm.DB) map[string]map[string]bool {
	tableMap := make(map[string]map[string]bool)

	if len(dbInfo.Tables) == 0 {
		dbInfo.Tables = []*configx.TableInfo{
			{SchemaName: "", TableList: []string{}},
		}
	}

	for _, tableInfo := range dbInfo.Tables {
		schemaName := strings.Trim(tableInfo.SchemaName, "\"")
		tableMap = mergeTableInfo(tableMap, schemaName, tableInfo, ds, db, dbInfo.DBName)
	}

	return tableMap
}

// mergeTableInfo 合并表信息
func mergeTableInfo(tableMap map[string]map[string]bool, schemaName string, tableInfo *configx.TableInfo, ds *datasource.DS, db *gorm.DB, dbName string) map[string]map[string]bool {
	if len(tableInfo.TableList) == 0 {
		return mergeAllTables(tableMap, schemaName, ds, db, dbName)
	}
	return mergeSpecificTables(tableMap, schemaName, tableInfo.TableList)
}

// mergeAllTables 合并所有表
func mergeAllTables(tableMap map[string]map[string]bool, schemaName string, ds *datasource.DS, db *gorm.DB, dbName string) map[string]map[string]bool {
	dbTableMap, err := ds.GetTablesUnderDB(context.TODO(), dbName)
	if err != nil {
		panic(err)
	}

	for schema, dbMeta := range dbTableMap {
		if schemaName != "" && schema != schemaName {
			continue
		}

		for _, table := range dbMeta.TableInfoList {
			addTableToMap(tableMap, schema, table.TableName)
		}
	}
	return tableMap
}

// mergeSpecificTables 合并指定表
func mergeSpecificTables(tableMap map[string]map[string]bool, schemaName string, tableList []string) map[string]map[string]bool {
	for _, tableName := range tableList {
		tableName = strings.Trim(tableName, "\"")
		addTableToMap(tableMap, schemaName, tableName)
	}
	return tableMap
}

// addTableToMap 添加表到映射
func addTableToMap(tableMap map[string]map[string]bool, schema, tableName string) {
	if _, exists := tableMap[schema]; !exists {
		tableMap[schema] = make(map[string]bool)
	}
	tableMap[schema][tableName] = true
}

// processTables 处理表
func processTables(dbInfo *configx.DBTableInfo, db *gorm.DB, tableMap map[string]map[string]bool) {
	for schema, tables := range tableMap {
		for tableName := range tables {
			processSingleTable(dbInfo, db, schema, tableName)
		}
	}
}

// processSingleTable 处理单个表
func processSingleTable(dbInfo *configx.DBTableInfo, db *gorm.DB, schema, tableName string) {
	fullTableName := buildFullTableName(schema, tableName)

	columnTypes, err := db.Migrator().ColumnTypes(fullTableName)
	if err != nil {
		log.Printf("获取表 %s 列信息失败: %v", fullTableName, err)
		return
	}
	indexes, getErr := db.Migrator().GetIndexes(fullTableName)
	if getErr != nil {
		log.Println(getErr)
	}
	WriteModel(dbInfo, schema, tableName, columnTypes, indexes)

	if !dbInfo.OnlyModel {
		WriteDao(dbInfo, schema, tableName, columnTypes)
	}
}

// buildFullTableName 构建完整表名
func buildFullTableName(schema, tableName string) string {
	if schema == "" {
		return tableName
	}
	return fmt.Sprintf("%s.%s", schema, tableName)
}
