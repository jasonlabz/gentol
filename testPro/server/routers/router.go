package routers

import (
	"fmt"

	"github.com/gin-gonic/gin"
	knife4go "github.com/jasonlabz/knife4go"
	"github.com/jasonlabz/potato/configx"
	"github.com/jasonlabz/potato/middleware"

	_ "testPro/docs"
	"testPro/server/controller"
)

// InitApiRouter 封装路由
func InitApiRouter() *gin.Engine {
	router := gin.Default()
	serverConfig := configx.GetConfig()

	// 全局中间件，查看定义的中间价在middlewares文件夹中
	rootMiddleware(router)

	registerRootAPI(router)

	// 对路由进行分组，处理不同的分组，根据自己的需求定义即可
	staticRouter := router.Group("/server")
	staticRouter.Static("/", "application")

	serverGroup := router.Group(fmt.Sprintf("/%s", serverConfig.Application.Name))
	//debug模式下，注册swagger路由
	//knife4go: beautify swagger-ui http://ip:port/server_name/doc.html
	// knife4go: beautify swagger-ui,
	if serverConfig.Debug {
		_ = knife4go.InitSwaggerKnife(serverGroup)
	}

	apiGroup := serverGroup.Group("/api")

	// 中间件拦截器
	groupMiddleware(apiGroup,
		middleware.RecoveryLog(true), middleware.SetContext(), middleware.RequestMiddleware())

	// base api
	registerBaseAPI(serverGroup)

	// v1 group api
	v1Group := apiGroup.Group("/v1")
	registerV1GroupAPI(v1Group)

	return router
}

func rootMiddleware(r *gin.Engine, middlewares ...gin.HandlerFunc) {
	r.Use(middlewares...)
}

func groupMiddleware(g *gin.RouterGroup, middlewares ...gin.HandlerFunc) {
	g.Use(middlewares...)
}

// 注册根路由  http://ip:port/**
func registerRootAPI(router *gin.Engine) {
	router.GET("/health-check", controller.HealthCheck)
}

// 注册服務路由  http://ip:port/server_name/api/**
func registerBaseAPI(router *gin.RouterGroup) {
}

// 注册組路由 http://ip:port/server_name/api/v1/**
func registerV1GroupAPI(router *gin.RouterGroup) {
	//v1.RegisterSchedulerManagerGroup(router)
}
