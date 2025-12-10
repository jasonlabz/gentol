// Package metadata
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
//  Created:   2025/12/10 1:23
//  Version:   v1.0.0

package metadata

const Bootstrap = `package bootstrap

import (
	"context"
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
	// 初始化RMQ
	initRMQ(ctx)
	// 初始化Redis
	initRedis(ctx)
	// 初始化ES
	initES(ctx)
	// 初始化客户端信息
	initServicer(ctx)
}

func initLogger(_ context.Context) {
	resource.Logger = log.GetLogger()
}

func initResource(_ context.Context) {
	// all global variable should be initial
}

func initCrypto(_ context.Context) {
	cryptoConfigs := GetConfig().Crypto
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
	dbConf := GetConfig().DataSource
	if !dbConf.Enable {
		return
	}
	gormConfig := &gormx.Config{}
	err := utils.CopyStruct(dbConf, gormConfig)
	if err != nil {
		panic(err)
	}
	gormConfig.DBName = gormx.DefaultDBNameMaster
	gormConfig.Logger =
		gormx.LoggerAdapter(resource.Logger.WithCallerSkip(3))
	_, err = gormx.InitConfig(gormConfig)
	if err != nil {
		panic(err)
	}
	// dao.SetGormDB(db)
}

func initRMQ(_ context.Context) {
	// 走默认初始化逻辑
	// resource.RMQClient = rabbitmqx.GetRabbitMQOperator()
}

func initRedis(_ context.Context) {
	// 走默认初始化逻辑
	// resource.RedisClient = goredis.GetRedisOperator()
}

func initES(_ context.Context) {
	// 走默认初始化逻辑
	// resource.EsClient = es.GetESOperator()
}

func initConfig(_ context.Context) {
	filePaths, err := utils.ListDir("conf", ".yaml")
	if err != nil {
		filePaths = []string{}
	}
	for _, filePath := range filePaths {
		fileName := filepath.Base(filePath)
		provider, err := file.NewConfigProvider(filePath)
		if err != nil {
			continue
		}
		configx.AddProviders(fileName, provider)
	}
}

func initServicer(_ context.Context) {
	filePaths, _ := utils.ListDir(filepath.Join("conf", "servicer"), ".yaml")
	for _, filePath := range filePaths {
		info := &httpx.Config{}
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

const SERVER_CONFIG = `package bootstrap

import (
	"errors"
	"fmt"
	"log"
	"math"
	"time"

	"github.com/jasonlabz/potato/configx/file"
	"github.com/jasonlabz/potato/utils"
)

var confPaths = []string{"./conf/application.yaml", "./conf/server.yaml", "./conf/config.yaml",
	"./conf/application.ini", "./conf/server.ini", "./conf/config.ini", "./conf/application.ini",
	"./conf/server.ini", "./conf/config.ini"}

type CryptoType string

const (
	CryptoTypeAES  CryptoType = "aes"
	CryptoTypeDES  CryptoType = "des"
	CryptoTypeHMAC CryptoType = "hmac"
)

// CryptoConfig 加密配置
type CryptoConfig struct {
	Type string ` + "`" + `mapstructure:"type" json:"type" ini:"type" yaml:"type"` + "`" + `
	Key  string ` + "`" + `mapstructure:"key" json:"key" ini:"key" yaml:"key"` + "`" + `
}

// KafkaConfig 配置
type KafkaConfig struct {
	Topic            []string ` + "`" + `mapstructure:"topic" json:"topic" yaml:"topic" ini:"topic"` + "`" + `
	Strict           bool     ` + "`" + `mapstructure:"strict" json:"strict" yaml:"strict" ini:"strict"` + "`" + `
	GroupId          string   ` + "`" + `mapstructure:"group_id" json:"group_id" yaml:"group_id" ini:"group_id"` + "`" + `
	BootstrapServers string   ` + "`" + `mapstructure:"bootstrap_servers" json:"bootstrap_servers" yaml:"bootstrap_servers" ini:"bootstrap_servers"` + "`" + `
	SecurityProtocol string   ` + "`" + `mapstructure:"security_protocol" json:"security_protocol" yaml:"security_protocol" ini:"security_protocol"` + "`" + `
	SaslMechanism    string   ` + "`" + `mapstructure:"sasl_mechanism" json:"sasl_mechanism" yaml:"sasl_mechanism" ini:"sasl_mechanism"` + "`" + `
	SaslUsername     string   ` + "`" + `mapstructure:"sasl_username" json:"sasl_username" yaml:"sasl_username" ini:"sasl_username"` + "`" + `
	SaslPassword     string   ` + "`" + `mapstructure:"sasl_password" json:"sasl_password" yaml:"sasl_password" ini:"sasl_password"` + "`" + `
}

// DataSource 连接配置
type DataSource struct {
	Enable  bool   ` + "`" + `mapstructure:"enable" json:"enable" yaml:"enable" ini:"enable"` + "`" + `
	Strict  bool   ` + "`" + `mapstructure:"strict" json:"strict" yaml:"strict" ini:"strict"` + "`" + `
	DBType  string ` + "`" + `mapstructure:"db_type" json:"db_type" yaml:"db_type" ini:"db_type"` + "`" + `
	LogMode string ` + "`" + `mapstructure:"log_mode" json:"log_mode" yaml:"log_mode" ini:"log_mode"` + "`" + `

	Connection      ` + "`" + `mapstructure:",squash"` + "`" + `
	Masters         []Connection ` + "`" + `mapstructure:"masters" json:"masters" yaml:"masters" ini:"masters"` + "`" + `
	Replicas        []Connection ` + "`" + `mapstructure:"replicas" json:"replicas" yaml:"replicas" ini:"replicas"` + "`" + `
	Args            []ARG        ` + "`" + `mapstructure:"args" json:"args" yaml:"args" ini:"args"` + "`" + `
	Charset         string       ` + "`" + `mapstructure:"charset" json:"charset" yaml:"charset" ini:"charset"` + "`" + `
	MaxIdleConn     int          ` + "`" + `mapstructure:"max_idle_conn" json:"max_idle_conn" yaml:"max_idle_conn" ini:"max_idle_conn"` + "`" + `
	MaxOpenConn     int          ` + "`" + `mapstructure:"max_open_conn" json:"max_open_conn" yaml:"max_open_conn" ini:"max_open_conn"` + "`" + `
	ConnMaxIdleTime int64        ` + "`" + `mapstructure:"conn_max_idle_time" json:"conn_max_idle_time"` + "`" + ` // 连接最大空闲时间
	ConnMaxLifeTime int64        ` + "`" + `mapstructure:"conn_max_life_time" json:"conn_max_life_time" yaml:"conn_max_life_time" ini:"conn_max_life_time"` + "`" + `
}

type Connection struct {
	DSN string ` + "`" + `mapstructure:"dsn" json:"dsn" yaml:"dsn" ini:"dsn"` + "`" + `

	Host     string ` + "`" + `mapstructure:"host" json:"host" yaml:"host" ini:"host"` + "`" + `
	Port     int    ` + "`" + `mapstructure:"port" json:"port" yaml:"port" ini:"port"` + "`" + `
	Username string ` + "`" + `mapstructure:"username" json:"username" yaml:"username" ini:"username"` + "`" + `
	Password string ` + "`" + `mapstructure:"password" json:"password" yaml:"password" ini:"password"` + "`" + `
	Database string ` + "`" + `mapstructure:"database" json:"database" yaml:"database" ini:"database"` + "`" + `
}

type ARG struct {
	Name  string ` + "`" + `mapstructure:"name" json:"name" yaml:"name" ini:"name"` + "`" + `
	Value string ` + "`" + `mapstructure:"value" json:"value" yaml:"value" ini:"value"` + "`" + `
}

// RedisConfig 连接配置
type RedisConfig struct {
	Enable           bool     ` + "`" + `mapstructure:"enable" json:"enable" yaml:"enable" ini:"enable"` + "`" + `
	Strict           bool     ` + "`" + `mapstructure:"strict" json:"strict" yaml:"strict" ini:"strict"` + "`" + `
	Endpoints        []string ` + "`" + `mapstructure:"endpoints" json:"endpoints" yaml:"endpoints" ini:"endpoints"` + "`" + `
	Username         string   ` + "`" + `mapstructure:"username" json:"username" yaml:"username" ini:"username"` + "`" + `
	Password         string   ` + "`" + `mapstructure:"password" json:"password" yaml:"password" ini:"password"` + "`" + `
	ClientName       string   ` + "`" + `mapstructure:"client_name" json:"client_name" yaml:"client_name" ini:"client_name"` + "`" + ` // 自定义客户端名
	MasterName       string   ` + "`" + `mapstructure:"master_name" json:"master_name" yaml:"master_name" ini:"master_name"` + "`" + ` // 主节点
	IndexDB          int      ` + "`" + `mapstructure:"index_db" json:"index_db" yaml:"index_db" ini:"index_db"` + "`" + `
	MinIdleConns     int      ` + "`" + `mapstructure:"min_idle_conns" json:"min_idle_conns" yaml:"min_idle_conns" ini:"min_idle_conns"` + "`" + `
	MaxIdleConns     int      ` + "`" + `mapstructure:"max_idle_conns" json:"max_idle_conns" yaml:"max_idle_conns" ini:"max_idle_conns"` + "`" + `
	MaxActiveConns   int      ` + "`" + `mapstructure:"max_active_conns" json:"max_active_conns" yaml:"max_active_conns" ini:"max_active_conns"` + "`" + `
	MaxRetryTimes    int      ` + "`" + `mapstructure:"max_retry_times" json:"max_retry_times" yaml:"max_retry_times" ini:"max_retry_times"` + "`" + `
	SentinelUsername string   ` + "`" + `mapstructure:"sentinel_username" json:"sentinel_username" yaml:"sentinel_username" ini:"sentinel_username"` + "`" + `
	SentinelPassword string   ` + "`" + `mapstructure:"sentinel_password" json:"sentinel_password" yaml:"sentinel_password" ini:"sentinel_password"` + "`" + `
}

type Elasticsearch struct {
	Enable             bool     ` + "`" + `mapstructure:"enable" json:"enable" yaml:"enable" ini:"enable"` + "`" + `
	Strict             bool     ` + "`" + `mapstructure:"strict" json:"strict" yaml:"strict" ini:"strict"` + "`" + `
	Endpoints          []string ` + "`" + `mapstructure:"endpoints" json:"endpoints" yaml:"endpoints" ini:"endpoints"` + "`" + `
	Username           string   ` + "`" + `mapstructure:"username" json:"username" yaml:"username" ini:"username"` + "`" + `
	Password           string   ` + "`" + `mapstructure:"password" json:"password" yaml:"password" ini:"password"` + "`" + `
	IsHttps            bool     ` + "`" + `mapstructure:"is_https" json:"is_https" yaml:"is_https" ini:"is_https"` + "`" + `
	CloudId            string   ` + "`" + `mapstructure:"cloud_id" json:"cloud_id" yaml:"cloud_id" ini:"cloud_id"` + "`" + `
	APIKey             string   ` + "`" + `mapstructure:"api_key" json:"api_key" yaml:"api_key" ini:"api_key"` + "`" + `
	CACert             string   ` + "`" + `mapstructure:"ca_cert" json:"ca_cert" yaml:"ca_cert" ini:"ca_cert"` + "`" + `                                                                          // 客户端证书, 例如："certs/client.pem"
	InsecureSkipVerify bool     ` + "`" + `mapstructure:"insecure_skip_verify" json:"insecure_skip_verify" yaml:"insecure_skip_verify" ini:"insecure_skip_verifyinsecure_skip_verifyv"` + "`" + ` // 跳过证书认证，生产应为false
}

type MongodbConf struct {
	Enable          bool   ` + "`" + `mapstructure:"enable" json:"enable" yaml:"enable" ini:"enable"` + "`" + `
	Strict          bool   ` + "`" + `mapstructure:"strict" json:"strict" yaml:"strict" ini:"strict"` + "`" + `
	Host            string ` + "`" + `mapstructure:"host" json:"host" yaml:"host" ini:"host"` + "`" + `
	Port            int    ` + "`" + `mapstructure:"port" json:"port" yaml:"port" ini:"port"` + "`" + `
	Username        string ` + "`" + `mapstructure:"username" json:"username" yaml:"username" ini:"username"` + "`" + `
	Password        string ` + "`" + `mapstructure:"password" json:"password" yaml:"password" ini:"password"` + "`" + `
	MaxPoolSize     int    ` + "`" + `mapstructure:"max_pool_size" json:"max_pool_size" yaml:"max_pool_size" ini:"max_pool_size"` + "`" + `
	ConnectTimeout  int    ` + "`" + `mapstructure:"connect_timeout" json:"connect_timeout" yaml:"connect_timeout" ini:"connect_timeout"` + "`" + `
	MaxConnIdleTime int    ` + "`" + `mapstructure:"max_conn_idle_time" json:"max_conn_idle_time" yaml:"max_conn_idle_time" ini:"max_conn_idle_time"` + "`" + `
}

type RabbitMQConf struct {
	Enable    bool      ` + "`" + `mapstructure:"enable" json:"enable" yaml:"enable" ini:"enable"` + "`" + `
	Strict    bool      ` + "`" + `mapstructure:"strict" json:"strict" yaml:"strict" ini:"strict"` + "`" + `
	Host      string    ` + "`" + `mapstructure:"host" json:"host" yaml:"host" ini:"host"` + "`" + `
	Port      int       ` + "`" + `mapstructure:"port" json:"port" yaml:"port" ini:"port"` + "`" + `
	Username  string    ` + "`" + `mapstructure:"username" json:"username" yaml:"username" ini:"username"` + "`" + `
	Password  string    ` + "`" + `mapstructure:"password" json:"password" yaml:"password" ini:"password"` + "`" + `
	LimitConf LimitConf ` + "`" + `mapstructure:"limit_conf" json:"limit_conf" yaml:"limit_conf" ini:"limit_conf"` + "`" + `
}

type LimitConf struct {
	Enable        bool  ` + "`" + `mapstructure:"enable" json:"enable" yaml:"enable" ini:"enable"` + "`" + `
	AttemptTimes  int   ` + "`" + `mapstructure:"attempt_times" json:"attempt_times" yaml:"attempt_times" ini:"attempt_times"` + "`" + `
	RetryWaitTime int64 ` + "`" + `mapstructure:"retry_wait_time" json:"retry_wait_time" yaml:"retry_wait_time" ini:"retry_wait_time"` + "`" + `
	PrefetchCount int   ` + "`" + `mapstructure:"prefetch_count" json:"prefetch_count" yaml:"prefetch_count" ini:"prefetch_count"` + "`" + `
	Timeout       int64 ` + "`" + `mapstructure:"timeout" json:"timeout" yaml:"timeout" ini:"timeout"` + "`" + `
	QueueLimit    int   ` + "`" + `mapstructure:"queue_limit" json:"queue_limit" yaml:"queue_limit" ini:"queue_limit"` + "`" + `
}

// ServerConfig 新增的配置结构体
type ServerConfig struct {
	HTTP   HTTPConfig   ` + "`" + `mapstructure:"http" json:"http" yaml:"http" ini:"http"` + "`" + `
	GRPC   GRPCConfig   ` + "`" + `mapstructure:"grpc" json:"grpc" yaml:"grpc" ini:"grpc"` + "`" + `
	Static StaticConfig ` + "`" + `mapstructure:"static" json:"static" yaml:"static" ini:"static"` + "`" + `
}

type HTTPConfig struct {
	Port int ` + "`" + `mapstructure:"port" json:"port" yaml:"port" ini:"port"` + "`" + `

	// Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h".
	// such as "300ms", "3000", "-1.5h" or "2h45m", default unit "ms".
	ReadTimeout string ` + "`" + `mapstructure:"read_timeout" json:"read_timeout" yaml:"read_timeout" ini:"read_timeout"` + "`" + `

	// Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h".
	// such as "300ms", "3000", "-1.5h" or "2h45m", default unit "ms".
	WriteTimeout string ` + "`" + `mapstructure:"write_timeout" json:"write_timeout" yaml:"write_timeout" ini:"write_timeout"` + "`" + `
}

type StaticConfig struct {
	Port     int    ` + "`" + `mapstructure:"port" json:"port" yaml:"port" ini:"port"` + "`" + `
	Path     string ` + "`" + `mapstructure:"path" json:"path" yaml:"path" ini:"path"` + "`" + `                 // 静态资源路径
	Username string ` + "`" + `mapstructure:"username" json:"username" yaml:"username" ini:"username"` + "`" + ` // 静态资源可以配置auth, 为空则不校验
	Password string ` + "`" + `mapstructure:"password" json:"password" yaml:"password" ini:"password"` + "`" + `
}

type GRPCConfig struct {
	Port                 int    ` + "`" + `mapstructure:"port" json:"port" yaml:"port" ini:"port"` + "`" + `
	MaxConcurrentStreams uint32 ` + "`" + `mapstructure:"max_concurrent_streams" json:"max_concurrent_streams" yaml:"max_concurrent_streams" ini:"max_concurrent_streams"` + "`" + `
}

type MonitorConfig struct {
	Prometheus PrometheusConfig ` + "`" + `mapstructure:"prometheus" json:"prometheus" yaml:"prometheus" ini:"prometheus"` + "`" + `
	PProf      PProfConfig      ` + "`" + `mapstructure:"pprof" json:"pprof" yaml:"pprof" ini:"pprof"` + "`" + `
}

type PrometheusConfig struct {
	Enable         bool   ` + "`" + `mapstructure:"enable" json:"enable" yaml:"enable" ini:"enable"` + "`" + `
	Path           string ` + "`" + `mapstructure:"path" json:"path" yaml:"path" ini:"path"` + "`" + `
	ScrapeInterval string ` + "`" + `mapstructure:"scrape_interval" json:"scrape_interval" yaml:"scrape_interval" ini:"scrape_interval"` + "`" + `
}

type PProfConfig struct {
	Enable           bool     ` + "`" + `mapstructure:"enable" json:"enable" yaml:"enable" ini:"enable"` + "`" + `
	Port             int      ` + "`" + `mapstructure:"port" json:"port" yaml:"port" ini:"port"` + "`" + `
	EnabledEndpoints []string ` + "`" + `mapstructure:"enabled_endpoints" json:"enabled_endpoints" yaml:"enabled_endpoints" ini:"enabled_endpoints"` + "`" + `
	Pusher           Pusher   ` + "`" + `mapstructure:"pusher" json:"pusher" yaml:"pusher" ini:"pusher"` + "`" + ` // 推送到 pushgateway
}

// Pusher push to pushGateway 配置
type Pusher struct {
	Enable     bool   ` + "`" + `mapstructure:"enable" json:"enable" yaml:"enable" ini:"enable"` + "`" + `                     // Enable backend job push metrics to remote pushgateway
	JobName    string ` + "`" + `mapstructure:"job_name" json:"job_name" yaml:"job_name" ini:"job_name"` + "`" + `             // Name of current push job
	RemoteAddr string ` + "`" + `mapstructure:"remote_addr" json:"remote_addr" yaml:"remote_addr" ini:"remote_addr"` + "`" + ` // Remote address of pushgateway
	IntervalMs int    ` + "`" + `mapstructure:"IntervalMs" json:"IntervalMs" yaml:"interval_ms" ini:"interval_ms"` + "`" + `   // Push interval in milliseconds
	BasicAuth  string ` + "`" + `mapstructure:"basic_auth" json:"basic_auth" yaml:"basic_auth" ini:"basic_auth"` + "`" + `     // Basic auth of pushgateway
}

// Application 配置
type Application struct {
	Host    string        ` + "`" + `mapstructure:"host" json:"host" yaml:"host" ini:"host"` + "`" + `
	Name    string        ` + "`" + `mapstructure:"name" json:"name" yaml:"name" ini:"name"` + "`" + `
	Port    int           ` + "`" + `mapstructure:"port" json:"port" yaml:"port" ini:"port"` + "`" + `
	Debug   bool          ` + "`" + `mapstructure:"debug" json:"debug" yaml:"debug" ini:"debug"` + "`" + `
	Server  ServerConfig  ` + "`" + `mapstructure:"server" json:"server" yaml:"server" ini:"server"` + "`" + `
	Monitor MonitorConfig ` + "`" + `mapstructure:"monitor" json:"monitor" yaml:"monitor" ini:"monitor"` + "`" + `
}

// Config 更新主配置结构体
type Config struct {
	Application Application    ` + "`" + `mapstructure:"application" json:"application" yaml:"application" ini:"application"` + "`" + `
	DataSource  DataSource     ` + "`" + `mapstructure:"datasource" json:"datasource" yaml:"datasource" ini:"datasource"` + "`" + `
	Crypto      []CryptoConfig ` + "`" + `mapstructure:"crypto" json:"crypto" yaml:"crypto" ini:"crypto"` + "`" + `
	Kafka       KafkaConfig    ` + "`" + `mapstructure:"kafka" json:"kafka" yaml:"kafka" ini:"kafka"` + "`" + `
	Rabbitmq    RabbitMQConf   ` + "`" + `mapstructure:"rabbitmq" json:"rabbitmq" yaml:"rabbitmq" ini:"rabbitmq"` + "`" + `
	Redis       RedisConfig    ` + "`" + `mapstructure:"redis" json:"redis" yaml:"redis" ini:"redis"` + "`" + `
	ES          Elasticsearch  ` + "`" + `mapstructure:"es" json:"es" yaml:"es" ini:"es"` + "`" + `
	Mongodb     MongodbConf    ` + "`" + `mapstructure:"mongodb" json:"mongodb" yaml:"mongodb" ini:"mongodb"` + "`" + `
}

var applicationConfig = new(Config)

func GetConfig() *Config {
	return applicationConfig
}

func (c *Config) GetServerConfig() ServerConfig {
	return c.Application.Server
}

func (c *Config) GetHTTPPort() int {
	// 优先使用新的 server.http.port 配置
	if c.Application.Server.HTTP.Port > 0 {
		return c.Application.Server.HTTP.Port
	}
	// 回退到旧的 application.port 配置
	return c.Application.Port
}

func (c *Config) GetName() string {
	return c.Application.Name
}

func (c *Config) IsDebugMode() bool {
	return c.Application.Debug
}

func (c *Config) GetPrometheusConfig() PrometheusConfig {
	return c.Application.Monitor.Prometheus
}

func (c *Config) GetPProfConfig() PProfConfig {
	return c.Application.Monitor.PProf
}

// GetHTTPReadTimeout 获取超时时间（转换为 time.Duration）
func (c *Config) GetHTTPReadTimeout() time.Duration {
	if c.Application.Server.HTTP.ReadTimeout != "" {
		duration, err := time.ParseDuration(c.Application.Server.HTTP.ReadTimeout)
		if err != nil {
			log.Printf("parse http read timeout error: %v", err)
			duration = time.Duration(math.MaxInt64)
		}
		return duration
	}
	return time.Duration(math.MaxInt64) // 默认值
}

func (c *Config) GetHTTPWriteTimeout() time.Duration {
	if c.Application.Server.HTTP.WriteTimeout != "" {
		duration, err := time.ParseDuration(c.Application.Server.HTTP.WriteTimeout)
		if err != nil {
			log.Printf("parse http write timeout error: %v", err)
			duration = time.Duration(math.MaxInt64)
		}
		return duration
	}
	return time.Duration(math.MaxInt64) // 默认值
}

func (c *Config) GetGRPCPort() int {
	if c.Application.Server.GRPC.Port > 0 {
		return c.Application.Server.GRPC.Port
	}
	return 8631 // 默认值
}

// Validate 配置验证
func (c *Config) Validate() error {
	if c.Application.Name == "" {
		return errors.New("application name is required")
	}

	httpPort := c.GetHTTPPort()
	if httpPort <= 0 || httpPort > 65535 {
		return fmt.Errorf("invalid HTTP port: %d", httpPort)
	}

	grpcPort := c.GetGRPCPort()
	if grpcPort <= 0 || grpcPort > 65535 {
		return fmt.Errorf("invalid GRPC port: %d", grpcPort)
	}

	return nil
}

func init() {
	// 读取服务配置文件
	var configLoad bool
	for _, config := range confPaths {
		if configLoad {
			break
		}
		if !utils.IsExist(config) {
			continue
		}
		err := file.ParseConfigByViper(config, applicationConfig)
		if err != nil {
			log.Printf("[init] -- failed to read config file: %s, err:%v", config, err)
			continue
		}
		configLoad = true
	}

	if !configLoad {
		log.Println("[init] -- there is no application config.")
	}
}
`
