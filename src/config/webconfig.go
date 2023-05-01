package config

import (
	"fmt"
	"log"
	"os"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"
)

var DefaultPath = "conf/table.yaml"

func init() {
	LoadConfigFromYaml(DefaultPath)
}

// Database 连接配置
type Database struct {
	DBName    string       `json:"db_name" yaml:"db_name"`
	DBType    string       `json:"db_type" yaml:"db_type"`
	DSN       string       `json:"dsn" yaml:"dsn"`
	OnlyModel bool         `json:"only_model" yaml:"only_model"`
	ModelPath string       `json:"model_path" yaml:"model_path"`
	DaoPath   string       `json:"dao_path" yaml:"dao_path"`
	Tables    []*TableInfo `json:"tables" yaml:"tables"`
}

// TableInfo 连接配置
type TableInfo struct {
	SchemaName string `json:"schema_name" yaml:"schema_name"`
	TableName  string `json:"table_name" yaml:"table_name"`
}

type config struct {
	Configs []*Database `json:"configs" yaml:"configs"`
}

var TableConfigs = new(config)

func GetConfig() *config {
	if TableConfigs != nil {
		return TableConfigs
	}
	return nil
}

func LoadConfigFromYaml(configPath string) {
	file, err := os.ReadFile(configPath)
	if err != nil {
		panic(err)
	}
	err = yaml.Unmarshal(file, TableConfigs)
	if err != nil {
		fmt.Println(err)
		return
	}
}

func ParseConfigByViper(configPath, configName, configType string) {
	v := viper.New()
	v.AddConfigPath(configPath)
	v.SetConfigName(configName)
	v.SetConfigType(configType)

	if err := v.ReadInConfig(); err != nil {
		panic(err)
	}
	v.WatchConfig()
	v.OnConfigChange(func(e fsnotify.Event) {
		if err := v.ReadInConfig(); err != nil {
			panic(err)
		}
	})
	//直接反序列化为Struct
	if err := v.Unmarshal(TableConfigs); err != nil {
		log.Fatal(err)
	}
	return
}
