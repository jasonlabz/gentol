// Package main
//
//   _ __ ___   __ _ _ __  _   _| |_
//  | '_ ` _ \ / _` | '_ \| | | | __|
//  | | | | | | (_| | | | | |_| | |_
//  |_| |_| |_|\__,_|_| |_|\__,_|\__|
//
//  Buddha bless, no bugs forever!
//
//  Author:    lucas
//  Email:     1783022886@qq.com
//  Created:   2025/12/6 2:16
//  Version:   v1.0.0

package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/jasonlabz/gentol/metadata"
)

// TemplateConfig 模板渲染配置
type TemplateConfig struct {
	TemplateName string
	FilePath     string
}

// initServiceMeta 初始化服务元数据
func initServiceMeta(serviceName string, isManager bool, serviceDir string) *metadata.ProjectMeta {
	serviceMeta := &metadata.ProjectMeta{
		ServiceName: serviceName,
		ServiceStructName: func() string {
			if isManager {
				return metadata.UnderscoreToUpperCamelCase(serviceName + "_manager")
			}
			return metadata.UnderscoreToUpperCamelCase(serviceName + "_service")
		}(),
	}

	// --service 指定的是相对项目根目录的完整路径
	if serviceDir != "" {
		serviceMeta.ServiceDir = "/" + serviceDir
		serviceMeta.ServicePackageName = filepath.Base(serviceDir)
	} else if isManager {
		serviceMeta.ServiceDir = "/server/manager"
		serviceMeta.ServicePackageName = "manager"
	} else {
		serviceMeta.ServiceDir = "/server/service"
		serviceMeta.ServicePackageName = "service"
	}

	return serviceMeta
}

func handleService(service string, isManager bool, serviceDir string) {
	serviceMeta := initServiceMeta(service, isManager, serviceDir)

	// 检查目录是否存在
	moduleName, serverPath := checkServiceDir(isManager, serviceDir)
	serviceMeta.ModulePath = moduleName
	// 执行各模块的创建步骤
	createSteps := []func(*metadata.ProjectMeta, string, bool) bool{
		createService,
	}

	for _, step := range createSteps {
		if !step(serviceMeta, serverPath, isManager) {
			return
		}
	}

	fmt.Println("Service added successfully!")
}

// checkServiceDir 检查service目录
func checkServiceDir(isManager bool, serviceDir string) (string, string) {
	var moduleName string
	var found bool
	dir, _ := os.Getwd()
	modFile := filepath.Join(dir, "go.mod")
	if moduleName, found = getModuleName(modFile); !found {
		log.Fatal("Please run this command from your project directory (where go.mod is located).")
	}

	var baseDir string
	// --service 指定的是相对项目根目录的完整路径，不再限定在 server/service 或 server/manager 下
	if serviceDir != "" {
		baseDir = filepath.Join(dir, serviceDir)
	} else if isManager {
		baseDir = filepath.Join(dir, "server", "manager")
	} else {
		baseDir = filepath.Join(dir, "server", "service")
	}

	if !IsExist(baseDir) {
		createDirectory(baseDir)
	}
	return moduleName, baseDir
}

func createService(serviceMeta *metadata.ProjectMeta, servicePath string, isManager bool) bool {
	// 创建service目录下的文件
	interfaceFileName := fmt.Sprintf("%s_service", serviceMeta.ServiceName)
	if isManager {
		interfaceFileName = fmt.Sprintf("%s_manager", serviceMeta.ServiceName)
	}
	serviceFiles := []TemplateConfig{
		{"add_service", filepath.Join(servicePath, fmt.Sprintf("%s.go", interfaceFileName))},
		{"add_service_impl", filepath.Join(servicePath, serviceMeta.ServiceName, fmt.Sprintf("%s_impl.go", interfaceFileName))},
		{"empty", filepath.Join(servicePath, serviceMeta.ServiceName, "body", "request.go")},
		{"empty", filepath.Join(servicePath, serviceMeta.ServiceName, "body", "response.go")},
		{"empty", filepath.Join(servicePath, serviceMeta.ServiceName, "body", "vo.go")},
		{"empty", filepath.Join(servicePath, serviceMeta.ServiceName, "body", "dto.go")},
	}
	for _, file := range serviceFiles {
		if !renderTemplate(file.TemplateName, serviceMeta, file.FilePath, false) {
			return false
		}
	}

	return true
}
