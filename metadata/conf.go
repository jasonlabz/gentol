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
//  Created:   2025/12/3 22:25
//  Version:   v1.0.0

package metadata

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
  strict: true
  topic: ["XXX"]
  group_id: "XXX"
  bootstrap_servers: "XXX:XX,XXX:XX,XXX:XX"
  security_protocol: "PLAINTEXT"
  sasl_mechanism: "PLAIN"
  sasl_username: "XXX"
  sasl_password: "XXX"
database:
  enable: false
  strict: true
  db_type: "mysql"
#  dsn: "user:passwd@tcp(*******:8306)/lg_server?charset=utf8mb4&parseTime=True&loc=Local&timeout=20s"
  host: "*******"
  port: 8306
  username: root
  password: "*******"
  database: dbname
  args:
    - name: charset
      value: utf8mb4
  log_mode: "info"
  max_idle_conn: 10
  max_open_conn: 100
es:
  enable: false
  strict: true
  endpoints:
    - "*******:8776"
  username: elastic
  password: "*************"
  api_key: # 认证方式2（可选）
  is_https: true
  ca_cert: # CA证书
  insecure_skip_verify: false   # 跳过证书认证，生产应为false
redis:
  enable: false
  strict: true
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
  enable: false                # 是否启用
  strict: true                 # 是否为下游必需，如为true则会启动时panic所遇error
  host: "*******"
  port: 8672
  username: lucas
  password: "*******"
  limit_conf:
    attempt_times: 3          # 重试次数
    retry_wait_time: 3000     # 重试等待时间，单位ms
    prefetch_count: 100       # 队列预读取数量
    timeout: 5000             # 超时时间
    queue_limit: 0            # 队列长度限制
crypto:
  - type: aes
    key: "wrEDGh75pxAUH8Mr"
  - type: des
    key: "b_K3prT8"
`

const SERVICER = `# service名
Name: demo
# 调试模式 true|false
Debug: false
# 连接协议 http|https
Protocol: http
# 重试次数
RetryCount: 3
# 重试等待时间 单位：毫秒
RetryWaitTime: 1000
# 请求超时时间 单位：毫秒
Timeout: 5000
# service ip地址
Host: 127.0.0.1
# service 端口
Port: 8080
# service basepath
BasePath: /
# 客户端证书, 例如："certs/client.pem"
CertFile:
# 客户端证书, 例如："certs/client.key"
KeyFile:
# 根证书, 例如："/path/to/root/pemFile.pem"
RootCertFile:
# 跳过证书认证，生产应为false
InsecureSkipVerify: false
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
