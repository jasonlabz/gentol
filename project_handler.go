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

// handleNewProject 处理新项目创建
// 使用 clone+replace 模式（从模板仓库克隆到内存，替换后写入磁盘）
func handleNewProject(projectName string, templateRepo string, templateDir string, offline bool) {
	projectMeta := initProjectMeta(projectName)
	if projectMeta == nil {
		return
	}

	useLocalDir := templateDir != ""
	source := templateDir
	if source == "" {
		source = templateRepo // 为空时 cloneAndReplaceProject 会使用 DefaultTemplateRepoURL
	}

	if err := cloneAndReplaceProject(projectMeta.ModulePath, source, useLocalDir, offline); err != nil {
		log.Fatalf("Failed to create project: %v\n", err)
	}

	log.Println("Project created successfully!")
}

// updateProject 处理项目更新
// 使用 clone+replace 模式，覆盖同名文件，保留项目中自定义文件
func updateProject(projectName string, templateRepo string, templateDir string, offline bool) {
	currentDir, _ := os.Getwd()
	var projectDir string

	if projectName == "" {
		modFile := filepath.Join(currentDir, "go.mod")
		if moduleName, found := getModuleName(modFile); found {
			projectName = moduleName
			projectDir = currentDir
		} else {
			log.Fatal("project name is needed or go.mod not found in current directory")
		}
	} else {
		projectMeta := initProjectMeta(projectName)
		if projectMeta == nil {
			return
		}

		projectDir = filepath.Join(currentDir, projectMeta.ProjectName)
		modFile := filepath.Join(projectDir, "go.mod")
		if moduleName, found := getModuleName(modFile); found {
			projectName = moduleName
		} else {
			log.Fatalf("[%s] is not a project, because [%s] not found, use gentol init|new first",
				projectMeta.ProjectName, modFile)
		}
	}

	if !IsExist(projectDir) {
		log.Fatalf("project directory not found: %s", projectDir)
	}

	useLocalDir := templateDir != ""
	source := templateDir
	if source == "" {
		source = templateRepo
	}

	if err := updateProjectFromTemplate(projectDir, projectName, source, useLocalDir, offline); err != nil {
		log.Fatalf("Failed to update project: %v\n", err)
	}

	log.Println("Project updated successfully!")
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

	for i := len(splitProjects) - 1; i >= 0; i-- {
		if len(splitProjects[i]) > 0 {
			projectMeta.ProjectName = splitProjects[i]
			break
		}
	}

	return projectMeta
}

// createDirectory 创建目录（被 service_handler.go 使用）
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

// renderTemplate 渲染模板到文件（被 service_handler.go 使用）
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
