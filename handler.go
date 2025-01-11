// Package main -----------------------------
// @file      : handler.go
// @author    : jasonlabz
// @contact   : 1783022886@qq.com
// @time      : 2024/11/29 22:02
// -------------------------------------------
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jasonlabz/gentol/metadata"
)

func handleNewProject(projectName string) {
	projectMeta := &metadata.ProjectMeta{
		ModulePath: projectName,
	}
	splitProjects := strings.Split(projectName, "/")
	if len(splitProjects) == 0 {
		return
	}

	for i := len(splitProjects) - 1; i >= 0; i-- {
		if len(splitProjects[i]) > 0 {
			projectMeta.ProjectName = splitProjects[i]
			break
		}
	}

	projectDir := filepath.Base(projectMeta.ProjectName)

	if IsExist(projectDir) {
		fmt.Println("[tips] project is already exist, please clear dir: ", projectDir, ", and try again")
		return
	}
	err := os.MkdirAll(projectDir, 0644)
	if err != nil {
		fmt.Println("err occured: ", err)
		return
	}

	cmdPath := filepath.Join(projectDir, "cmd")
	err = os.MkdirAll(cmdPath, 0644)
	if err != nil {
		fmt.Println("err occured: ", err)
		return
	}
	demoPath := filepath.Join(cmdPath, "demo_program")
	err = os.MkdirAll(demoPath, 0644)
	if err != nil {
		fmt.Println("err occured: ", err)
		return
	}
	mainTpl, ok := metadata.LoadTpl("main")
	if !ok {
		fmt.Println("undefined template" + "conf")
		return
	}
	err = RenderingTemplate(mainTpl, projectMeta, filepath.Join(demoPath, "main.go"), true)
	if err != nil {
		fmt.Println("err occured: ", err)
		return
	}

	confPath := filepath.Join(projectDir, "conf")
	err = os.MkdirAll(confPath, 0644)
	if err != nil {
		fmt.Println("err occured: ", err)
		return
	}
	confTpl, ok := metadata.LoadTpl("conf")
	if !ok {
		fmt.Println("undefined template" + "conf")
		return
	}
	err = RenderingTemplate(confTpl, projectMeta, filepath.Join(confPath, "application.yaml"), true)
	if err != nil {
		fmt.Println("err occured: ", err)
		return
	}
	schemaPath := filepath.Join(confPath, "schema")
	err = os.MkdirAll(schemaPath, 0644)
	if err != nil {
		fmt.Println("err occured: ", err)
		return
	}

	// bootstrap 写入
	bootstrapPath := filepath.Join(projectDir, "bootstrap")
	err = os.MkdirAll(bootstrapPath, 0644)
	if err != nil {
		fmt.Println("err occured: ", err)
		return
	}
	bootstrapTpl, ok := metadata.LoadTpl("bootstrap")
	if !ok {
		fmt.Println("undefined template" + "bootstrap")
		return
	}
	err = RenderingTemplate(bootstrapTpl, projectMeta, filepath.Join(bootstrapPath, "bootstrap.go"), true)
	if err != nil {
		fmt.Println("err occured: ", err)
		return
	}

	// common 写入
	commonPath := filepath.Join(projectDir, "common")
	err = os.MkdirAll(bootstrapPath, 0644)
	if err != nil {
		fmt.Println("err occured: ", err)
		return
	}
	constsPath := filepath.Join(commonPath, "consts")
	err = os.MkdirAll(constsPath, 0644)
	if err != nil {
		fmt.Println("err occured: ", err)
		return
	}
	constantTpl, ok := metadata.LoadTpl("constant")
	if !ok {
		fmt.Println("undefined template" + "constant")
		return
	}
	err = RenderingTemplate(constantTpl, projectMeta, filepath.Join(constsPath, "constant.go"), true)
	if err != nil {
		fmt.Println("err occured: ", err)
		return
	}

	ginxPath := filepath.Join(commonPath, "ginx")
	err = os.MkdirAll(ginxPath, 0644)
	if err != nil {
		fmt.Println("err occured: ", err)
		return
	}
	ginxTpl, ok := metadata.LoadTpl("ginx")
	if !ok {
		fmt.Println("undefined template" + "ginx")
		return
	}
	err = RenderingTemplate(ginxTpl, projectMeta, filepath.Join(ginxPath, "ginx.go"), true)
	if err != nil {
		fmt.Println("err occured: ", err)
		return
	}
	pageTpl, ok := metadata.LoadTpl("page")
	if !ok {
		fmt.Println("undefined template" + "page")
		return
	}
	err = RenderingTemplate(pageTpl, projectMeta, filepath.Join(ginxPath, "page.go"), true)
	if err != nil {
		fmt.Println("err occured: ", err)
		return
	}

	helperPath := filepath.Join(commonPath, "helper")
	err = os.MkdirAll(helperPath, 0644)
	if err != nil {
		fmt.Println("err occured: ", err)
		return
	}
	helperTpl, ok := metadata.LoadTpl("helper")
	if !ok {
		fmt.Println("undefined template" + "helper")
		return
	}
	err = RenderingTemplate(helperTpl, projectMeta, filepath.Join(helperPath, "helper.go"), true)
	if err != nil {
		fmt.Println("err occured: ", err)
		return
	}

	// docs 写入
	docsPath := filepath.Join(projectDir, "docs")
	err = os.MkdirAll(docsPath, 0644)
	if err != nil {
		fmt.Println("err occured: ", err)
		return
	}
	docsTpl, ok := metadata.LoadTpl("docs")
	if !ok {
		fmt.Println("undefined template" + "docs")
		return
	}
	err = RenderingTemplate(docsTpl, projectMeta, filepath.Join(docsPath, "docs.go"), true)
	if err != nil {
		fmt.Println("err occured: ", err)
		return
	}

	// resource 写入
	resourcePath := filepath.Join(projectDir, "global", "resource")
	err = os.MkdirAll(resourcePath, 0644)
	if err != nil {
		fmt.Println("err occured: ", err)
		return
	}
	resourceTpl, ok := metadata.LoadTpl("resource")
	if !ok {
		fmt.Println("undefined template" + "resource")
		return
	}
	err = RenderingTemplate(resourceTpl, projectMeta, filepath.Join(resourcePath, "resource.go"), true)
	if err != nil {
		fmt.Println("err occured: ", err)
		return
	}

	// server 写入
	serverPath := filepath.Join(projectDir, "server")
	err = os.MkdirAll(serverPath, 0644)
	if err != nil {
		fmt.Println("err occured: ", err)
		return
	}

	controllerPath := filepath.Join(serverPath, "controller")
	err = os.MkdirAll(controllerPath, 0644)
	if err != nil {
		fmt.Println("err occured: ", err)
		return
	}
	controllerTpl, ok := metadata.LoadTpl("controller")
	if !ok {
		fmt.Println("undefined template" + "controller")
		return
	}
	err = RenderingTemplate(controllerTpl, projectMeta, filepath.Join(controllerPath, "health_check.go"), true)
	if err != nil {
		fmt.Println("err occured: ", err)
		return
	}

	middlewarePath := filepath.Join(serverPath, "middleware")
	err = os.MkdirAll(middlewarePath, 0644)
	if err != nil {
		fmt.Println("err occured: ", err)
		return
	}
	loggerMiddlewareTpl, ok := metadata.LoadTpl("loggerMiddleware")
	if !ok {
		fmt.Println("undefined template" + "loggerMiddleware")
		return
	}
	err = RenderingTemplate(loggerMiddlewareTpl, projectMeta, filepath.Join(middlewarePath, "logger.go"), true)
	if err != nil {
		fmt.Println("err occured: ", err)
		return
	}
	contextMiddlewareTpl, ok := metadata.LoadTpl("contextMiddleware")
	if !ok {
		fmt.Println("undefined template" + "contextMiddleware")
		return
	}
	err = RenderingTemplate(contextMiddlewareTpl, projectMeta, filepath.Join(middlewarePath, "context.go"), true)
	if err != nil {
		fmt.Println("err occured: ", err)
		return
	}

	routerPath := filepath.Join(serverPath, "routers")
	err = os.MkdirAll(routerPath, 0644)
	if err != nil {
		fmt.Println("err occured: ", err)
		return
	}
	routerTpl, ok := metadata.LoadTpl("router")
	if !ok {
		fmt.Println("undefined template" + "router")
		return
	}
	err = RenderingTemplate(routerTpl, projectMeta, filepath.Join(routerPath, "router.go"), true)
	if err != nil {
		fmt.Println("err occured: ", err)
		return
	}

	servicePath := filepath.Join(serverPath, "service")
	err = os.MkdirAll(servicePath, 0644)
	if err != nil {
		fmt.Println("err occured: ", err)
		return
	}
	serviceTpl, ok := metadata.LoadTpl("service")
	if !ok {
		fmt.Println("undefined template" + "service")
		return
	}
	err = RenderingTemplate(serviceTpl, projectMeta, filepath.Join(servicePath, "health_check.go"), true)
	if err != nil {
		fmt.Println("err occured: ", err)
		return
	}

	healthCheckPath := filepath.Join(servicePath, "health_check")
	err = os.MkdirAll(healthCheckPath, 0644)
	if err != nil {
		fmt.Println("err occured: ", err)
		return
	}
	serviceImplTpl, ok := metadata.LoadTpl("serviceImpl")
	if !ok {
		fmt.Println("undefined template" + "serviceImpl")
		return
	}
	err = RenderingTemplate(serviceImplTpl, projectMeta, filepath.Join(healthCheckPath, "health_check_impl.go"), true)
	if err != nil {
		fmt.Println("err occured: ", err)
		return
	}

	dtoPath := filepath.Join(healthCheckPath, "dto")
	err = os.MkdirAll(dtoPath, 0644)
	if err != nil {
		fmt.Println("err occured: ", err)
		return
	}
	reqDTOTpl, ok := metadata.LoadTpl("reqDTO")
	if !ok {
		fmt.Println("undefined template" + "reqDTO")
		return
	}
	err = RenderingTemplate(reqDTOTpl, projectMeta, filepath.Join(dtoPath, "request.go"), true)
	if err != nil {
		fmt.Println("err occured: ", err)
		return
	}
	resDtoTpl, ok := metadata.LoadTpl("resDto")
	if !ok {
		fmt.Println("undefined template" + "resDto")
		return
	}
	err = RenderingTemplate(resDtoTpl, projectMeta, filepath.Join(dtoPath, "response.go"), true)
	if err != nil {
		fmt.Println("err occured: ", err)
		return
	}

	gomodTpl, ok := metadata.LoadTpl("gomod")
	if !ok {
		fmt.Println("undefined template" + "gomod")
		return
	}
	err = RenderingTemplate(gomodTpl, projectMeta, filepath.Join(projectDir, "go.mod"), true)
	if err != nil {
		fmt.Println("err occured: ", err)
		return
	}
	//mainTpl, ok := metadata.LoadTpl("main")
	//if !ok {
	//	fmt.Println("undefined template" + "main")
	//	return
	//}
	err = RenderingTemplate(mainTpl, projectMeta, filepath.Join(projectDir, "main.go"), true)
	if err != nil {
		fmt.Println("err occured: ", err)
		return
	}
	readmeTpl, ok := metadata.LoadTpl("readme")
	if !ok {
		fmt.Println("undefined template" + "readme")
		return
	}
	err = RenderingTemplate(readmeTpl, projectMeta, filepath.Join(projectDir, "README.md"), true)
	if err != nil {
		fmt.Println("err occured: ", err)
		return
	}

}
