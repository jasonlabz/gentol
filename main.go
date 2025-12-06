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
//  Created:   2023/8/14 1:39
//  Version:   v1.0.0

package main

import (
	"log"
	"os"
	"regexp"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		processDB()
		return
	}

	switch os.Args[1] {
	case "init", "new":
		// 项目生成
		projectName := getProjectName()
		if !isValidProjectName(projectName) {
			log.Fatal("项目名称无效，只允许字母、数字、斜杠、下划线和连字符")
		}
		handleNewProject(projectName)
	case "update":
		// 项目更新
		updateProject(getProjectName())
	case "add":
		// 增加service模板
		handleService(getServiceInfo())
	default:
		processDB()
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

// getServiceInfo 获取并验证服务名称，是否为manager（包含多个service的逻辑整合）
func getServiceInfo() (string, bool) {
	serviceName := ""
	if len(os.Args) > 2 {
		serviceName = os.Args[2]
	}
	if strings.HasSuffix(serviceName, "_manager") {
		return strings.TrimSuffix(serviceName, "_manager"), true
	}
	if strings.HasSuffix(serviceName, "_service") {
		return strings.TrimSuffix(serviceName, "_service"), false
	}
	return serviceName, false
}

// isValidProjectName 验证项目名称格式
func isValidProjectName(name string) bool {
	if len(name) == 0 {
		return false
	}
	match, _ := regexp.MatchString("^[/.a-zA-Z0-9_-]+$", name)
	return match
}
