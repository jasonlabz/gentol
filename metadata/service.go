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
//  Created:   2025/12/6 16:48
//  Version:   v1.0.0

package metadata

const AddService = `package {{.ServicePackageName}}

// {{.ServiceStructName}} 接口定义。
// 外部调用者只依赖此接口，不感知具体实现。

// {{.ServiceStructName}} ...
type {{.ServiceStructName}} interface {
	// TODO: add definition of method
}`

const AddServiceImpl = `package {{.ServiceName}}

import (
	"sync"

	"{{.ModulePath}}{{.ServiceDir}}"
)

var (
	svc  *{{.ServiceStructName}}Impl
	once sync.Once
)

// Get{{.ServiceStructName}} returns the singleton instance.
func Get{{.ServiceStructName}}() {{.ServicePackageName}}.{{.ServiceStructName}} {
	if svc != nil {
		return svc
	}
	once.Do(func() {
		svc = &{{.ServiceStructName}}Impl{}
	})
	return svc
}

// {{.ServiceStructName}}Impl implements {{.ServicePackageName}}.{{.ServiceStructName}}.
type {{.ServiceStructName}}Impl struct {
	// add properties, eg: userDao dao.UserDao
}
`

const AddDto = `package {{.ServiceName}}

// 本服务使用的请求/响应 DTO 定义。
// 入参校验、出参序列化相关的结构体放在此文件。
`

const AddHelper = `package {{.ServiceName}}

// 服务内部使用的工具/辅助函数。
// 与外部调用无关的纯逻辑辅助方法放在此文件。
`
