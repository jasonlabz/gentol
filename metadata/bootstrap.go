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
	"os"
	"path/filepath"

	"github.com/jasonlabz/potato/configx"
	"github.com/jasonlabz/potato/configx/file"
	"github.com/jasonlabz/potato/cryptox"
	"github.com/jasonlabz/potato/cryptox/aes"
	"github.com/jasonlabz/potato/cryptox/des"
	"github.com/jasonlabz/potato/es"
	"github.com/jasonlabz/potato/goredis"
	"github.com/jasonlabz/potato/gormx"
	"github.com/jasonlabz/potato/httpx"
	"github.com/jasonlabz/potato/log"
	"github.com/jasonlabz/potato/rabbitmqx"
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
	dbConf := configx.GetConfig().DataSource
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
	resource.RMQClient = rabbitmqx.GetRabbitMQOperator()
}

func initRedis(_ context.Context) {
	// 走默认初始化逻辑
	resource.RedisClient = goredis.GetRedisOperator()
}

func initES(_ context.Context) {
	// 走默认初始化逻辑
	resource.EsClient = es.GetESOperator()
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
