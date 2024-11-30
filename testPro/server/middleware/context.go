package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/jasonlabz/potato/consts"
)

type Options struct {
	headerMap      map[string]string
	customFieldMap map[string]func(ctx *gin.Context) string
}

type Option func(options *Options)

func WithHeaderField(headerMap map[string]string) Option {
	return func(options *Options) {
		options.headerMap = headerMap
	}
}

func WithCustomField(customFieldMap map[string]func(ctx *gin.Context) string) Option {
	return func(options *Options) {
		options.customFieldMap = customFieldMap
	}
}

func SetContextMiddleware(opts ...Option) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var options = &Options{}
		for _, opt := range opts {
			opt(options)
		}

		for headerKey, contextKey := range options.headerMap {
			if headerKey == "" || contextKey == "" {
				continue
			}
			value := ctx.Request.Header.Get(headerKey)
			ctx.Set(contextKey, value)
		}

		for contextKey, handler := range options.customFieldMap {
			value := handler(ctx)
			ctx.Set(contextKey, value)
		}

		traceID := ctx.Request.Header.Get(consts.HeaderRequestID)
		if traceID == "" {
			traceID = strings.ReplaceAll(uuid.New().String(), consts.SignDash, consts.EmptyString)
		}
		userID := ctx.Request.Header.Get(consts.HeaderUserID)
		authorization := ctx.Request.Header.Get(consts.HeaderAuthorization)
		remote := ctx.ClientIP()

		ctx.Set(consts.ContextToken, authorization)
		ctx.Set(consts.ContextUserID, userID)
		ctx.Set(consts.ContextTraceID, traceID)
		ctx.Set(consts.ContextClientAddr, remote)

		ctx.Next()
	}
}
