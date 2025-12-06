package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/jasonlabz/gentol/metadata"
)

// TemplateConfig 模板配置结构体
type TemplateConfig struct {
	TemplateName string
	FilePath     string
}

// PathConfig 路径配置结构体
type PathConfig struct {
	Path  string
	Files []TemplateConfig
}

func getModuleName(modFile string) (string, bool) {
	if !IsExist(modFile) {
		return "", false
	}

	file, err := os.Open(modFile)
	if err != nil {
		log.Fatal(fmt.Sprintf("open go.mod error: %s", err))
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "module ") {
			return strings.TrimPrefix(line, "module "), true
		}
	}
	return "", false
}

func updateProject(projectName string) {
	var isProjectDir bool
	currentDir, _ := os.Getwd()

	// 确定项目名称和位置
	if projectName == "" {
		// 当前目录模式
		modFile := filepath.Join(currentDir, "go.mod")
		if moduleName, found := getModuleName(modFile); found {
			projectName = moduleName
			isProjectDir = true
		} else {
			log.Fatal("project name is needed or go.mod not found")
		}
	} else {
		// 指定项目名称模式
		projectMeta := initProjectMeta(projectName)
		if projectMeta == nil {
			return
		}

		modFile := filepath.Join(currentDir, projectMeta.ProjectName, "go.mod")
		if moduleName, found := getModuleName(modFile); found {
			projectName = moduleName
		} else {
			log.Fatal(fmt.Errorf("[%s] is not a project, because [%s] not found, use gentol init|new first",
				projectMeta.ProjectName, modFile))
		}
	}

	// 切换到父目录（如果需要）
	if isProjectDir {
		parentDir := getParentPath(currentDir)
		if err := os.Chdir(parentDir); err != nil {
			log.Fatal(fmt.Errorf("chdir error: %s", err))
		}
	}

	// 初始化项目元数据
	projectMeta := initProjectMeta(projectName)
	if projectMeta == nil {
		return
	}

	// 创建项目根目录
	if !createProjectRoot(projectMeta.ProjectName, true) {
		return
	}

	// 执行各模块的创建步骤
	createSteps := []func(*metadata.ProjectMeta, bool) bool{
		createCmdStructure,
		createConfStructure,
		createBootstrap,
		createCommonStructure,
		createDocs,
		createGlobalResource,
		createServerStructure,
		createIDL,
		createScriptFile,
		createRootFiles,
	}

	for _, step := range createSteps {
		if !step(projectMeta, true) {
			return
		}
	}

	log.Println("Project updated successfully!")
}

// handleNewProject 处理新项目创建
func handleNewProject(projectName string) {
	// 初始化项目元数据
	projectMeta := initProjectMeta(projectName)
	if projectMeta == nil {
		return
	}

	// 创建项目根目录
	if !createProjectRoot(projectMeta.ProjectName, false) {
		return
	}

	// 执行各模块的创建步骤
	createSteps := []func(*metadata.ProjectMeta, bool) bool{
		createCmdStructure,
		createConfStructure,
		createBootstrap,
		createCommonStructure,
		createDocs,
		createGlobalResource,
		createServerStructure,
		createIDL,
		createScriptFile,
		createRootFiles,
	}

	for _, step := range createSteps {
		if !step(projectMeta, false) {
			return
		}
	}

	log.Println("Project created successfully!")
}

// initProjectMeta 初始化项目元数据
func initProjectMeta(projectName string) *metadata.ProjectMeta {
	projectMeta := &metadata.ProjectMeta{
		ModulePath: projectName,
	}

	splitProjects := strings.Split(projectName, "/")
	if len(splitProjects) == 0 {
		projectMeta.ModulePath = "generate_example"
		projectMeta.ProjectName = projectMeta.ModulePath
		return projectMeta
	}

	// 从路径中提取项目名
	for i := len(splitProjects) - 1; i >= 0; i-- {
		if len(splitProjects[i]) > 0 {
			projectMeta.ProjectName = splitProjects[i]
			break
		}
	}

	return projectMeta
}

// createProjectRoot 创建项目根目录
func createProjectRoot(projectName string, update bool) bool {
	projectDir := filepath.Base(projectName)

	if IsExist(projectDir) && !update {
		log.Printf("[tips] project is already exist, please clear dir: %s, and try again\n", projectDir)
		return false
	}

	if err := os.MkdirAll(projectDir, 0755); err != nil {
		log.Printf("Error creating project directory: %v\n", err)
		return false
	}

	return true
}

// createDirectory 创建目录
func createDirectory(path string) bool {
	if IsExist(path) {
		return true
	}
	if err := os.MkdirAll(path, 0755); err != nil {
		log.Printf("Error creating directory %s: %v\n", path, err)
		return false
	}
	return true
}

// renderTemplate 渲染模板到文件
func renderTemplate(templateName string, meta *metadata.ProjectMeta, filePath string, update bool) bool {
	tpl, ok := metadata.LoadTpl(templateName)
	if !ok {
		log.Printf("Undefined template: %s\n", templateName)
		return false
	}

	if err := RenderingTemplate(tpl, meta, filePath, update); err != nil {
		log.Printf("Error rendering template %s to %s: %v\n", templateName, filePath, err)
		return false
	}

	return true
}

// createCmdStructure 创建cmd目录结构
func createCmdStructure(meta *metadata.ProjectMeta, update bool) bool {
	basePath := filepath.Base(meta.ProjectName)
	cmdPath := filepath.Join(basePath, "cmd")
	demoPath := filepath.Join(cmdPath, "demo_program")

	if !createDirectory(demoPath) {
		return false
	}

	return renderTemplate("main", meta, filepath.Join(demoPath, "main.go"), update)
}

// createConfStructure 创建conf目录结构
func createConfStructure(meta *metadata.ProjectMeta, update bool) bool {
	basePath := filepath.Base(meta.ProjectName)
	confPath := filepath.Join(basePath, "conf")

	// 创建conf子目录
	confDirs := []string{
		confPath,
		filepath.Join(confPath, "log"),
		filepath.Join(confPath, "servicer"),
		filepath.Join(confPath, "schema"),
	}

	for _, dir := range confDirs {
		if !createDirectory(dir) {
			return false
		}
	}

	// 创建conf目录下的文件
	confFiles := []TemplateConfig{
		{"conf", filepath.Join(confPath, "application.yaml")},
		{"log", filepath.Join(confPath, "log", "service.yaml")},
		{"servicer", filepath.Join(confPath, "servicer", "demo.yaml")},
	}

	for _, file := range confFiles {
		if !renderTemplate(file.TemplateName, meta, file.FilePath, update) {
			return false
		}
	}

	return true
}

// createBootstrap 创建bootstrap
func createBootstrap(meta *metadata.ProjectMeta, update bool) bool {
	basePath := filepath.Base(meta.ProjectName)
	bootstrapPath := filepath.Join(basePath, "bootstrap")

	if !createDirectory(bootstrapPath) {
		return false
	}

	return renderTemplate("bootstrap", meta, filepath.Join(bootstrapPath, "bootstrap.go"), update)
}

// createCommonStructure 创建common目录结构
func createCommonStructure(meta *metadata.ProjectMeta, update bool) bool {
	basePath := filepath.Base(meta.ProjectName)
	commonPath := filepath.Join(basePath, "common")

	// 创建common子目录
	commonDirs := []string{
		commonPath,
		filepath.Join(commonPath, "consts"),
		filepath.Join(commonPath, "ginx"),
		filepath.Join(commonPath, "helper"),
	}

	for _, dir := range commonDirs {
		if !createDirectory(dir) {
			return false
		}
	}

	// 创建common目录下的文件
	commonFiles := []TemplateConfig{
		{"constant", filepath.Join(commonPath, "consts", "constant.go")},
		{"ginx", filepath.Join(commonPath, "ginx", "response.go")},
		{"page", filepath.Join(commonPath, "ginx", "page.go")},
		{"helper", filepath.Join(commonPath, "helper", "context.go")},
	}

	for _, file := range commonFiles {
		if !renderTemplate(file.TemplateName, meta, file.FilePath, update) {
			return false
		}
	}

	return true
}

// createDocs 创建docs
func createDocs(meta *metadata.ProjectMeta, update bool) bool {
	basePath := filepath.Base(meta.ProjectName)
	docsPath := filepath.Join(basePath, "docs")

	if !createDirectory(docsPath) {
		return false
	}

	return renderTemplate("docs", meta, filepath.Join(docsPath, "docs.go"), update)
}

// createGlobalResource 创建global/resource
func createGlobalResource(meta *metadata.ProjectMeta, update bool) bool {
	basePath := filepath.Base(meta.ProjectName)
	resourcePath := filepath.Join(basePath, "global", "resource")

	if !createDirectory(resourcePath) {
		return false
	}

	return renderTemplate("resource", meta, filepath.Join(resourcePath, "resource.go"), update)
}

// createIDL 创建idl/client、idl/server
func createIDL(meta *metadata.ProjectMeta, update bool) bool {
	basePath := filepath.Base(meta.ProjectName)
	idlPath := filepath.Join(basePath, "idl")
	// 创建idl子目录
	idlDirs := []string{
		idlPath,
		filepath.Join(idlPath, "client"),
		filepath.Join(idlPath, "server"),
	}

	for _, dir := range idlDirs {
		if !createDirectory(dir) {
			return false
		}
	}
	// 创建idl目录下的文件
	idlFiles := []TemplateConfig{
		{"idl_client", filepath.Join(idlPath, "client", "demo.thrift")},
		{"idl_server", filepath.Join(idlPath, "server", "demo.thrift")},
	}

	for _, file := range idlFiles {
		if !renderTemplate(file.TemplateName, meta, file.FilePath, update) {
			return false
		}
	}
	return true
}

// createScriptFile 创建script
func createScriptFile(meta *metadata.ProjectMeta, update bool) bool {
	basePath := filepath.Base(meta.ProjectName)
	scriptPath := filepath.Join(basePath, "script")

	if !createDirectory(scriptPath) {
		return false
	}
	// 创建script目录下的文件
	scriptFiles := []TemplateConfig{
		{"script_gentol", filepath.Join(scriptPath, "gentol.sh")},
		{"script_kitex", filepath.Join(scriptPath, "generate_idl.sh")},
		{"script_swag", filepath.Join(scriptPath, "swag.sh")},
		{"script_readme", filepath.Join(scriptPath, "README.md")},
	}

	for _, file := range scriptFiles {
		if !renderTemplate(file.TemplateName, meta, file.FilePath, update) {
			return false
		}
	}
	return true
}

// createServerStructure 创建server目录结构
func createServerStructure(meta *metadata.ProjectMeta, update bool) bool {
	basePath := filepath.Base(meta.ProjectName)
	serverPath := filepath.Join(basePath, "server")

	// 创建server子目录
	serverDirs := []string{
		serverPath,
		filepath.Join(serverPath, "controller"),
		filepath.Join(serverPath, "middleware"),
		filepath.Join(serverPath, "routers"),
		filepath.Join(serverPath, "service"),
		filepath.Join(serverPath, "service", "health_check"),
		filepath.Join(serverPath, "service", "health_check", "body"),
	}

	for _, dir := range serverDirs {
		if !createDirectory(dir) {
			return false
		}
	}

	// 创建server目录下的文件
	serverFiles := []TemplateConfig{
		{"controller", filepath.Join(serverPath, "controller", "health_check.go")},
		{"loggerMiddleware", filepath.Join(serverPath, "middleware", "logger.go")},
		{"contextMiddleware", filepath.Join(serverPath, "middleware", "context.go")},
		{"router", filepath.Join(serverPath, "routers", "router.go")},
		{"service", filepath.Join(serverPath, "service", "health_check.go")},
		{"serviceImpl", filepath.Join(serverPath, "service", "health_check", "health_check_impl.go")},
		{"reqDTO", filepath.Join(serverPath, "service", "health_check", "body", "request.go")},
		{"resDto", filepath.Join(serverPath, "service", "health_check", "body", "response.go")},
	}

	for _, file := range serverFiles {
		if !renderTemplate(file.TemplateName, meta, file.FilePath, update) {
			return false
		}
	}

	return true
}

// createRootFiles 创建根目录文件
func createRootFiles(meta *metadata.ProjectMeta, update bool) bool {
	basePath := filepath.Base(meta.ProjectName)

	rootFiles := []TemplateConfig{
		{"makefile", filepath.Join(basePath, "Makefile")},
		{"gomod", filepath.Join(basePath, "go.mod")},
		{"main", filepath.Join(basePath, "main.go")},
		{"readme", filepath.Join(basePath, "README.md")},
	}

	for _, file := range rootFiles {
		if !renderTemplate(file.TemplateName, meta, file.FilePath, update) {
			return false
		}
	}

	return true
}
