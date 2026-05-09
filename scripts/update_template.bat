@echo off
REM update_template.bat - 更新嵌入式模板
REM 用法: scripts\update_template.bat [template_repo_url]
REM
REM 从模板仓库 clone，打包为 template.tar.gz，
REM 放到 embedded/ 目录供 //go:embed 编译进 gentol 二进制

setlocal enabledelayedexpansion

set "REPO_URL=%~1"
if "%REPO_URL%"=="" set "REPO_URL=https://github.com/jasonlabz/generate-example-project.git"

set "SCRIPT_DIR=%~dp0"
set "PROJECT_DIR=%SCRIPT_DIR%.."
set "OUTPUT=%PROJECT_DIR%\embedded\template.tar.gz"

echo Cloning template from: %REPO_URL%

set "TMP_DIR=%TEMP%\gentol-template-%RANDOM%"
mkdir "%TMP_DIR%"

git clone --depth 1 %REPO_URL% "%TMP_DIR%\template"
if errorlevel 1 (
    echo Clone failed!
    rmdir /s /q "%TMP_DIR%"
    exit /b 1
)

echo Creating template.tar.gz...
REM 需要系统有 tar 命令（Windows 10 1803+ 内置）
cd /d "%TMP_DIR%\template"
tar czf "%OUTPUT%" --exclude=".git" --exclude="go.sum" .

if errorlevel 1 (
    echo tar failed!
    cd /d "%PROJECT_DIR%"
    rmdir /s /q "%TMP_DIR%"
    exit /b 1
)

cd /d "%PROJECT_DIR%"
echo Template archive created: %OUTPUT%
echo.
echo Now rebuild gentol to embed the updated template:
echo   go build .

rmdir /s /q "%TMP_DIR%"
