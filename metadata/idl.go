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
//  Created:   2025/12/1 23:57
//  Version:   v1.0.0

package metadata

var IDL_CLIENT = `namespace go api

struct Request {
1: string message
}

struct Response {
1: string message
}

service Echo {
Response echo(1: Request req)
}
`

var IDL_SERVER = `namespace go api

struct Request {
1: string message
}

struct Response {
1: string message
}

service Echo {
Response echo(1: Request req)
}
`
