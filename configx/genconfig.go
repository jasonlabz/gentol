package configx

import (
	"fmt"
	"github.com/jasonlabz/gentol/gormx"
	"log"
	"os"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"
)

var DefaultPath = "conf/table.yaml"

func Init() {
	LoadConfigFromYaml(DefaultPath)
}

// DBTableInfo 连接配置
type DBTableInfo struct {
	DBName      string       `json:"db_name" yaml:"db_name"`
	DBType      string       `json:"db_type" yaml:"db_type"`
	DSN         string       `json:"dsn" yaml:"dsn"`
	OnlyModel   bool         `json:"only_model" yaml:"only_model"`
	ServicePath string       `json:"service_path" yaml:"service_path"`
	ModelPath   string       `json:"model_path" yaml:"model_path"`
	DaoPath     string       `json:"dao_path" yaml:"dao_path"`
	Tables      []*TableInfo `json:"tables" yaml:"tables"`

	ModelModule string
	DaoModule   string
	// DSN 可选
	Host     string `json:"host" yaml:"host"`
	Port     int    `json:"port" yaml:"port"`
	User     string `json:"user" yaml:"user"`
	Password string `json:"password" yaml:"password"`
	Database string `json:"database" yaml:"database"`
}

func (c *DBTableInfo) GenDSN() (dsn string) {
	if c.DSN != "" {
		return c.DSN
	}

	dbName := c.Database
	if dbName == "" {
		dbName = c.DBName
	}
	dsnTemplate, ok := gormx.DatabaseDsnMap[gormx.DBType(c.DBType)]
	if !ok {
		return
	}
	dsn = fmt.Sprintf(dsnTemplate, c.User, c.Password, c.Host, c.Port, dbName)

	c.DSN = dsn
	return
}

// TableInfo 连接配置
type TableInfo struct {
	SchemaName string   `json:"schema_name" yaml:"schema_name"`
	TableList  []string `json:"table_list" yaml:"table_list"`
}

type config struct {
	Configs               []*DBTableInfo `json:"configs" yaml:"configs"`
	JsonFormat            string         `json:"json_format" yaml:"json_format"`
	ProtobufFormat        string         `json:"protobuf_format" yaml:"protobuf_format"`
	GoModule              string         `json:"module" yaml:"module"`
	UseSQLNullable        bool           `json:"use_sql_nullable" yaml:"use_sql_nullable"`
	RunGoFmt              bool           `json:"rungofmt" yaml:"rungofmt"`
	AddProtobufAnnotation bool           `json:"addProtobufAnnotation" yaml:"addProtobufAnnotation"`
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
