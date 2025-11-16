// Package metadata -----------------------------
// @file      : project.go
// @author    : jasonlabz
// @contact   : 1783022886@qq.com
// @time      : 2024/11/29 22:08
// -------------------------------------------
package metadata

type ProjectMeta struct {
	ModulePath  string
	ProjectName string
}

func (p *ProjectMeta) GenRenderData() map[string]any {
	result := map[string]any{
		"ModulePath":  p.ModulePath,
		"ProjectName": p.ProjectName,
	}
	return result
}

const Bootstrap = `package bootstrap

import (
	"context"
	"os"
	"path/filepath"

	"github.com/jasonlabz/potato/configx"
	"github.com/jasonlabz/potato/configx/file"
	"github.com/jasonlabz/potato/cryptox"
	"github.com/jasonlabz/potato/cryptox/aes"
	"github.com/jasonlabz/potato/cryptox/des"
	"github.com/jasonlabz/potato/gormx"
	"github.com/jasonlabz/potato/httpx"
	"github.com/jasonlabz/potato/log"
	"github.com/jasonlabz/potato/utils"

	"{{.ModulePath}}/global/resource"
)

func MustInit(ctx context.Context) {
	// 初始化配置文件
	initConfig(ctx)
	// 初始化日志对象
	initLogger(ctx)
	// 初始化全局变量
	initResource(ctx)
	// 初始化加解秘钥
	initCrypto(ctx)
	// 初始化DB
	initDB(ctx)
	// 初始化客户端信息
	initServicer(ctx)
}

func initLogger(_ context.Context) {
	resource.Logger = log.GetLogger()
}

func initResource(_ context.Context) {
	resource.Username = func() string {
		user := os.Getenv("AUTH_USER")
		if user != "" {
			return user
		}
		return "admin"
	}()
	resource.Password = func() string {
		passwd := os.Getenv("AUTH_PASSWD")
		if passwd != "" {
			return passwd
		}
		return "admin"
	}()
}

func initCrypto(_ context.Context) {
	cryptoConfigs := configx.GetConfig().Crypto
	for _, conf := range cryptoConfigs {
		if conf.Key == "" {
			continue
		}
		switch conf.Type {
		case cryptox.CryptoTypeAES:
			aes.SetAESCrypto(aes.NewAESCrypto([]byte(conf.Key)))
		case cryptox.CryptoTypeDES:
			des.SetDESCrypto(des.NewDESCrypto([]byte(conf.Key)))
		}
	}
}

func initDB(_ context.Context) {
	dbConf := configx.GetConfig().Database
	if !dbConf.Enable {
		return
	}
	gormConfig := &gormx.Config{}
	err := utils.CopyStruct(dbConf, gormConfig)
	if err != nil {
		panic(err)
	}
	gormConfig.DBName = gormx.DefaultDBNameMaster
	_, err = gormx.InitConfig(gormConfig)
	if err != nil {
		panic(err)
	}
	// dao.SetGormDB(db)
}

func initConfig(_ context.Context) {
	filePaths, err := utils.ListDir("conf", ".yaml")
	if err != nil {
		filePaths = []string{filepath.Join("conf", "core.yaml")}
	}
	for _, filePath := range filePaths {
		provider, err := file.NewConfigProvider(filePath)
		if err != nil {
			continue
		}
		configx.AddProviders(filePath, provider)
	}
}

func initServicer(_ context.Context) {
	filePaths, _ := utils.ListDir(filepath.Join("conf", "servicer"), ".yaml")
	for _, filePath := range filePaths {
		info := &httpx.ServerInfo{}
		err := configx.ParseConfigByViper(filePath, info)
		if err != nil {
			continue
		}
		service := filepath.Base(filePath)
		if info.Name != "" {
			service = info.Name
		}
		httpx.Store(service, info)
	}
}
`

const Main = `package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
	"github.com/jasonlabz/potato/configx"
	"github.com/jasonlabz/potato/ginmetrics"

	"{{.ModulePath}}/bootstrap"
	"{{.ModulePath}}/global/resource"
	"{{.ModulePath}}/server/routers"
)

// @title		    TODO: ***********服务
// @version		    1.0
// @description	    TODO: 旨在***********
// @host			TODO: localhost:port
// @contact.name	TODO: your name
// @contact.url	    TODO: http://www.*****.io/support
// @contact.email	TODO: mail_name@qq.com
// @BasePath		TODO: /base_path
func main() {
	// context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// bootstrap init
	bootstrap.MustInit(ctx)

	// gin mode
	serverMode := gin.ReleaseMode
	serverConfig := configx.GetConfig()
	if serverConfig.IsDebugMode() {
		serverMode = gin.DebugMode
	}
	gin.SetMode(serverMode)

	r := routers.InitApiRouter()

	prometheusConf := serverConfig.GetPrometheusConfig()
	if prometheusConf.Enable {
		// get global Monitor object
		m := ginmetrics.GetMonitor()

		// +optional set metric path, default /debug/metrics
		m.SetMetricPath(prometheusConf.Path)
		// +optional set slow time, default 5s
		m.SetSlowTime(10)
		// +optional set request duration, default {0.1, 0.3, 1.2, 5, 10}
		// used to p95, p99
		m.SetDuration([]float64{0.1, 0.3, 1.2, 5, 10})

		// set middleware for gin
		m.Use(r)
	}

	pprofConf := serverConfig.GetPProfConfig()
	if pprofConf.Enable {
		r.GET("/debug/pprof/*any", gin.WrapH(http.DefaultServeMux))

		go func() {
			if err := http.ListenAndServe(fmt.Sprintf(":%d", pprofConf.Port), nil); err != nil {
				log.Fatalf("pprof server failed: %v", err)
			}
		}()
	}

	if serverConfig.Application.FileServer {
		go func() {
			fileServer(serverConfig.Application.Port + 1)
		}()
	}

	// start program
	srv := startServer(r, serverConfig.Application.Port)

	// receive quit signal, ready to exit
	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	<-quit
	log.Println("Shutdown Server ...")

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server Shutdown:", err)
	}
	log.Println("Server exiting")
}

// startServer 自定义http配置
func startServer(router *gin.Engine, port int) *http.Server {
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: router,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	return srv
}

// fileServer 文件服务
func fileServer(port int) {
	// 创建 HTTP 服务器
	mux := http.NewServeMux()
	filePath, _ := os.Getwd()
	mux.Handle("/", http.FileServer(http.Dir(filePath)))
	// 使用基本认证保护文件下载路由
	authMux := basicAuth(mux)

	// 启动 HTTP 服务器
	// log.Printf("Starting file server at :%d", config.GetConfig().Application.Port+1)
	err := http.ListenAndServe(fmt.Sprintf(":%d", port), authMux)
	if err != nil {
		log.Fatalf("file server listen: %s\n", err)
	}
	return
}

// basicAuth 认证检查
func basicAuth(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, pass, ok := r.BasicAuth()
		if !ok || user != resource.Username || pass != resource.Password {
			w.Header().Set("WWW-Authenticate", ` + "`Basic realm" + `="Restricted"` + "`)" + `
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		handler.ServeHTTP(w, r)
	})
}
`

const Router = `package routers

import (
	"fmt"

	"github.com/gin-gonic/gin"
	knife4go "github.com/jasonlabz/knife4go"
	"github.com/jasonlabz/potato/configx"
	"github.com/jasonlabz/potato/middleware"

	_ "{{.ModulePath}}/docs"
	"{{.ModulePath}}/server/controller"
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

	serverGroup := router.Group(fmt.Sprintf("/%s", serverConfig.GetName()))
	// debug模式下，注册swagger路由
	// knife4go: beautify swagger-ui http://ip:port/server_name/doc.html
	// knife4go: beautify swagger-ui,
	if serverConfig.IsDebugMode() {
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
func registerBaseAPI(router *gin.RouterGroup) {}

// 注册組路由 http://ip:port/server_name/api/v1/**
func registerV1GroupAPI(router *gin.RouterGroup) {
	// v1.RegisterSchedulerManagerGroup(router)
}
`

const LoggerMiddleware = `package middleware

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jasonlabz/potato/consts"
	"github.com/jasonlabz/potato/log"
	"github.com/jasonlabz/potato/utils"
)

const (
	requestBodyMaxLen = 204800
)

type BodyLog struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (bl BodyLog) Header() http.Header {
	return bl.ResponseWriter.Header()
}

func (bl BodyLog) Write(b []byte) (int, error) {
	bl.body.Write(b)
	return bl.ResponseWriter.Write(b)
}

func (bl BodyLog) WriteHeader(statusCode int) {
	bl.ResponseWriter.WriteHeader(statusCode)
}

func RequestMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		traceID := utils.StringValue(c.Value(consts.ContextTraceID))
		if traceID != "" {
			c.Writer.Header().Set(consts.HeaderRequestID, traceID)
		}

		var requestBodyBytes []byte
		if c.Request.Body != nil {
			requestBodyBytes, _ = io.ReadAll(c.Request.Body)
		}
		c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBodyBytes))
		bodyLog := &BodyLog{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
		c.Writer = bodyLog

		start := time.Now() // Start timer
		log.GetLogger().Info(c, "	[GIN] request",
			log.String("proto", c.Request.Proto),
			log.String("client_ip", c.ClientIP()),
			log.Int64("content_length", c.Request.ContentLength),
			log.String("agent", c.Request.UserAgent()),
			log.String("request_body", string(logBytes(requestBodyBytes, requestBodyMaxLen))),
			log.String("method", c.Request.Method),
			log.String("path", c.Request.URL.Path))

		c.Next()

		log.GetLogger().Info(c, "	[GIN] response",
			log.Int("status_code", c.Writer.Status()),
			log.String("error_message", c.Errors.ByType(gin.ErrorTypePrivate).String()),
			log.String("response_body", string(logBytes(bodyLog.body.Bytes(), requestBodyMaxLen))),
			log.String("path", c.Request.URL.Path),
			log.String("cost", fmt.Sprintf("%dms", time.Since(start).Milliseconds())))
	}
}

func logBytes(src []byte, maxLen int) []byte {
	srcLen := len(src)
	length := srcLen
	if maxLen > 0 && srcLen > maxLen {
		length = maxLen
	}
	requestBodyLogBytes := make([]byte, length)
	copy(requestBodyLogBytes, src)
	if length < srcLen {
		requestBodyLogBytes = append(requestBodyLogBytes, []byte(" ......")...)
	}
	return requestBodyLogBytes
}
`

const ContextMiddleware = `package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/jasonlabz/potato/consts"
)

type Options struct {
	headerMap      map[string]string
	customFieldMap map[string]func(ctx *gin.Context) string
}

type Option func(options *Options)

func WithHeaderField(headerMap map[string]string) Option {
	return func(options *Options) {
		options.headerMap = headerMap
	}
}

func WithCustomField(customFieldMap map[string]func(ctx *gin.Context) string) Option {
	return func(options *Options) {
		options.customFieldMap = customFieldMap
	}
}

func SetContextMiddleware(opts ...Option) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var options = &Options{}
		for _, opt := range opts {
			opt(options)
		}

		for headerKey, contextKey := range options.headerMap {
			if headerKey == "" || contextKey == "" {
				continue
			}
			value := ctx.Request.Header.Get(headerKey)
			ctx.Set(contextKey, value)
		}

		for contextKey, handler := range options.customFieldMap {
			value := handler(ctx)
			ctx.Set(contextKey, value)
		}

		traceID := ctx.Request.Header.Get(consts.HeaderRequestID)
		if traceID == "" {
			traceID = strings.ReplaceAll(uuid.New().String(), consts.SignDash, consts.EmptyString)
		}
		userID := ctx.Request.Header.Get(consts.HeaderUserID)
		authorization := ctx.Request.Header.Get(consts.HeaderAuthorization)
		remote := ctx.ClientIP()

		ctx.Set(consts.ContextToken, authorization)
		ctx.Set(consts.ContextUserID, userID)
		ctx.Set(consts.ContextTraceID, traceID)
		ctx.Set(consts.ContextClientAddr, remote)

		ctx.Next()
	}
}
`

const Service = `package service

import "context"

type HealthCheckService interface {
	DoCheck(ctx context.Context) string
}
`

const ServiceImpl = `package health_check

import (
	"context"
	"sync"

	"{{.ModulePath}}/server/service"
)

var svc *Service
var once sync.Once

func GetService() service.HealthCheckService {
	if svc != nil {
		return svc
	}
	once.Do(func() {
		svc = &Service{}
	})

	return svc
}

type Service struct {
}

func (s Service) DoCheck(ctx context.Context) string {
	return "success"
}

`

const Controller = `package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/jasonlabz/potato/consts"

	base "{{.ModulePath}}/common/ginx"
	"{{.ModulePath}}/server/service/health_check"
)

// HealthCheck 健康检查
//
//	@Summary	健康检查
//	@Tags		健康检查
//	@Accept		json
//	@Produce	json
//	@Router		/health-check [get]
func HealthCheck(c *gin.Context) {
	status := health_check.GetService().DoCheck(c)
	base.JsonResult(c, consts.APIVersionV1, status, nil)
}
`

const ReqDTO = `package dto

type HealthCheckReqDto struct {
	FieldName string` + " `json:" + `"field_name"` + "` // 属性名" + `
}

`

const ResDto = `package dto

type HealthCheckResDto struct {
	FieldName string` + " `json:" + `"field_name"` + "` // 属性名" + `
}
`

const Ginx = `package ginx

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	errors2 "github.com/jasonlabz/potato/errors"
	"github.com/jasonlabz/potato/log"

	"{{.ModulePath}}/global/resource"
)

// Response 响应结构体
type Response struct {
	Code        int    ` + " `json:" + `"code"` + "` // 错误码" + `
	Message     string ` + " `json:" + `"message,omitempty"` + "` // 错误信息" + `
	ErrTrace    string ` + " `json:" + `"err_trace,omitempty"` + "` // 错误追踪链路信息" + `
	Version     string ` + " `json:" + `"version"` + "` // 版本信息" + `
	CurrentTime string ` + " `json:" + `"current_time"` + "` // 接口返回时间（当前时间）" + `
	Data        any    ` + " `json:" + `"data,omitempty"` + "` //返回数据" + `
}

type ResponseWithPagination struct {
	Response
	Pagination *Pagination ` + " `json:" + `"pagination,omitempty"` + "`" + `
}

// FileDownloadConfig 文件下载配置
type FileDownloadConfig struct {
	Filename    string    // 下载文件名
	Preview     bool      // 是否开启预览模式
	ContentType string    // 内容类型，默认为 application/octet-stream
	Content     []byte    // 文件内容
	Reader      io.Reader // 文件读取器，与 Content 二选一
	Filepath    string    // 文件路径，与 Content 和 Reader 互斥
	Disposition string    // 内容处置类型，默认为 attachment
	BufferSize  int       // 缓冲区大小，默认为 4096
	DeleteAfter bool      // 下载后是否删除文件（仅对 Filepath 有效）
}

// ResponseOK 返回正确结果及数据
func ResponseOK(c *gin.Context, version string, data any) {
	c.JSON(prepareResponse(c, version, data, nil))
}

// ResponseErr 返回错误
func ResponseErr(c *gin.Context, version string, err error) {
	c.JSON(prepareResponse(c, version, nil, err))
}

// JsonResult 返回结果Json
func JsonResult(c *gin.Context, version string, data any, err error) {
	c.JSON(prepareResponse(c, version, data, err))
}

// FileResult 返回文件流下载
func FileResult(c *gin.Context, version string, config *FileDownloadConfig) {
	handleFileDownload(c, version, config)
}

// FileResultWithError 返回文件流下载，支持错误处理
func FileResultWithError(c *gin.Context, version string, config *FileDownloadConfig, err error) {
	if err != nil {
		ResponseErr(c, version, err)
		return
	}
	FileResult(c, version, config)
}

// PaginationResult 返回结果Json带分页
func PaginationResult(c *gin.Context, version string, data any, err error, pagination *Pagination) {
	c.JSON(prepareResponseWithPagination(c, version, data, err, pagination))
}

// PureJsonResult 返回结果PureJson
func PureJsonResult(c *gin.Context, version string, data any, err error) {
	c.PureJSON(prepareResponse(c, version, data, err))
}

// prepareResponse 准备响应信息
func prepareResponse(c *gin.Context, version string, data any, err error) (int, *Response) {
	// 格式化返回数据，非数组及切片时，转为切片
	data = handleData(data)
	code := http.StatusOK
	var errCode int
	var errMessage string
	var errTrace string

	if err != nil {
		var ex *errors2.Error
		if errors.As(err, &ex) {
			errCode = ex.Code()
			errMessage = ex.Message()
			errTrace = ex.Error()
		} else {
			code = http.StatusInternalServerError
			errMessage = err.Error()
			errTrace = err.Error()
		}
		resource.Logger.Error(c, "        "+errTrace,
			log.Int("err_code", errCode), log.String("err_message", errMessage))
	}
	// 组装响应结果
	resp := &Response{
		Code:        errCode,
		Message:     errMessage,
		ErrTrace:    errTrace,
		Version:     version,
		Data:        data,
		CurrentTime: time.Now().Format(time.DateTime),
	}
	return code, resp
}

// prepareResponseWithPagination 准备响应信息
func prepareResponseWithPagination(c *gin.Context, version string,
	data any, err error, pagination *Pagination) (int, *ResponseWithPagination) {
	code, resp := prepareResponse(c, version, data, err)
	respWithPagination := &ResponseWithPagination{
		Response:   *resp,
		Pagination: pagination,
	}

	return code, respWithPagination
}

// handleData 格式化返回数据，非数组及切片时，转为切片
func handleData(data any) any {
	v := reflect.ValueOf(data)
	if !v.IsValid() || v.Kind() == reflect.Ptr && v.IsNil() {
		return make([]any, 0)
	}
	if v.Kind() == reflect.Slice || v.Kind() == reflect.Array {
		return data
	}
	return []any{data}
}

// handleFileDownload 处理文件下载
func handleFileDownload(c *gin.Context, version string, config *FileDownloadConfig) {
	if config == nil {
		ResponseErr(c, version, errors.New("file download config is nil"))
		return
	}

	// 设置默认值
	if config.ContentType == "" {
		config.ContentType = "application/octet-stream"
	}
	if config.Disposition == "" {
		config.Disposition = "attachment"
	}
	if config.BufferSize <= 0 {
		config.BufferSize = 4096
	}
	if config.Preview {
		config.Disposition = "inline"
	}

	// 设置响应头
	filename := getDownloadFilename(config.Filename)
	c.Header("Content-Type", config.ContentType)
	c.Header("Content-Disposition", config.Disposition+"; filename="+filename)
	c.Header("Content-Transfer-Encoding", "binary")
	c.Header("Cache-Control", "no-cache")

	// 处理不同的文件来源
	if config.Filepath != "" {
		handleFileDownloadFromPath(c, version, config)
	} else if config.Reader != nil {
		handleFileDownloadFromReader(c, version, config)
	} else if config.Content != nil {
		handleFileDownloadFromContent(c, version, config)
	}

	ResponseErr(c, version, errors.New("no file content provided"))
}

// handleFileDownloadFromPath 从文件路径下载
func handleFileDownloadFromPath(c *gin.Context, version string, config *FileDownloadConfig) {
	// 检查文件是否存在
	if _, err := os.Stat(config.Filepath); errors.Is(err, fs.ErrNotExist) {
		ResponseErr(c, version, errors.New("file not found: "+config.Filepath))
		return
	}

	// 打开文件
	file, err := os.Open(config.Filepath)
	if err != nil {
		ResponseErr(c, version, fmt.Errorf("open file error: %w", err))
		return
	}
	defer file.Close()

	// 获取文件信息
	fileInfo, err := file.Stat()
	if err != nil {
		ResponseErr(c, version, fmt.Errorf("get file info error: %w", err))
		return
	}

	// 设置文件大小
	c.Header("Content-Length", strconv.FormatInt(fileInfo.Size(), 10))

	// 如果文件名为空，使用原文件名
	if config.Filename == "" {
		config.Filename = filepath.Base(config.Filepath)
	}

	// 流式传输文件
	_, err = io.CopyBuffer(c.Writer, file, make([]byte, config.BufferSize))
	if err != nil {
		ResponseErr(c, version, fmt.Errorf("copy file error: %w", err))
		return
	}

	// 下载后删除文件
	if config.DeleteAfter {
		if err = os.Remove(config.Filepath); err != nil {
			resource.Logger.Error(c, "failed to delete file after download: "+err.Error())
		}
	}

	return
}

// handleFileDownloadFromReader 从 Reader 下载
func handleFileDownloadFromReader(c *gin.Context, version string, config *FileDownloadConfig) {
	// 流式传输
	if _, err := io.CopyBuffer(c.Writer, config.Reader, make([]byte, config.BufferSize)); err != nil {
		ResponseErr(c, version, fmt.Errorf("download file error: %w", err))
	}
	return
}

// handleFileDownloadFromContent 从字节内容下载
func handleFileDownloadFromContent(c *gin.Context, version string, config *FileDownloadConfig) {
	// 设置内容长度
	c.Header("Content-Length", strconv.Itoa(len(config.Content)))

	// 直接写入内容
	if _, err := c.Writer.Write(config.Content); err != nil {
		ResponseErr(c, version, fmt.Errorf("download file error: %w", err))
	}
	return
}

// getDownloadFilename 处理下载文件名，确保浏览器兼容
func getDownloadFilename(filename string) string {
	if filename == "" {
		return "download"
	}

	// 对文件名进行编码，确保特殊字符正确处理
	return strings.ReplaceAll(strconv.Quote(filename), "\"", "")
}

// SimpleFileDownload 简化版文件下载（快速使用）
func SimpleFileDownload(c *gin.Context, version, filePath string, fileName string) {
	config := &FileDownloadConfig{
		Filepath:    filePath,
		Filename:    fileName,
		ContentType: getContentType(filePath),
	}
	FileResult(c, version, config)
}

// getContentType 根据文件扩展名获取内容类型
func getContentType(fileName string) string {
	if fileName == "" {
		return "application/octet-stream"
	}

	// 首先尝试标准库的检测
	contentType := mime.TypeByExtension(filepath.Ext(fileName))
	if contentType != "" {
		return contentType
	}

	// 标准库没有的类型，使用自定义映射
	ext := strings.ToLower(strings.TrimPrefix(filepath.Ext(fileName), "."))
	switch ext {
	case ".pdf":
		return "application/pdf"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".gif":
		return "image/gif"
	case ".txt":
		return "text/plain"
	case ".html", ".htm":
		return "text/html"
	case ".json":
		return "application/json"
	case ".xml":
		return "application/xml"
	case ".csv":
		return "text/csv"
	case ".xlsx":
		return "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	case ".xls":
		return "application/vnd.ms-excel"
	case ".docx":
		return "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	case ".doc":
		return "application/msword"
	case ".zip":
		return "application/zip"
	case ".rar":
		return "application/x-rar-compressed"
	default:
		return "application/octet-stream"
	}
}
`

const Page = `package base

import "math"

// Pagination 分页结构体（该分页只适合数据量很少的情况）
type Pagination struct {
	Page      int64 ` + " `json:" + `"page"` + "` // 当前页" + `
	PageSize  int64 ` + " `json:" + `"page_size"` + "` // 每页多少条记录" + `
	PageCount int64 ` + " `json:" + `"page_count"` + "` // 一共多少页" + `
	Total     int64 ` + " `json:" + `"total"` + "` // 一共多少条记录" + `
}

func (p *Pagination) GetPageCount() {
	p.PageCount = int64(math.Ceil(float64(p.Total) / float64(p.PageSize)))
	return
}

func (p *Pagination) GetOffset() (offset int64) {
	offset = (p.Page - 1) * p.PageSize
	return
}`

const Constant = `
package consts

import "os"

const APIVersionV1 = "v1"

`

const Readme = `# 工具介绍
### 1、 gorm gen使用` +
	"```shell" + `
## install gen tool (should be installed to ~/go/bin, make sure ~/go/bin is in your path.
## go version < 1.17
$ go get -u github.com/smallnest/gen

## go version == 1.17
$ go install github.com/smallnest/gen@v0.9.29

## generate code based on the sqlite database (project will be contained within the ./example dir)
$ gen --sqltype=postgres  --connstr "host=localhost user=postgres password=halojeff dbname=postgres port=8432 sslmode=disable"  --database postgres --table user --out ./dal --json --gorm --guregu --run-gofmt --json-fmt=snake --overwrite
` + "```" + `
### 1、 gentol使用
` + "```shell" + `
## install gentol
$ go install github.com/jasonlabz/gentol@master

## generate code based on the sqlite database (project will be contained within the ./example dir)
$ gentol --db_type="postgres" --dsn="user=postgres password=XXXXX host=127.0.0.1 port=8432 dbname=dbName sslmode=disable TimeZone=Asia/Shanghai" --schema="public" --table="table1,table2" --only_model --gen_hook
` + "```" + `

### 2、swagger使用
` + "```shell" + `
## swagger 依赖
go get "github.com/swaggo/files"
go get "github.com/swaggo/gin-swagger"


## swagger 命令行工具
go install github.com/swaggo/swag/cmd/swag@v1.8.12

###注释文档 main函数
// @title 这里写标题
// @version 这里写版本号
// @description 这里写描述信息
// @termsOfService http://swagger.io/terms/

// @contact.name 这里写联系人信息
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host 这里写接口服务的host
// @BasePath 这里写base path（eg：/api/v1）
func main() {}

### 接口层 controller
// @Summary 升级版帖子列表接口
// @Description 可按社区按时间或分数排序查询帖子列表接口
// @Tags 帖子相关接口
// @Accept application/json
// @Produce application/json
// @Param Authorization header string false "Bearer 用户令牌"
// @Param object query models.ParamArtList(请求参数结构体) false "查询参数"
// @Security ApiKeyAuth
// @Success 200 {object} _ResponseArtList
// @Router /接口路由 [请求类型]
func GetArt(c *gin.Context) {}

### 结构体 struct
// 文章列表接口数据信息
type _ResponseArticle struct {
	Code    int              ` + " `json:" + `"code"` + "` // 业务状态码" + `
	Message string           ` + " `json:" + `"message"` + "` // 提示信息" + `
	Data    *[]model.Article ` + " `json:" + `"data"` + "` // 数据" + `
}

### 生成文档，执行：
swag init
}
` + "```" + `
`

const Docs = `// Package docs Code generated by swaggo/swag. DO NOT EDIT
package docs
`

const Helper = `package helper

import (
	"context"

	"github.com/jasonlabz/potato/consts"
)

func contextValue(ctx context.Context, key string) any {
	return ctx.Value(key)
}

func GetClientIP(ctx context.Context) string {
	return contextValue(ctx, consts.ContextClientAddr).(string)
}

func GetUserID(ctx context.Context) string {
	return contextValue(ctx, consts.ContextUserID).(string)
}

func GetToken(ctx context.Context) string {
	return contextValue(ctx, consts.ContextToken).(string)
}
`

const GOMOD = `module {{.ModulePath}}

go 1.22.0

require (
	github.com/gin-gonic/gin v1.10.0
	github.com/google/uuid v1.6.0
	github.com/jasonlabz/knife4go v1.0.1-0.20241118142759-6386e3973279
	github.com/jasonlabz/potato v1.0.3-0.20251115162246-a5ca1bca7a31
)

require (
	filippo.io/edwards25519 v1.1.0 // indirect
	github.com/BurntSushi/toml v1.4.0 // indirect
	github.com/KyleBanks/depth v1.2.1 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/bits-and-blooms/bitset v1.13.0 // indirect
	github.com/bytedance/sonic v1.12.9-0.20250213092650-2d5c1c62c7e8 // indirect
	github.com/bytedance/sonic/loader v0.2.3 // indirect
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/cloudwego/base64x v0.1.5 // indirect
	github.com/dgrijalva/jwt-go v3.2.1-0.20210802184156-9742bd7fca1c+incompatible // indirect
	github.com/dustin/go-humanize v1.0.1 // indirect
	github.com/emirpasic/gods v1.18.1 // indirect
	github.com/fsnotify/fsnotify v1.6.0 // indirect
	github.com/gabriel-vasile/mimetype v1.4.6 // indirect
	github.com/gin-contrib/sse v0.1.0 // indirect
	github.com/glebarez/go-sqlite v1.21.2 // indirect
	github.com/go-logfmt/logfmt v0.6.0 // indirect
	github.com/go-openapi/jsonpointer v0.21.0 // indirect
	github.com/go-openapi/jsonreference v0.21.0 // indirect
	github.com/go-openapi/spec v0.21.0 // indirect
	github.com/go-openapi/swag v0.23.0 // indirect
	github.com/go-playground/locales v0.14.1 // indirect
	github.com/go-playground/universal-translator v0.18.1 // indirect
	github.com/go-playground/validator/v10 v10.23.0 // indirect
	github.com/go-resty/resty/v2 v2.9.1 // indirect
	github.com/go-sql-driver/mysql v1.8.1 // indirect
	github.com/goccy/go-json v0.10.3 // indirect
	github.com/godror/godror v0.44.0 // indirect
	github.com/godror/knownpb v0.1.1 // indirect
	github.com/golang-sql/civil v0.0.0-20220223132316-b832511892a9 // indirect
	github.com/golang-sql/sqlexp v0.1.0 // indirect
	github.com/golang/snappy v0.0.4 // indirect
	github.com/hashicorp/hcl v1.0.0 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/pgx/v5 v5.6.0 // indirect
	github.com/jackc/puddle/v2 v2.2.1 // indirect
	github.com/jasonlabz/gorm-dm-driver v0.1.1 // indirect
	github.com/jasonlabz/oracle v1.1.1-0.20240609161033-cf780c860ebb // indirect
	github.com/jasonlabz/sqlite v1.11.1 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/klauspost/cpuid/v2 v2.2.9 // indirect
	github.com/leodido/go-urn v1.4.0 // indirect
	github.com/magiconair/properties v1.8.7 // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/microsoft/go-mssqldb v1.7.2 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/pelletier/go-toml/v2 v2.2.3 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/prometheus/client_golang v1.19.0 // indirect
	github.com/prometheus/client_model v0.6.0 // indirect
	github.com/prometheus/common v0.51.1 // indirect
	github.com/prometheus/procfs v0.13.0 // indirect
	github.com/remyoudompheng/bigfft v0.0.0-20230129092748-24d4a6f8daec // indirect
	github.com/sagikazarmark/locafero v0.3.0 // indirect
	github.com/sagikazarmark/slog-shim v0.1.0 // indirect
	github.com/sourcegraph/conc v0.3.0 // indirect
	github.com/spf13/afero v1.10.0 // indirect
	github.com/spf13/cast v1.5.1 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/spf13/viper v1.17.0 // indirect
	github.com/subosito/gotenv v1.6.0 // indirect
	github.com/swaggo/swag v1.16.4 // indirect
	github.com/thoas/go-funk v0.9.3 // indirect
	github.com/twitchyliquid64/golang-asm v0.15.1 // indirect
	github.com/ugorji/go/codec v1.2.12 // indirect
	go.uber.org/multierr v1.10.0 // indirect
	go.uber.org/zap v1.27.0 // indirect
	golang.org/x/arch v0.14.0 // indirect
	golang.org/x/crypto v0.33.0 // indirect
	golang.org/x/exp v0.0.0-20250207012021-f9890c6ad9f3 // indirect
	golang.org/x/net v0.34.0 // indirect
	golang.org/x/sync v0.11.0 // indirect
	golang.org/x/sys v0.30.0 // indirect
	golang.org/x/text v0.22.0 // indirect
	golang.org/x/tools v0.29.0 // indirect
	google.golang.org/protobuf v1.35.2 // indirect
	gopkg.in/ini.v1 v1.67.0 // indirect
	gopkg.in/natefinch/lumberjack.v2 v2.2.1 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	gorm.io/driver/mysql v1.5.7 // indirect
	gorm.io/driver/postgres v1.5.9 // indirect
	gorm.io/driver/sqlserver v1.5.3 // indirect
	gorm.io/gorm v1.25.12 // indirect
	modernc.org/libc v1.22.5 // indirect
	modernc.org/mathutil v1.5.0 // indirect
	modernc.org/memory v1.5.0 // indirect
	modernc.org/sqlite v1.23.1 // indirect
)
`

const Conf = `application:
  name: {{.ProjectName}}    # 应用名
  debug: true        # 调试模式
  server:
    http:
      port: 8080
      read_timeout: 30s    # 添加超时配置
      write_timeout: 30s
    grpc:
      port: 8082
      max_concurrent_streams: 100
  monitor:
    prometheus:
      enable: false      # Enable prometheus client
      path: "metrics"   # Default value is "metrics", set path as needed.
      scrape_interval: "15s"  # 添加采集间隔
    pprof:
      enable: false  # Enable PProf tool
      port: 8080
      enabled_endpoints: ["goroutine", "heap"]  # 指定启用的端点
kafka:
  enable: false
  topic: ["XXX"]
  group_id: "XXX"
  bootstrap_servers: "XXX:XX,XXX:XX,XXX:XX"
  security_protocol: "PLAINTEXT"
  sasl_mechanism: "PLAIN"
  sasl_username: "XXX"
  sasl_password: "XXX"
database:
  enable: false
  db_type: "mysql"
  dsn: "user:passwd@tcp(*******:8306)/lg_server?charset=utf8mb4&parseTime=True&loc=Local&timeout=20s"
#  dsn: "user=postgres password=halojeff host=127.0.0.1 port=8432 dbname=lg_server sslmode=disable TimeZone=Asia/Shanghai"
  charset: "utf-8"
  log_mode: "info"
  max_idle_conn: 10
  max_open_conn: 100
redis:
  enable: false
  endpoints:
    - "*******:8379"
  password: "*******"
  index_db: 0
  MinIdleConns: 10
  max_idle_conns: 50
  max_active_conns: 10
  max_retry_times: 5
  master_name:
  sentinel_username:
  sentinel_password:
rabbitmq:
  enable: false
  host: "*******"
  port: 8672
  username: lucas
  password: "*******"
crypto:
  - type: aes
    key: "wrEDGh75pxAUH8Mr"
  - type: des
    key: "b_K3prT8"
`
const Resource = `
package resource

import "github.com/jasonlabz/potato/log"

// 文件服务账号密码
var (
	Username string
	Password string
)

// Logger 日志对象
var Logger *log.LoggerWrapper
`
const LOG = `# 是否写入文件
name: service
# json|console
format: console
# error|warn|info|debug|fatal
log_level: debug
# 文件配置
write_file: true
# 日志文件路径
base_path: log
# 日志文件大小
max_size: 10
# 日志文件最大天数
max_age: 28
# 最大存在数量
max_backups: 100
# 是否压缩日志
compress: false
`

const SERVICER = `# service名
Name: demo
# 连接协议
Protocol: http
# 重试次数
RetryCount: 3
# 重试等待时间
RetryWaitTime: 1000
# 请求超时时间
Timeout: 5000
# service ip地址
Host: 127.0.0.1
# service 端口
Port: 8080
# service basepath
BasePath: /
`

const MAKEFILE = `# 工作目录变量
WORKDIR := $(shell pwd)
OUTDIR := $(WORKDIR)/output

# 目标二进制名称
TARGETNAME = {{.ProjectName}}

GOPKGS := $$(go list ./.. | grep -vE "vendor")

# 设置编译时所需要的 Go 环境
export GOENV = $(WORKDIR)/go.env

#执行编译，可使用命令 make 或 make all 执行， 顺序执行 prepare -> compile -> test -> package 几个阶段
all: prepare compile test package

# prepare阶段， 下载 Go 依赖，可单独执行命令: make prepare
prepare:
	git version     # 低于 2.17.1 可能不能正常工作
	go env          # 打印出 go 环境信息，可用于排查问题
	go mod download || go mod download -x # 下载 Go 依赖

# compile 阶段，执行编译命令，可单独执行命令: make compile
compile:build
build: prepare
	go build -o $(WORKDIR)/bin/$(TARGETNAME)
	#bash cmd/build.sh

# test 阶段，进行单元测试， 可单独执行命令: make test
# cover 平台会优先执行此命令
test: prepare
	go test -race -timeout=300s -v -cover $(GOPKGS) -coverprofile=coverage.out | tee unittest.txt

# package 阶段，对编译产出进行打包，输出到 output 目录， 可单独执行命令: make package
package:
	$(shell rm -rf $(OUTDIR))
	$(shell mkdir -p $(OUTDIR))
	$(shell mkdir -p $(OUTDIR)/var/)
	$(shell cp -a bin $(OUTDIR)/bin)
	$(shell cp -a conf $(OUTDIR)/conf)
	$(shell if [ -d "data" ]; then cp -r data $(OUTDIR)/data; fi)
	$(shell if [ -d "script" ]; then cp -r script $(OUTDIR)/script; fi)
	$(shell if [ -d "webroot" ]; then cp -r webroot $(OUTDIR)/; fi)
	tree $(OUTDIR)

# clean 阶段，清除过程中的输出， 可单独执行命令: make clean
clean:
	rm -rf $(OUTDIR)

# avoid filename conflict and speed up build
.PHONY: all prepare compile test package  clean build`
