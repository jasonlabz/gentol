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

type {{.ServiceStructName}} interface {
	// TODO: add definition of method
}`

const AddServiceImpl = `package {{.ServiceName}}

import (
	"sync"

	"{{.ModulePath}}/server/{{.ServicePackageName}}"
)

var svc *service
var once sync.Once

func GetService() {{.ServicePackageName}}.{{.ServiceStructName}} {
	if svc != nil {
		return svc
	}
	once.Do(func() {
		// init service
		svc = &service{}
	})

	return svc
}

type service struct {
   // add properties, eg: userDao dao.UserDao
}
`

const EmptyMeta = `package body
`
