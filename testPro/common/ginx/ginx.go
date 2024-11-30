package base

import (
	"errors"
	"net/http"
	"reflect"
	"time"

	"github.com/gin-gonic/gin"

	errors2 "github.com/jasonlabz/potato/errors"
	"github.com/jasonlabz/potato/log"
)

// Response 响应结构体
type Response struct {
	Code        int         `json:"code"`                // 错误码
	Message     string      `json:"message,omitempty"`   // 错误信息
	ErrTrace    string      `json:"err_trace,omitempty"` // 错误追踪链路信息
	Version     string      `json:"version"`             // 版本信息
	CurrentTime string      `json:"current_time"`        // 接口返回时间（当前时间）
	Data        interface{} `json:"data,omitempty"`      //返回数据
}

type ResponseWithPagination struct {
	Response
	Pagination *Pagination `json:"pagination,omitempty"`
}

// ResponseOK 返回正确结果及数据
func ResponseOK(c *gin.Context, version string, data interface{}) {
	c.JSON(prepareResponse(c, version, data, nil))
	return
}

// ResponseErr 返回错误
func ResponseErr(c *gin.Context, version string, err error) {
	c.JSON(prepareResponse(c, version, nil, err))
	return
}

// JsonResult 返回结果Json
func JsonResult(c *gin.Context, version string, data interface{}, err error) {
	c.JSON(prepareResponse(c, version, data, err))
	return
}

// PaginationResult 返回结果Json带分页
func PaginationResult(c *gin.Context, version string, data interface{}, err error, pagination *Pagination) {
	c.JSON(prepareResponseWithPagination(c, version, data, err, pagination))
	return
}

// PureJsonResult 返回结果PureJson
func PureJsonResult(c *gin.Context, version string, data interface{}, err error) {
	c.PureJSON(prepareResponse(c, version, data, err))
	return
}

// prepareResponse 准备响应信息
func prepareResponse(c *gin.Context, version string, data interface{}, err error) (int, *Response) {
	// 格式化返回数据，非数组及切片时，转为切片
	data = handleData(data)
	code := http.StatusOK
	var errCode int
	var errMessage string
	var errTrace string

	if err != nil {
		var ex *errors2.Error
		if errors.As(err, &ex) {
			errCode = ex.Code()
			errMessage = ex.Message()
			errTrace = ex.Error()
		} else {
			code = http.StatusInternalServerError
			errMessage = err.Error()
			errTrace = err.Error()
		}
		log.GetLogger().WithContext(c).Error("        "+errTrace,
			log.Int("err_code", errCode), log.String("err_message", errMessage))
	}
	// 组装响应结果
	resp := &Response{
		Code:        errCode,
		Message:     errMessage,
		ErrTrace:    errTrace,
		Version:     version,
		Data:        data,
		CurrentTime: time.Now().Format(time.DateTime),
	}
	return code, resp
}

// prepareResponseWithPagination 准备响应信息
func prepareResponseWithPagination(c *gin.Context, version string,
	data interface{}, err error, pagination *Pagination) (int, *ResponseWithPagination) {
	code, resp := prepareResponse(c, version, data, err)
	respWithPagination := &ResponseWithPagination{
		*resp,
		pagination,
	}

	return code, respWithPagination
}

// handleData 格式化返回数据，非数组及切片时，转为切片
func handleData(data interface{}) interface{} {
	v := reflect.ValueOf(data)
	if !v.IsValid() || v.Kind() == reflect.Ptr && v.IsNil() {
		return make([]interface{}, 0)
	}
	if v.Kind() == reflect.Slice || v.Kind() == reflect.Array {
		return data
	}
	return []interface{}{data}
}
