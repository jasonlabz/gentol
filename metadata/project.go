package metadata

type ProjectMeta struct {
	ModulePath         string
	ProjectName        string
	ServiceName        string
	ServicePackageName string
	ServiceStructName  string
}

func (p *ProjectMeta) GenRenderData() map[string]any {
	result := map[string]any{
		"ModulePath":         p.ModulePath,
		"ProjectName":        p.ProjectName,
		"ServiceName":        p.ServiceName,
		"ServiceStructName":  p.ServiceStructName,
		"ServicePackageName": p.ServicePackageName,
	}
	return result
}

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
	"github.com/jasonlabz/potato/ginmetrics"

	"{{.ModulePath}}/bootstrap"
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
	serverConfig := bootstrap.GetConfig()
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

	go func() {
		fileServer(serverConfig)
	}()

	// start program
	srv := startServer(r, serverConfig)

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
func startServer(router *gin.Engine, c *bootstrap.Config) *http.Server {
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", c.GetHTTPPort()),
		Handler:      router,
		ReadTimeout:  c.GetHTTPReadTimeout(),
		WriteTimeout: c.GetHTTPWriteTimeout(),
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	return srv
}

// fileServer 文件服务
func fileServer(c *bootstrap.Config) {
	config := c.GetServerConfig().Static
	// 创建 HTTP 服务器
	if config.Path == "" {
		return
	}
	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.Dir(config.Path)))
	if config.Username != "" && config.Password != "" {
		// 使用基本认证保护文件下载路由
		authMux := basicAuth(mux, config.Username, config.Password)
		// 启动 HTTP 服务器
		// log.Printf("Starting file server at :%d", config.GetConfig().Application.Port+1)
		err := http.ListenAndServe(fmt.Sprintf(":%d", config.Port), authMux)
		if err != nil {
			log.Fatalf("file server listen: %s\n", err)
		}
		return
	}
	// 启动 HTTP 服务器
	err := http.ListenAndServe(fmt.Sprintf(":%d", config.Port), mux)
	if err != nil {
		log.Fatalf("file server listen: %s\n", err)
	}
	return
}

// basicAuth 认证检查
func basicAuth(handler http.Handler, username, password string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, pass, ok := r.BasicAuth()
		if !ok || user != username || pass != password {
			w.Header().Set("WWW-Authenticate", ` + "`" + `Basic realm="Restricted"` + "`" + `)
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
	router := gin.New()
	serverConfig := configx.GetConfig()

	// 全局中间件，查看定义的中间价在middlewares文件夹中
	rootMiddleware(router)

	registerRootAPI(router)

	// 对路由进行分组，处理不同的分组，根据自己的需求定义即可
	staticRouter := router.Group("/server")
	staticRouter.Static("/", "application")

	serverGroup := router.Group(fmt.Sprintf("/%s", serverConfig.GetName()))
	// debug模式下，注册swagger路由
	// knife4go: beautify swagger-ui, http://ip:port/server_name/doc.html
	if serverConfig.IsDebugMode() {
		_ = knife4go.InitSwaggerKnife(serverGroup)
	}

	// base api
	registerBaseAPI(serverGroup)

	apiGroup := serverGroup.Group("/api")

	// 中间件拦截器
	groupMiddleware(apiGroup,
		middleware.RecoveryLog(true), middleware.SetContext(), middleware.RequestMiddleware())

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

	"{{.ModulePath}}/global/resource"
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
		resource.Logger.Info(c, "	[GIN] request",
			log.String("proto", c.Request.Proto),
			log.String("client_ip", c.ClientIP()),
			log.Int64("content_length", c.Request.ContentLength),
			log.String("agent", c.Request.UserAgent()),
			log.String("request_body", string(logBytes(requestBodyBytes, requestBodyMaxLen))),
			log.String("method", c.Request.Method),
			log.String("path", c.Request.URL.Path))

		c.Next()

		resource.Logger.Info(c, "	[GIN] response",
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

const ReqDTO = `package body

type HealthCheckReqDto struct {
	FieldName string` + " `json:" + `"field_name"` + "` // 属性名" + `
}

`

const ResDto = `package body

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

const Page = `package ginx

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

const APIVersionV1 = "v1"

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

const Resource = `package resource

import (
	"github.com/jasonlabz/potato/es"
	"github.com/jasonlabz/potato/goredis"
	"github.com/jasonlabz/potato/log"
	"github.com/jasonlabz/potato/rabbitmqx"
)

// Logger 日志对象
var Logger *log.LoggerWrapper

// RMQClient rabbitmq 客户端
var RMQClient *rabbitmqx.RabbitMQOperator

// RedisClient redis 客户端
var RedisClient *goredis.RedisOperator

// EsClient es 客户端
var EsClient *es.ElasticSearchOperator
`
