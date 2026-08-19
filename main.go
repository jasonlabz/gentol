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

	if isHelpFlag(os.Args[1]) {
		printTopLevelUsage()
		return
	}

	switch os.Args[1] {
	case "init", "new":
		if hasHelpFlag(os.Args[2:]) {
			printSubUsage(newUsage)
			return
		}
		// 项目生成
		projectName := getProjectName()
		if !isValidProjectName(projectName) {
			log.Fatal("项目名称无效，只允许字母、数字、斜杠、下划线和连字符")
		}
		templateRepo, templateDir, templateBranch := getTemplateFlags()
		handleNewProject(projectName, templateRepo, templateDir, templateBranch)
	case "update":
		if hasHelpFlag(os.Args[2:]) {
			printSubUsage(updateUsage)
			return
		}
		// 项目更新
		templateRepo, templateDir, templateBranch := getTemplateFlags()
		updateProject(getProjectName(), templateRepo, templateDir, templateBranch)
	case "ddl":
		if hasHelpFlag(os.Args[2:]) {
			printSubUsage(ddlUsage)
			return
		}
		// 执行DDL语句
		processDDL()
	default:
		processDB()
	}
}

// getProjectName 获取并验证项目名称（跳过 -- 开头的标志参数）
func getProjectName() string {
	projectName := ""
	if len(os.Args) > 2 {
		// 取第一个非标志参数作为项目名称
		for _, arg := range os.Args[2:] {
			if !strings.HasPrefix(arg, "--") {
				projectName = arg
				break
			}
		}
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

// getTemplateFlags 从命令行参数中解析 --template_repo、--template_dir 和 --template_branch 标志
func getTemplateFlags() (templateRepo string, templateDir string, templateBranch string) {
	args := os.Args[2:]
	for i := 0; i < len(args); i++ {
		if strings.HasPrefix(args[i], "--template_repo=") {
			templateRepo = strings.TrimPrefix(args[i], "--template_repo=")
		} else if args[i] == "--template_repo" && i+1 < len(args) {
			i++
			templateRepo = args[i]
		} else if strings.HasPrefix(args[i], "--template_dir=") {
			templateDir = strings.TrimPrefix(args[i], "--template_dir=")
		} else if args[i] == "--template_dir" && i+1 < len(args) {
			i++
			templateDir = args[i]
		} else if strings.HasPrefix(args[i], "--template_branch=") {
			templateBranch = strings.TrimPrefix(args[i], "--template_branch=")
		} else if args[i] == "--template_branch" && i+1 < len(args) {
			i++
			templateBranch = args[i]
		}
	}
	return
}
