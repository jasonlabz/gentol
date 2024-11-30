# 工具介绍
### 1、 gorm gen使用```shell
## install gen tool (should be installed to ~/go/bin, make sure ~/go/bin is in your path.
## go version < 1.17
$ go get -u github.com/smallnest/gen

## go version == 1.17
$ go install github.com/smallnest/gen@v0.9.29

## generate code based on the sqlite database (project will be contained within the ./example dir)
$ gen --sqltype=postgres  --connstr "host=localhost user=postgres password=halojeff dbname=postgres port=8432 sslmode=disable"  --database postgres --table user --out ./dal --json --gorm --guregu --run-gofmt --json-fmt=snake --overwrite
```
### 1、 gentol使用
```shell
## install gentol
$ go install github.com/jasonlabz/gentol@master

## generate code based on the sqlite database (project will be contained within the ./example dir)
$ gentol --db_type="postgres" --dsn="user=postgres password=XXXXX host=127.0.0.1 port=8432 dbname=dbName sslmode=disable TimeZone=Asia/Shanghai" --schema="public" --table="table1,table2" --only_model --gen_hook
```

### 2、swagger使用
```shell
## swagger 依赖
go get "github.com/swaggo/files"
go get "github.com/swaggo/gin-swagger"


## swagger 命令行工具
go install github.com/swaggo/swag/cmd/swag@v1.8.12

###注释文档 main函数
// @title 这里写标题
// @version 这里写版本号
// @description 这里写描述信息
// @termsOfService http://swagger.io/terms/

// @contact.name 这里写联系人信息
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host 这里写接口服务的host
// @BasePath 这里写base path（eg：/api/v1）
func main() {}

### 接口层 controller
// @Summary 升级版帖子列表接口
// @Description 可按社区按时间或分数排序查询帖子列表接口
// @Tags 帖子相关接口
// @Accept application/json
// @Produce application/json
// @Param Authorization header string false "Bearer 用户令牌"
// @Param object query models.ParamArtList(请求参数结构体) false "查询参数"
// @Security ApiKeyAuth
// @Success 200 {object} _ResponseArtList
// @Router /接口路由 [请求类型]
func GetArt(c *gin.Context) {}

### 结构体 struct
// 文章列表接口数据信息
type _ResponseArticle struct {
	Code    int               `json:"code"` // 业务状态码
	Message string            `json:"message"` // 提示信息
	Data    *[]model.Article  `json:"data"` // 数据
}

### 生成文档，执行：
swag init
}
```
