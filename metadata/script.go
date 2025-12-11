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
//  Created:   2025/12/1 23:48
//  Version:   v1.0.0

package metadata

const SCRIPT_README = `## 概念
### 1、gentol.sh|gentol.ps1
> 生成dao、model层代码，提高代码开发效率。

### 2、generate_idl.sh|generate_idl.ps1
> 解析idl文件，生成rpc代码

### 3、swag.sh|swag.ps1
> 解析生成swag文档，方便接口开发
`
const SCRIPT_SWAG_PS1 = `#!/usr/bin/env pwsh

param()

# 配置
$SWAG_CMD = if ($env:SWAG_CMD) { $env:SWAG_CMD } else { "swag" }
$SWAG_DIR = if ($env:SWAG_DIR) { $env:SWAG_DIR } else { "./bin" }
$PROJECT_DIR = if ($env:PROJECT_DIR) { $env:PROJECT_DIR } else { "." }

# 日志函数
function Write-Log {
    param([string]$Message)
    Write-Host "[$(Get-Date -Format 'HH:mm:ss')] $Message"
}

function Write-InfoLog {
    param([string]$Message)
    Write-Log "INFO: $Message"
}

function Write-ErrorLog {
    param([string]$Message)
    Write-Log "ERROR: $Message"
    exit 1
}

# 检查依赖
function Check-Swag {
    # 首先检查是否在PATH中
    if (Get-Command $SWAG_CMD -ErrorAction SilentlyContinue) {
        return $true
    }

    # 检查指定目录下的swag可执行文件
    $swagPath = Join-Path $SWAG_DIR "swag.exe"
    if (Test-Path $swagPath) {
        $script:SWAG_CMD = $swagPath
        return $true
    }

    # 检查Linux/Mac格式的可执行文件（在PowerShell Core跨平台环境中）
    $swagPathUnix = Join-Path $SWAG_DIR "swag"
    if (Test-Path $swagPathUnix) {
        $script:SWAG_CMD = $swagPathUnix
        return $true
    }

    Write-InfoLog "Installing swag..."
    try {
        go install github.com/swaggo/swag/cmd/swag@latest

        # 安装后重新检查
        if (Get-Command $SWAG_CMD -ErrorAction SilentlyContinue) {
            return $true
        }

        # 检查Go的bin目录
        if ($env:GOPATH) {
            $goBinSwag = Join-Path $env:GOPATH "bin" "swag.exe"
            if (Test-Path $goBinSwag) {
                $script:SWAG_CMD = $goBinSwag
                return $true
            }

            $goBinSwagUnix = Join-Path $env:GOPATH "bin" "swag"
            if (Test-Path $goBinSwagUnix) {
                $script:SWAG_CMD = $goBinSwagUnix
                return $true
            }
        }

        return $false
    }
    catch {
        Write-Host "Failed to install swag: $_" -ForegroundColor Red
        return $false
    }
}

# 运行命令
function Run-Swag {
    param([string]$Command)

    Write-InfoLog "Running: swag $Command"

    try {
        # 分割命令参数
        $arguments = $Command -split ' '
        & $SWAG_CMD @arguments

        if ($LASTEXITCODE -ne 0) {
            Write-ErrorLog "swag $Command failed with exit code: $LASTEXITCODE"
        }
    }
    catch {
        Write-ErrorLog "swag $Command failed: $_"
    }

    Write-InfoLog "swag $Command completed"
}

# 主流程
function Main {
    Write-InfoLog "Starting swag documentation generation..."

    if (-not (Check-Swag)) {
        Write-ErrorLog "swag not found and installation failed"
    }

    Run-Swag "init"
    Run-Swag "fmt"

    Write-InfoLog "Documentation generation completed!"
}

# 设置错误处理
$ErrorActionPreference = "Stop"

# 执行主函数
Main`

const SCRIPT_SWAG = `#!/bin/bash

set -euo pipefail

# 配置
SWAG_CMD="${SWAG_CMD:-swag}"
SWAG_DIR="${SWAG_DIR:-./bin}"
PROJECT_DIR="${PROJECT_DIR:-.}"

# 日志函数
log() { echo "[$(date '+%H:%M:%S')] $1"; }
log_info() { log "INFO: $1"; }
log_error() { log "ERROR: $1"; exit 1; }

# 检查依赖
check_swag() {
    if command -v "$SWAG_CMD" &>/dev/null; then
        return 0
    fi

    if [[ -f "$SWAG_DIR/swag" ]]; then
        SWAG_CMD="$SWAG_DIR/swag"
        return 0
    fi

    log_info "Installing swag..."
    go install github.com/swaggo/swag/cmd/swag@latest
}

# 运行命令
run_swag() {
    local command="$1"
    log_info "Running: swag $command"

    if ! $SWAG_CMD $command; then
        log_error "swag $command failed"
    fi

    log_info "swag $command completed"
}

# 主流程
main() {
    log_info "Starting swag documentation generation..."

    check_swag || log_error "swag not found and installation failed"
    run_swag "init"
    run_swag "fmt"

    log_info "Documentation generation completed!"
}

main "$@"
`
const SCRIPT_GENTOL_PS1 = `#!/usr/bin/env pwsh

param()

# 配置参数
$GENTOL_CMD = if ($env:GENTOL_CMD) { $env:GENTOL_CMD } else { "gentol" }
$OUTPUT_DIR = if ($env:OUTPUT_DIR) { $env:OUTPUT_DIR } else { "." }
$TEMPLATE_DIR = if ($env:TEMPLATE_DIR) { $env:TEMPLATE_DIR } else { "./template" }
$DSN = ""
# 数据库配置 #TODO: 修改对应参数
# TODO: 数据库类型  "mysql|postgres|sqlserver|oracle|sqlite|dm"
$DB_TYPE = "postgres"
# TODO: 数据库host
$DB_HOST = "****************"
# TODO: 数据库port
$DB_PORT = "8530"
# TODO: 数据库 用户
$DB_USER = "postgres"
# TODO: 数据库 密码
$DB_PASS = "****************"
# TODO: 数据库 库名
$DB_NAME = "database"
# TODO: 数据库 模式
$DB_SCHEMA = ""
# TODO: 需要生成的表结构，不配置则为全部
$TABLES = if ($env:TABLES) { $env:TABLES } else { "" }

# 生成配置
$MODEL_DIR = if ($env:MODEL_DIR) { $env:MODEL_DIR } else { "dal/db/model" }
$DAO_DIR = if ($env:DAO_DIR) { $env:DAO_DIR } else { "dal/db/dao" }

# 功能开关
$ONLY_MODEL = if ($env:ONLY_MODEL) { [bool]::Parse($env:ONLY_MODEL) } else { $false }
$USE_SQL_NULLABLE = if ($env:USE_SQL_NULLABLE) { [bool]::Parse($env:USE_SQL_NULLABLE) } else { $false }
$RUN_GOFMT = if ($env:RUN_GOFMT) { [bool]::Parse($env:RUN_GOFMT) } else { $true }
$GEN_HOOK = if ($env:GEN_HOOK) { [bool]::Parse($env:GEN_HOOK) } else { $true }

# 日志函数
function Write-Log {
    param([string]$Message)
    Write-Host "[$(Get-Date -Format 'HH:mm:ss')] $Message"
}

function Write-InfoLog {
    param([string]$Message)
    Write-Log "INFO: $Message"
}

function Write-ErrorLog {
    param([string]$Message)
    Write-Log "ERROR: $Message"
    exit 1
}

# 构建 DSN（如果未直接提供）
function Build-Dsn {
    if ($DSN -ne "") {
        return
    }

    switch ($DB_TYPE) {
        "mysql" {
            $script:DSN = "${DB_USER}:${DB_PASS}@tcp(${DB_HOST}:${DB_PORT})/${DB_NAME}?parseTime=True&loc=Local"
        }
        "postgres" {
            $script:DSN = "user=$DB_USER password=$DB_PASS host=$DB_HOST port=$DB_PORT dbname=$DB_NAME sslmode=disable TimeZone=Asia/Shanghai"
        }
        "sqlserver" {
            $script:DSN = "user id=$DB_USER;password=$DB_PASS;server=$DB_HOST;port=$DB_PORT;database=$DB_NAME;encrypt=disable"
        }
        "oracle" {
            $script:DSN = "${DB_USER}/${DB_PASS}@${DB_HOST}:${DB_PORT}/${DB_NAME}"
        }
        "sqlite" {
            $script:DSN = $DB_NAME
        }
        "dm" {
            $script:DSN = "dm://${DB_USER}:${DB_PASS}@${DB_HOST}:${DB_PORT}?schema=$DB_SCHEMA"
        }
        default {
            Write-ErrorLog "Unsupported database type: $DB_TYPE"
        }
    }
}

# 构建命令参数
function Build-Args {
    $argsList = @(
        "--db_type=$DB_TYPE"
        "--dsn=` + "`\"$DSN`\"" + `
		"--model=$MODEL_DIR"
		"--dao=$DAO_DIR"
	)

    if ($TABLES -ne "") {
        $argsList += "--table=$TABLES"
    }

    if ($DB_SCHEMA -ne "") {
        $argsList += "--schema=$DB_SCHEMA"
    }

    if ($ONLY_MODEL) {
        $argsList += "--only_model"
    }

    if ($USE_SQL_NULLABLE) {
        $argsList += "--use_sql_nullable"
    }

    if ($RUN_GOFMT) {
        $argsList += "--rungofmt"
    }

    if ($GEN_HOOK) {
        $argsList += "--gen_hook"
    }

    return $argsList
}

# 检查依赖
function Check-Gentol {
    if (Get-Command $GENTOL_CMD -ErrorAction SilentlyContinue) {
        return
    }

    Write-InfoLog "Installing gentol..."
    go install github.com/jasonlabz/gentol@master

    # 重新检查安装是否成功
    if (-not (Get-Command $GENTOL_CMD -ErrorAction SilentlyContinue)) {
        # 尝试在Go的bin目录中查找
        $goBinPath = Join-Path $env:GOPATH "bin" "gentol.exe"
        if (Test-Path $goBinPath) {
            $script:GENTOL_CMD = $goBinPath
        } else {
            # 如果Go在默认位置，尝试使用完整路径
            if ($env:GOPATH) {
                $goBinPath = Join-Path $env:GOPATH "bin" "gentol"
                if (Test-Path $goBinPath) {
                    $script:GENTOL_CMD = $goBinPath
                }
            }
        }
    }
}

# 主流程
function Main {
    Write-InfoLog "Starting code generation with gentol..."

    Check-Gentol
    Build-Dsn

    $args = Build-Args
    $command = "$GENTOL_CMD $args"

    Write-InfoLog "Running: $command"

    try {
        Invoke-Expression $command
        if ($LASTEXITCODE -ne 0) {
            Write-ErrorLog "Code generation failed with exit code: $LASTEXITCODE"
        }
    }
    catch {
        Write-ErrorLog "Code generation failed: $_"
    }

    Write-InfoLog "Code generation completed!"
}

# 设置错误处理
$ErrorActionPreference = "Stop"

# 执行主函数
Main`

const SCRIPT_GENTOL = `#!/bin/bash

set -euo pipefail

# 配置参数
GENTOL_CMD="${GENTOL_CMD:-gentol}"
OUTPUT_DIR="${OUTPUT_DIR:-.}"
TEMPLATE_DIR="${TEMPLATE_DIR:-./template}"
DSN=""
# 数据库配置 #TODO: 修改对应参数
# TODO: 数据库类型  "mysql|postgres|sqlserver|oracle|sqlite|dm"
DB_TYPE="postgres"
# TODO: 数据库host
DB_HOST="****************"
# TODO: 数据库port
DB_PORT="8530"
# TODO: 数据库 用户
DB_USER="postgres"
# TODO: 数据库 密码
DB_PASS="****************"
# TODO: 数据库 库名
DB_NAME="database"
# TODO: 数据库 模式
DB_SCHEMA=""
# TODO: 需要生成的表结构，不配置则为全部
TABLES="${TABLES:-}"

# 生成配置
MODEL_DIR="${MODEL_DIR:-dal/db/model}"
DAO_DIR="${DAO_DIR:-dal/db/dao}"


# 功能开关
ONLY_MODEL="${ONLY_MODEL:-false}"
USE_SQL_NULLABLE="${USE_SQL_NULLABLE:-false}"
RUN_GOFMT="${RUN_GOFMT:-true}"
GEN_HOOK="${GEN_HOOK:-true}"

# 日志函数
log() { echo "[$(date '+%H:%M:%S')] $1"; }
log_info() { log "INFO: $1"; }
log_error() { log "ERROR: $1"; exit 1; }

# 构建 DSN（如果未直接提供）
build_dsn() {
    if [[ -n "$DSN" ]]; then
        return 0
    fi
    if [[ "$DSN" != "" ]]; then
        return 0
    fi

    case "$DB_TYPE" in
        "mysql")
            DSN="$DB_USER:$DB_PASS@tcp($DB_HOST:$DB_PORT)/$DB_NAME?parseTime=True&loc=Local"
            ;;
        "postgres")
            DSN="user=$DB_USER password=$DB_PASS host=$DB_HOST port=$DB_PORT dbname=$DB_NAME sslmode=disable TimeZone=Asia/Shanghai"
            ;;
        "sqlserver")
            DSN="user id=$DB_USER;password=$DB_PASS;server=$DB_HOST;port=$DB_PORT;database=$DB_NAME;encrypt=disable"
            ;;
        "oracle")
            DSN="$DB_USER/$DB_PASS@$DB_HOST:$DB_PORT/$DB_NAME"
            ;;
        "sqlite")
            DSN="$DB_NAME"
            ;;
        "dm")
            DSN="dm://$DB_USER:$DB_PASS@$DB_HOST:$DB_PORT?schema=$DB_SCHEMA"
            ;;
        *)
            log_error "Unsupported database type: $DB_TYPE"
            ;;
    esac
}

# 构建命令参数
build_args() {
    local args=(
        "--db_type=$DB_TYPE"
        "--dsn=\"$DSN\""
        "--model=$MODEL_DIR"
        "--dao=$DAO_DIR"
    )

    [[ -n "$TABLES" ]] && args+=("--table=$TABLES")
    [[ -n "$DB_SCHEMA" ]] && args+=("--schema=$DB_SCHEMA")
    [[ "$ONLY_MODEL" == "true" ]] && args+=("--only_model")
    [[ "$USE_SQL_NULLABLE" == "true" ]] && args+=("--use_sql_nullable")
    [[ "$RUN_GOFMT" == "true" ]] && args+=("--rungofmt")
    [[ "$GEN_HOOK" == "true" ]] && args+=("--gen_hook")

    echo "${args[@]}"
}

# 检查依赖
check_gentol() {
    if command -v "$GENTOL_CMD" &>/dev/null; then
        return 0
    fi

    log_info "Installing gentol..."
    go install github.com/jasonlabz/gentol@master
}

# 主流程
main() {
    log_info "Starting code generation with gentol..."

    check_gentol
    build_dsn

    local args
    args=$(build_args)

    log_info "Running: $GENTOL_CMD $args"

    if ! eval $GENTOL_CMD $args; then
        log_error "Code generation failed"
    fi

    log_info "Code generation completed!"
}

main "$@"
`

const SCRIPT_KITEX_PS1 = `#!/usr/bin/env pwsh

param(
    [string]$KitexCmd = "kitex"
)

# 配置变量
$BASE_MODULE = "{{.ModulePath}}"
$ROOT_DIR = Get-Location
$IDL_DIR = "idl"
$CLIENT_DIR = "client/kitex"
$SERVER_DIR = "server/kitex"

# 颜色输出函数
$ESC = [char]27
$RED = "${ESC}[0;31m"
$GREEN = "${ESC}[0;32m"
$YELLOW = "${ESC}[1;33m"
$NC = "${ESC}[0m" # No Color

function Write-InfoLog {
    param([string]$Message)
    Write-Host "${GREEN}[INFO]${NC} $Message"
}

function Write-WarnLog {
    param([string]$Message)
    Write-Host "${YELLOW}[WARN]${NC} $Message"
}

function Write-ErrorLog {
    param([string]$Message)
    Write-Host "${RED}[ERROR]${NC} $Message"
    exit 1
}

# 检查 kitex 是否可用
function Check-Kitex {
    # 首先检查命令是否存在
    try {
        $cmd = Get-Command $KitexCmd -ErrorAction Stop
        $script:KitexCmd = $cmd.Source
    }
    catch {
        Write-InfoLog "Installing kitex..."
        try {
            go install github.com/cloudwego/kitex/tool/cmd/kitex@latest
            
            # 安装后重新检查
            $kitexPath = Join-Path $env:GOPATH "bin" "kitex.exe"
            if (Test-Path $kitexPath) {
                $script:KitexCmd = $kitexPath
            }
            else {
                $kitexPath = Join-Path $env:GOPATH "bin" "kitex"
                if (Test-Path $kitexPath) {
                    $script:KitexCmd = $kitexPath
                }
                else {
                    Write-ErrorLog "kitex command not found and installation failed"
                }
            }
        }
        catch {
            Write-ErrorLog "Failed to install kitex: $_"
        }
    }

    # 检查模块是否存在
    try {
        $moduleList = go list -m all 2>$null
        if ($moduleList -match "github.com/cloudwego/kitex") {
            Write-InfoLog "module kitex already exist, skipping go get ..."
        }
        else {
            Write-InfoLog "module kitex not exist，ready get it..."
            go get -u github.com/cloudwego/kitex@latest
        }
    }
    catch {
        Write-WarnLog "Failed to check kitex module: $_"
    }

    Write-InfoLog "Using kitex: $KitexCmd"
}

# 从文件路径提取服务名（不带后缀）
function Get-ServiceName {
    param([string]$FilePath)
    
    $filename = Split-Path $FilePath -Leaf
    $serviceName = [System.IO.Path]::GetFileNameWithoutExtension($filename)
    return $serviceName
}

# 通用生成函数
function Invoke-KitexGeneration {
    param(
        [string]$Type,
        [string]$Message,
        [string]$ExtraArgs,
        [string]$IdlFile,
        [string]$GenPath = "",
        [bool]$UseService = $false  # 新增：是否使用服务名参数
    )

    Write-InfoLog $Message

    $baseArgs = "-module $BASE_MODULE"
    if ($GenPath -ne "") {
        $baseArgs = "$baseArgs -gen-path $GenPath"
    }

    # 如果启用服务模式，从文件名生成服务名
    if ($UseService) {
        $serviceName = Get-ServiceName $IdlFile
        $baseArgs = "$baseArgs -service $serviceName"
        Write-InfoLog "Generating service: $serviceName"
    }

    if (-not (Test-Path $IdlFile -PathType Leaf)) {
        Write-WarnLog "IDL file not found: $IdlFile, skipping..."
        return $false
    }

    Write-InfoLog "Generating from: $IdlFile"
    
    # 构建完整的命令
    $fullArgs = @()
    $baseArgs -split ' ' | Where-Object { $_ -ne "" } | ForEach-Object { $fullArgs += $_ }
    if ($ExtraArgs -ne "") {
        $ExtraArgs -split ' ' | Where-Object { $_ -ne "" } | ForEach-Object { $fullArgs += $_ }
    }
    $fullArgs += $IdlFile

    Write-InfoLog "Run command: $KitexCmd $($fullArgs -join ' ')"

    try {
        & $KitexCmd @fullArgs
        if ($LASTEXITCODE -ne 0) {
            Write-ErrorLog "Failed to generate $Type from $IdlFile"
            return $false
        }
    }
    catch {
        Write-ErrorLog "Failed to generate $Type from $IdlFile, error: $_"
        return $false
    }

    Write-InfoLog "Successfully generated $Type from $IdlFile"
    return $true
}

# 遍历生成所有 thrift 文件（默认不加 -service）
function Generate-Thrift {
    Write-InfoLog "Scanning for thrift files in $IDL_DIR/client..."

    $clientIdlPath = Join-Path $IDL_DIR "client"
    if (-not (Test-Path $clientIdlPath -PathType Container)) {
        Write-WarnLog "IDL directory '$clientIdlPath' not found, skipping thrift generation"
        return $true
    }

    if (-not (Test-Path $CLIENT_DIR -PathType Container)) {
        New-Item -ItemType Directory -Path $CLIENT_DIR -Force | Out-Null
    }

    $thriftFiles = Get-ChildItem -Path $clientIdlPath -Filter "*.thrift" -File -Recurse | Select-Object -ExpandProperty FullName

    if ($thriftFiles.Count -eq 0) {
        Write-WarnLog "No thrift files found in $clientIdlPath"
        return $true
    }

    Write-InfoLog "Found $($thriftFiles.Count) thrift file(s):"
    foreach ($file in $thriftFiles) {
        Write-InfoLog "  - $file"
    }

    $successCount = 0
    $failCount = 0

    foreach ($thriftFile in $thriftFiles) {
        Write-InfoLog "Processing: $thriftFile"

        if (Invoke-KitexGeneration -Type "thrift"` + " `" + `
			-Message "Generate thrift from $thriftFile..."` + " `" + `
            -ExtraArgs "-thrift frugal_tag -invoker"` + " `" + `
			-IdlFile $thriftFile` + " `" + `
            -GenPath $CLIENT_DIR) {
            $successCount++
        }
        else {
            $failCount++
        }
    }

    Write-InfoLog "Thrift generation completed: $successCount successful, $failCount failed"
    return $failCount -eq 0
}

# 生成 thrift 服务端代码（新增：带 -service 参数，基于文件名）
function Generate-ThriftService {
    Write-InfoLog "Scanning for thrift files in $IDL_DIR/server (with service mode)..."

    $serverIdlPath = Join-Path $IDL_DIR "server"
    if (-not (Test-Path $serverIdlPath -PathType Container)) {
        Write-WarnLog "IDL directory '$serverIdlPath' not found, skipping thrift service generation"
        return $true
    }

    if (-not (Test-Path $SERVER_DIR -PathType Container)) {
        New-Item -ItemType Directory -Path $SERVER_DIR -Force | Out-Null
    }

    $thriftFiles = Get-ChildItem -Path $serverIdlPath -Filter "*.thrift" -File -Recurse | Select-Object -ExpandProperty FullName

    if ($thriftFiles.Count -eq 0) {
        Write-WarnLog "No thrift files found in $serverIdlPath for service generation"
        return $true
    }

    $successCount = 0
    $failCount = 0

    foreach ($thriftFile in $thriftFiles) {
        Write-InfoLog "Processing (service): $thriftFile"

        if (Invoke-KitexGeneration -Type "thrift service"` + " `" + `
			-Message "Generate thrift service from $thriftFile..."` + " `" + `
            -ExtraArgs "-thrift frugal_tag -invoker"` + " `" + `
			-IdlFile $thriftFile` + " `" + `
            -GenPath $SERVER_DIR` + " `" + `
			-UseService $true) {
			$successCount++
		}
		else {
			$failCount++
		}
	}

	Write-InfoLog "Thrift service generation completed: $successCount successful, $failCount failed"
	return $failCount -eq 0
}

# 遍历生成所有 slim thrift 文件（默认不加 -service）
function Generate-ThriftSlim {
    Write-InfoLog "Scanning for thrift files in $IDL_DIR/client (slim mode)..."

    $clientIdlPath = Join-Path $IDL_DIR "client"
    if (-not (Test-Path $clientIdlPath -PathType Container)) {
        Write-WarnLog "IDL directory '$clientIdlPath' not found, skipping slim thrift generation"
        return $true
    }

    if (-not (Test-Path $CLIENT_DIR -PathType Container)) {
        New-Item -ItemType Directory -Path $CLIENT_DIR -Force | Out-Null
    }

    $thriftFiles = Get-ChildItem -Path $clientIdlPath -Filter "*.thrift" -File -Recurse | Select-Object -ExpandProperty FullName

    if ($thriftFiles.Count -eq 0) {
        Write-WarnLog "No thrift files found in $clientIdlPath for slim generation"
        return $true
    }

    $successCount = 0
    $failCount = 0

    foreach ($thriftFile in $thriftFiles) {
        Write-InfoLog "Processing (slim): $thriftFile"

        if (Invoke-KitexGeneration -Type "thrift(slim)"` + " `" + `
			-Message "Generate slim thrift from $thriftFile..."` + " `" + `
            -ExtraArgs "-thrift frugal_tag -thrift template=slim"` + " `" + `
			-IdlFile $thriftFile` + " `" + `
            -GenPath "$CLIENT_DIR/slim") {
            $successCount++
        }
        else {
            $failCount++
        }
    }

    Write-InfoLog "Slim thrift generation completed: $successCount successful, $failCount failed"
    return $failCount -eq 0
}

# 生成 slim thrift 服务端代码（新增：带 -service 参数，基于文件名）
function Generate-ThriftSlimService {
    Write-InfoLog "Scanning for thrift files in $IDL_DIR/server (slim mode with service)..."

    $serverIdlPath = Join-Path $IDL_DIR "server"
    if (-not (Test-Path $serverIdlPath -PathType Container)) {
        Write-WarnLog "IDL directory '$serverIdlPath' not found, skipping slim thrift service generation"
        return $true
    }

    if (-not (Test-Path $SERVER_DIR -PathType Container)) {
        New-Item -ItemType Directory -Path $SERVER_DIR -Force | Out-Null
    }

    $thriftFiles = Get-ChildItem -Path $serverIdlPath -Filter "*.thrift" -File -Recurse | Select-Object -ExpandProperty FullName

    if ($thriftFiles.Count -eq 0) {
        Write-WarnLog "No thrift files found in $serverIdlPath for slim service generation"
        return $true
    }

    $successCount = 0
    $failCount = 0

    foreach ($thriftFile in $thriftFiles) {
        Write-InfoLog "Processing (slim service): $thriftFile"

        if (Invoke-KitexGeneration -Type "thrift(slim) service"` + " `" + `
			-Message "Generate slim thrift service from $thriftFile..."` + " `" + `
            -ExtraArgs "-thrift frugal_tag -thrift template=slim"` + " `" + `
			-IdlFile $thriftFile` + " `" + `
            -GenPath "$SERVER_DIR/slim"` + " `" + `
			-UseService $true) {
			$successCount++
		}
		else {
			$failCount++
		}
	}

	Write-InfoLog "Slim thrift service generation completed: $successCount successful, $failCount failed"
	return $failCount -eq 0
}

# Protobuf 生成函数（默认不加 -service）
function Generate-Protobuf {
    Write-InfoLog "Scanning for proto files in $IDL_DIR/client..."

    $clientIdlPath = Join-Path $IDL_DIR "client"
    if (-not (Test-Path $clientIdlPath -PathType Container)) {
        Write-WarnLog "IDL directory '$clientIdlPath' not found, skipping protobuf generation"
        return $true
    }

    if (-not (Test-Path $CLIENT_DIR -PathType Container)) {
        New-Item -ItemType Directory -Path $CLIENT_DIR -Force | Out-Null
    }

    $protoFiles = Get-ChildItem -Path $clientIdlPath -Filter "*.proto" -File -Recurse | Select-Object -ExpandProperty FullName

    if ($protoFiles.Count -eq 0) {
        Write-WarnLog "No proto files found in $clientIdlPath"
        return $true
    }

    Write-InfoLog "Found $($protoFiles.Count) proto file(s):"
    foreach ($file in $protoFiles) {
        Write-InfoLog "  - $file"
    }

    $successCount = 0
    $failCount = 0

    foreach ($protoFile in $protoFiles) {
        Write-InfoLog "Processing: $protoFile"

        if (Invoke-KitexGeneration -Type "protobuf"` + " `" + `
			-Message "Generate protobuf from $protoFile..."` + " `" + `
            -ExtraArgs ""` + " `" + `
			-IdlFile $protoFile` + " `" + `
            -GenPath $CLIENT_DIR) {
            $successCount++
        }
        else {
            $failCount++
        }
    }

    Write-InfoLog "Protobuf generation completed: $successCount successful, $failCount failed"
    return $failCount -eq 0
}

# 生成 protobuf 服务端代码（新增：带 -service 参数，基于文件名）
function Generate-ProtobufService {
    Write-InfoLog "Scanning for proto files in $IDL_DIR/server (with service mode)..."

    $serverIdlPath = Join-Path $IDL_DIR "server"
    if (-not (Test-Path $serverIdlPath -PathType Container)) {
        Write-WarnLog "IDL directory '$serverIdlPath' not found, skipping protobuf service generation"
        return $true
    }

    if (-not (Test-Path $SERVER_DIR -PathType Container)) {
        New-Item -ItemType Directory -Path $SERVER_DIR -Force | Out-Null
    }

    $protoFiles = Get-ChildItem -Path $serverIdlPath -Filter "*.proto" -File -Recurse | Select-Object -ExpandProperty FullName

    if ($protoFiles.Count -eq 0) {
        Write-WarnLog "No proto files found in $serverIdlPath for service generation"
        return $true
    }

    $successCount = 0
    $failCount = 0

    foreach ($protoFile in $protoFiles) {
        Write-InfoLog "Processing (service): $protoFile"

        if (Invoke-KitexGeneration -Type "protobuf service" ` + "`" + `
			-Message "Generate protobuf service from $protoFile..."` + " `" + `
            -ExtraArgs ""` + " `" + `
	      	-IdlFile $protoFile` + " `" + `
            -GenPath $SERVER_DIR` + " `" + `
			-UseService $true) {
            $successCount++
        }
        else {
            $failCount++
        }
    }

    Write-InfoLog "Protobuf service generation completed: $successCount successful, $failCount failed"
    return $failCount -eq 0
}

# 子模块重新生成函数
function Regenerate-Submodule {
    param(
        [string]$Path,
        [string]$Module,
        [string]$ThriftFile,
        [bool]$UseService = $false
    )

    if (-not (Test-Path $Path -PathType Container)) {
        New-Item -ItemType Directory -Path $Path -Force | Out-Null
    }

    $originalDir = Get-Location
    Set-Location $Path

    Write-InfoLog "Executing kitex command in $Path with module: $Module and thrift file: $ThriftFile"

    if (-not (Test-Path $ThriftFile -PathType Leaf)) {
        Write-WarnLog "Thrift file not found: $ThriftFile in $Path, skipping..."
        Set-Location $originalDir
        return $true
    }

    # 构建 kitex 命令
    $kitexArgs = @("-module", $Module)
    if ($UseService) {
        $serviceName = Get-ServiceName $ThriftFile
        $kitexArgs += "-service"
        $kitexArgs += $serviceName
    }
    $kitexArgs += $ThriftFile

    # 执行 kitex 命令
    try {
        & $KitexCmd @kitexArgs
        if ($LASTEXITCODE -ne 0) {
            Write-ErrorLog "Failed to generate for $Path"
            Set-Location $originalDir
            return $false
        }
    }
    catch {
        Write-ErrorLog "Failed to generate for $Path, error: $_"
        Set-Location $originalDir
        return $false
    }

    # 更新依赖
    Write-InfoLog "Updating dependencies for $Path..."
    try {
        go get github.com/cloudwego/kitex@latest
        go mod tidy
    }
    catch {
        Write-WarnLog "Failed to update dependencies for $Path, error: $_"
    }

    Set-Location $originalDir
    Write-InfoLog "Successfully regenerated $Path"
    return $true
}

# 主执行函数
function Main {
    Write-InfoLog "Starting IDL code generation..."
    Check-Kitex

    # 生成主 IDL 文件（默认不加 -service）
    if (-not (Generate-Thrift)) {
        Write-WarnLog "Thrift generation had issues, continuing..."
    }

    # if (-not (Generate-ThriftSlim)) {
    #     Write-WarnLog "Thrift slim generation had issues, continuing..."
    # }

    if (-not (Generate-Protobuf)) {
        Write-WarnLog "Protobuf generation had issues, continuing..."
    }

    # 如果需要生成服务端代码，取消注释下面的调用：
    if (-not (Generate-ThriftService)) {
        Write-WarnLog "Thrift service generation had issues, continuing..."
    }

    # if (-not (Generate-ThriftSlimService)) {
    #     Write-WarnLog "Thrift slim service generation had issues, continuing..."
    # }

    if (-not (Generate-ProtobufService)) {
        Write-WarnLog "Protobuf service generation had issues, continuing..."
    }

    # 子模块配置数组
    $submodules = @(
        # TODO: 模板格式-> @($path, $module, $thriftFile, $useService)
        # 例如："basic/example_shop", "example_shop", "idl/item.thrift", $false           # 仅客户端
        # 例如："basic/example_shop", "example_shop", "idl/item.thrift", $true      # 服务端（基于文件名）
    )

    # 遍历所有子模块（仅在数组非空时执行）
    if ($submodules.Count -gt 0) {
        foreach ($submod in $submodules) {
            $path, $module, $thriftFile, $useService = $submod
            Regenerate-Submodule -Path $path -Module $module -ThriftFile $thriftFile -UseService $useService
        }
    }
    else {
        Write-InfoLog "No submodules configured, skipping submodule generation."
    }

    Write-InfoLog "All IDL code generation completed!"
}

# 设置错误处理
$ErrorActionPreference = "Stop"

# 执行主函数
Main`

const SCRIPT_KITEX = `#!/bin/bash

set -euo pipefail

# 配置变量
KITEX_CMD="${1:-kitex}"
BASE_MODULE="{{.ModulePath}}"
ROOT_DIR=$(pwd)
IDL_DIR="idl"
CLIENT_DIR="client/kitex"
SERVER_DIR="server/kitex"

# 颜色输出函数
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

log_info() { echo -e "${GREEN}[INFO]${NC} $1"; }
log_warn() { echo -e "${YELLOW}[WARN]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }

# 检查 kitex 是否可用
check_kitex() {
    if ! command -v "$KITEX_CMD" &> /dev/null; then
        log_info "Installing kitex..."
        go install github.com/cloudwego/kitex/tool/cmd/kitex@latest
        if ! command -v "$KITEX_CMD" &> /dev/null; then
            log_error "kitex command not found and installation failed"
            exit 1
        fi
    fi

    if go list -m all | grep "github.com/cloudwego/kitex" &> /dev/null; then
        log_info "module kitex already exist, skipping go get ..."
    else
        log_info "module kitex not exist，ready get it..."
        go get -u github.com/cloudwego/kitex@latest
    fi

    log_info "Using kitex: $(which $KITEX_CMD)"
}

# 从文件路径提取服务名（不带后缀）
get_service_name() {
    local file_path="$1"
    local filename=$(basename "$file_path")
    local service_name="${filename%.*}"  # 移除文件后缀
    echo "$service_name"
}

# 通用生成函数
gen_kitex() {
    local type="$1"
    local msg="$2"
    local extra_args="$3"
    local idl_file="$4"
    local gen_path="${5:-}"
    local use_service="${6:-false}"  # 新增：是否使用服务名参数

    log_info "$msg"

    local base_args="-module $BASE_MODULE"
    if [[ -n "$gen_path" ]]; then
        base_args="$base_args -gen-path $gen_path"
    fi

    # 如果启用服务模式，从文件名生成服务名
    if [[ "$use_service" == "true" ]]; then
        local service_name=$(get_service_name "$idl_file")
        base_args="$base_args -service $service_name"
        log_info "Generating service: $service_name"
    fi

    if [[ ! -f "$idl_file" ]]; then
        log_warn "IDL file not found: $idl_file, skipping..."
        return 1
    fi

    log_info "Generating from: $idl_file"
    log_info "Run command: $KITEX_CMD $base_args $extra_args "$idl_file""

    if ! $KITEX_CMD $base_args $extra_args "$idl_file"; then
        log_error "Failed to generate $type from $idl_file"
        return 1
    fi

    log_info "Successfully generated $type from $idl_file"
    return 0
}

# 遍历生成所有 thrift 文件（默认不加 -service）
gen_thrift() {
    log_info "Scanning for thrift files in $IDL_DIR/client..."

    if [[ ! -d "$IDL_DIR/client" ]]; then
        log_warn "IDL directory '$IDL_DIR/client' not found, skipping thrift generation"
        return 0
    fi

    if [[ ! -d "$CLIENT_DIR" ]]; then
        mkdir -p $CLIENT_DIR
    fi

    local thrift_files=()
    while IFS= read -r -d $'\0' file; do
        thrift_files+=("$file")
    done < <(find "$IDL_DIR/client" -name "*.thrift" -type f -print0)

    if [[ ${#thrift_files[@]} -eq 0 ]]; then
        log_warn "No thrift files found in $IDL_DIR/client"
        return 0
    fi

    log_info "Found ${#thrift_files[@]} thrift file(s):"
    for file in "${thrift_files[@]}"; do
        log_info "  - $file"
    done

    local success_count=0
    local fail_count=0

    for thrift_file in "${thrift_files[@]}"; do
        log_info "Processing: $thrift_file"

        if gen_kitex "thrift" "Generate thrift from $thrift_file..." "-thrift frugal_tag -invoker" "$thrift_file" $CLIENT_DIR; then
            ((success_count++))
        else
            ((fail_count++))
        fi
    done

    log_info "Thrift generation completed: $success_count successful, $fail_count failed"

    if [[ $fail_count -gt 0 ]]; then
        return 1
    fi
    return 0
}

# 生成 thrift 服务端代码（新增：带 -service 参数，基于文件名）
gen_thrift_service() {
    log_info "Scanning for thrift files in $IDL_DIR/server (with service mode)..."

    if [[ ! -d "$IDL_DIR/server" ]]; then
        log_warn "IDL directory '$IDL_DIR/server' not found, skipping thrift service generation"
        return 0
    fi

    if [[ ! -d "$SERVER_DIR" ]]; then
        mkdir -p $SERVER_DIR
    fi

    local thrift_files=()
    while IFS= read -r -d $'\0' file; do
        thrift_files+=("$file")
    done < <(find "$IDL_DIR/server" -name "*.thrift" -type f -print0)

    if [[ ${#thrift_files[@]} -eq 0 ]]; then
        log_warn "No thrift files found in $IDL_DIR/server for service generation"
        return 0
    fi

    local success_count=0
    local fail_count=0

    for thrift_file in "${thrift_files[@]}"; do
        log_info "Processing (service): $thrift_file"

        if gen_kitex "thrift service" "Generate thrift service from $thrift_file..." "-thrift frugal_tag -invoker" "$thrift_file" $SERVER_DIR "true"; then
            ((success_count++))
        else
            ((fail_count++))
        fi
    done

    log_info "Thrift service generation completed: $success_count successful, $fail_count failed"

    if [[ $fail_count -gt 0 ]]; then
        return 1
    fi
    return 0
}

# 遍历生成所有 slim thrift 文件（默认不加 -service）
gen_thrift_slim() {
    log_info "Scanning for thrift files in $IDL_DIR/client (slim mode)..."

    if [[ ! -d "$IDL_DIR/client" ]]; then
        log_warn "IDL directory '$IDL_DIR/client' not found, skipping slim thrift generation"
        return 0
    fi

    if [[ ! -d "$CLIENT_DIR" ]]; then
        mkdir -p $CLIENT_DIR
    fi

    local thrift_files=()
    while IFS= read -r -d $'\0' file; do
        thrift_files+=("$file")
    done < <(find "$IDL_DIR/client" -name "*.thrift" -type f -print0)

    if [[ ${#thrift_files[@]} -eq 0 ]]; then
        log_warn "No thrift files found in $IDL_DIR/client for slim generation"
        return 0
    fi

    local success_count=0
    local fail_count=0

    for thrift_file in "${thrift_files[@]}"; do
        log_info "Processing (slim): $thrift_file"

        if gen_kitex "thrift(slim)" "Generate slim thrift from $thrift_file..." "-thrift frugal_tag -thrift template=slim" "$thrift_file" "$CLIENT_DIR/slim"; then
            ((success_count++))
        else
            ((fail_count++))
        fi
    done

    log_info "Slim thrift generation completed: $success_count successful, $fail_count failed"

    if [[ $fail_count -gt 0 ]]; then
        return 1
    fi
    return 0
}

# 生成 slim thrift 服务端代码（新增：带 -service 参数，基于文件名）
gen_thrift_slim_service() {
    log_info "Scanning for thrift files in $IDL_DIR/server (slim mode with service)..."

    if [[ ! -d "$IDL_DIR/server" ]]; then
        log_warn "IDL directory '$IDL_DIR/server' not found, skipping slim thrift service generation"
        return 0
    fi

    if [[ ! -d "$SERVER_DIR" ]]; then
        mkdir -p $SERVER_DIR
    fi

    local thrift_files=()
    while IFS= read -r -d $'\0' file; do
        thrift_files+=("$file")
    done < <(find "$IDL_DIR/server" -name "*.thrift" -type f -print0)

    if [[ ${#thrift_files[@]} -eq 0 ]]; then
        log_warn "No thrift files found in $IDL_DIR/server for slim service generation"
        return 0
    fi

    local success_count=0
    local fail_count=0

    for thrift_file in "${thrift_files[@]}"; do
        log_info "Processing (slim service): $thrift_file"

        if gen_kitex "thrift(slim) service" "Generate slim thrift service from $thrift_file..." "-thrift frugal_tag -thrift template=slim" "$thrift_file" "$SERVER_DIR/slim" "true"; then
            ((success_count++))
        else
            ((fail_count++))
        fi
    done

    log_info "Slim thrift service generation completed: $success_count successful, $fail_count failed"

    if [[ $fail_count -gt 0 ]]; then
        return 1
    fi
    return 0
}

# Protobuf 生成函数（默认不加 -service）
gen_protobuf() {
    log_info "Scanning for proto files in $IDL_DIR/client..."

    if [[ ! -d "$IDL_DIR/client" ]]; then
        log_warn "IDL directory '$IDL_DIR/client' not found, skipping protobuf generation"
        return 0
    fi

    if [[ ! -d "$CLIENT_DIR" ]]; then
        mkdir -p $CLIENT_DIR
    fi

    local proto_files=()
    while IFS= read -r -d $'\0' file; do
        proto_files+=("$file")
    done < <(find "$IDL_DIR/client" -name "*.proto" -type f -print0)

    if [[ ${#proto_files[@]} -eq 0 ]]; then
        log_warn "No proto files found in $IDL_DIR/client"
        return 0
    fi

    log_info "Found ${#proto_files[@]} proto file(s):"
    for file in "${proto_files[@]}"; do
        log_info "  - $file"
    done

    local success_count=0
    local fail_count=0

    for proto_file in "${proto_files[@]}"; do
        log_info "Processing: $proto_file"

        if gen_kitex "protobuf" "Generate protobuf from $proto_file..." "" "$proto_file" $CLIENT_DIR; then
            ((success_count++))
        else
            ((fail_count++))
        fi
    done

    log_info "Protobuf generation completed: $success_count successful, $fail_count failed"

    if [[ $fail_count -gt 0 ]]; then
        return 1
    fi
    return 0
}

# 生成 protobuf 服务端代码（新增：带 -service 参数，基于文件名）
gen_protobuf_service() {
    log_info "Scanning for proto files in $IDL_DIR/server (with service mode)..."

    if [[ ! -d "$IDL_DIR/server" ]]; then
        log_warn "IDL directory '$IDL_DIR/server' not found, skipping protobuf service generation"
        return 0
    fi

    if [[ ! -d "$SERVER_DIR" ]]; then
        mkdir -p $SERVER_DIR
    fi

    local proto_files=()
    while IFS= read -r -d $'\0' file; do
        proto_files+=("$file")
    done < <(find "$IDL_DIR/server" -name "*.proto" -type f -print0)

    if [[ ${#proto_files[@]} -eq 0 ]]; then
        log_warn "No proto files found in $IDL_DIR/server for service generation"
        return 0
    fi

    local success_count=0
    local fail_count=0

    for proto_file in "${proto_files[@]}"; do
        log_info "Processing (service): $proto_file"

        if gen_kitex "protobuf service" "Generate protobuf service from $proto_file..." "" "$proto_file" $SERVER_DIR "true"; then
            ((success_count++))
        else
            ((fail_count++))
        fi
    done

    log_info "Protobuf service generation completed: $success_count successful, $fail_count failed"

    if [[ $fail_count -gt 0 ]]; then
        return 1
    fi
    return 0
}

# 子模块重新生成函数
regenerate_submod() {
    local path="$1"
    local module="$2"
    local thrift_file="$3"
    local use_service="${4:-false}"  # 修改：是否使用服务模式

    if [[ ! -d "$path" ]]; then
        mkdir -p $path
    fi

    local original_dir=$(pwd)
    cd "$path"

    log_info "Executing kitex command in $path with module: $module and thrift file: $thrift_file"

    if [[ ! -f "$thrift_file" ]]; then
        log_warn "Thrift file not found: $thrift_file in $path, skipping..."
        cd "$original_dir"
        return 0
    fi

    # 构建 kitex 命令
    local kitex_cmd="$KITEX_CMD -module $module"
    if [[ "$use_service" == "true" ]]; then
        local service_name=$(get_service_name "$thrift_file")
        kitex_cmd="$kitex_cmd -service $service_name"
    fi
    kitex_cmd="$kitex_cmd $thrift_file"

    # 执行 kitex 命令
    if ! eval $kitex_cmd; then
        log_error "Failed to generate for $path"
        cd "$original_dir"
        return 1
    fi

    # 更新依赖
    log_info "Updating dependencies for $path..."
    go get github.com/cloudwego/kitex@latest
    go mod tidy

    cd "$original_dir"
    log_info "Successfully regenerated $path"
    return 0
}

# 主执行函数
main() {
    log_info "Starting IDL code generation..."
    check_kitex

    # 生成主 IDL 文件（默认不加 -service）
    gen_thrift || log_warn "Thrift generation had issues, continuing..."
#    gen_thrift_slim || log_warn "Thrift slim generation had issues, continuing..."
    gen_protobuf || log_warn "Protobuf generation had issues, continuing..."

    # 如果需要生成服务端代码，取消注释下面的调用：
     gen_thrift_service || log_warn "Thrift service generation had issues, continuing..."  # 生成 thrift 服务端（基于文件名）
#     gen_thrift_slim_service || log_warn "Thrift slim service generation had issues, continuing..."  # 生成 slim thrift 服务端（基于文件名）
     gen_protobuf_service || log_warn "Protobuf service generation had issues, continuing..." # 生成 protobuf 服务端（基于文件名）

    # 子模块配置数组
    declare -a submodules=(
        # TODO: 模板格式-> $path:$module:$thrift_file:$use_service(可选)
        # 例如："basic/example_shop:example_shop:idl/item.thrift"           # 仅客户端
        # 例如："basic/example_shop:example_shop:idl/item.thrift:true"      # 服务端（基于文件名）
    )

    # 遍历所有子模块（仅在数组非空时执行）
    if [[ ${#submodules[@]} -gt 0 ]]; then
       for submod in "${submodules[@]}"; do
           IFS=':' read -r path module thrift_file use_service <<< "$submod"
           regenerate_submod "$path" "$module" "$thrift_file" "$use_service"
       done
    else
       log_info "No submodules configured, skipping submodule generation."
    fi

    log_info "All IDL code generation completed!"
}

# 执行主函数
main "$@"
`
