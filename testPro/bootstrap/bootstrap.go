package bootstrap

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
	_ "github.com/jasonlabz/potato/log"
	"github.com/jasonlabz/potato/utils"
)

var (
	Username string
	Password string
)

func MustInit(ctx context.Context) {
	initResource(ctx)
	// 初始化加解秘钥
	initCrypto(ctx)
	// 初始化DB
	initDB(ctx)
	// 初始化配置文件
	initConfig(ctx)
}

func initResource(_ context.Context) {
	Username = func() string {
		user := os.Getenv("AUTH_USER")
		if user != "" {
			return user
		}
		return "admin"
	}()
	Password = func() string {
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
	conf := configx.GetConfig()
	if conf.Database == nil {
		panic("no db config")
	}
	gormConfig := &gormx.Config{}
	err := utils.CopyStruct(conf.Database, gormConfig)
	if err != nil {
		panic(err)
	}
	gormConfig.DBName = gormx.DefaultDBNameMaster
	err = gormx.InitConfig(gormConfig)
	if err != nil {
		panic(err)
	}
	//impl.SetGormDB(gormx.GetDBWithPanic(gormConfig.DBName))
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
