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
