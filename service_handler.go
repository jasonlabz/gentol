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

// initServiceMeta 初始化服务元数据
func initServiceMeta(serviceName string, isManager bool) *metadata.ProjectMeta {
	serviceMeta := &metadata.ProjectMeta{
		ServiceName: serviceName,
		ServicePackageName: func() string {
			if isManager {
				return "manager"
			}
			return "service"
		}(),
		ServiceStructName: func() string {
			if isManager {
				return metadata.UnderscoreToUpperCamelCase(serviceName + "_manager")
			}
			return metadata.UnderscoreToUpperCamelCase(serviceName + "_service")
		}(),
	}

	return serviceMeta
}

func handleService(service string, isManager bool) {
	serviceMeta := initServiceMeta(service, isManager)

	// 检查目录是否存在
	moduleName, serverPath := checkServiceDir(isManager)
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
func checkServiceDir(isManager bool) (string, string) {
	var moduleName string
	var found bool
	dir, _ := os.Getwd()
	modFile := filepath.Join(dir, "go.mod")
	if moduleName, found = getModuleName(modFile); !found {
		log.Fatal("Please run this command from your project directory (where go.mod is located).")
	}

	serverDir := filepath.Join(dir, "server")
	serviceDir := filepath.Join(serverDir, "service")
	managerDir := filepath.Join(serverDir, "manager")

	if isManager {
		if !IsExist(managerDir) {
			createDirectory(managerDir)
		}
		return moduleName, managerDir
	}

	if !IsExist(serviceDir) {
		createDirectory(serviceDir)
	}
	return moduleName, serviceDir
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
