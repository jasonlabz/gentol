package main

import (
	"github.com/jasonlabz/gentol/configx"
	"github.com/pborman/getopt/v2"
	"sync"
)

var once sync.Once

var (
	dbType = getopt.StringLong("db_type", 0, "mysql", "database type such as [mysql, sqlserver, postgres, oracle, greenplum etc. ]")
	dsn    = getopt.StringLong("dsn", 0, "", "option database connection string, or provide host and port ...")

	host     = getopt.StringLong("host", 'h', "", "db host, if there is a dsn, ignore it")
	port     = getopt.IntLong("port", 'p', 0, "db port, if there is a dsn, ignore it")
	username = getopt.StringLong("username", 'U', "", "db username, if there is a dsn, ignore it")
	password = getopt.StringLong("password", 'P', "", "db password, if there is a dsn, ignore it")
	database = getopt.StringLong("database", 'd', "", "database to for db table")

	schema          = getopt.StringLong("schema", 's', "public", "schema to for db table")
	table           = getopt.StringLong("table", 't', "", "table name to build struct from")
	templateDir     = getopt.StringLong("template_dir", 0, "./template", "Template Dir")
	saveTemplateDir = getopt.StringLong("save", 0, "", "save templates to dir")

	modelPath   = getopt.StringLong("model", 0, "./model", "name to set for model package")
	daoPath     = getopt.StringLong("dao", 0, "./dao", "name to set for dao package")
	servicePath = getopt.StringLong("service", 0, "./service", "name to set for service package")
	grpcPath    = getopt.StringLong("grpc", 0, "./grpc", "name to set for grpc package")
	outDir      = getopt.StringLong("out", 0, ".", "output dir")

	jsonNameFormat  = getopt.StringLong("json_format", 0, "snake", "json name format [snake | camel | lower_camel | none]")
	xmlNameFormat   = getopt.StringLong("xml_format", 0, "snake", "xml name format [snake | camel | lower_camel | none]")
	protoNameFormat = getopt.StringLong("protobuf_format", 0, "snake", "proto name format [snake | camel | lower_camel | none]")
	gogoProtoImport = getopt.StringLong("gogoproto", 0, "", "location of gogo import ")

	onlyModel             = getopt.BoolLong("only_model", 0, "overwrite existing files (default)", "disable overwriting files")
	addGormAnnotation     = getopt.BoolLong("gorm", 0, "add gorm annotations (tags)", "")
	addProtobufAnnotation = getopt.BoolLong("proto", 0, "add protobuf annotations (tags)", "")
	runGoFmt              = getopt.BoolLong("gofmt", 0, "run gofmt on output dir", "")
	DefaultDBName         = "_default_db_"
)

func init() {
	exist := IsExist("./conf/table.yaml")
	if !exist {
		exist = IsExist("./table.yaml")
	}
	if exist {
		configx.Init()
	} else {
		// check args
		getopt.Lookup("db_type").SetGroup("check")
		getopt.Lookup("database").SetGroup("check")
		getopt.Lookup("table").SetGroup("check")
		getopt.Lookup("dsn").SetOptional()
		getopt.Lookup("host").SetOptional()
		getopt.Lookup("port").SetOptional()
		getopt.Lookup("username").SetOptional()
		getopt.Lookup("password").SetOptional()
		getopt.Lookup("database").SetOptional()
		getopt.RequiredGroup("check")
		getopt.ParseV2()

		// fill args
		configx.TableConfigs.AddGormAnnotation = *addGormAnnotation
		configx.TableConfigs.AddProtobufAnnotation = *addProtobufAnnotation
		configx.TableConfigs.RunGoFmt = *runGoFmt
		configx.TableConfigs.JsonFormat = *jsonNameFormat
		configx.TableConfigs.XMLFormat = *xmlNameFormat
		configx.TableConfigs.ProtobufFormat = *protoNameFormat
		configx.TableConfigs.Configs = []*configx.Database{
			{
				DBName:      DefaultDBName,
				DBType:      *dbType,
				DSN:         *dsn,
				OnlyModel:   *onlyModel,
				ModelPath:   *modelPath,
				DaoPath:     *daoPath,
				ServicePath: *servicePath,
				Host:        *host,
				Port:        *port,
				User:        *username,
				Password:    *password,
				Tables: []*configx.TableInfo{
					{
						SchemaName: *schema,
						TableName:  *table,
					},
				},
			},
		}
	}
}
