package helper

import (
	"context"

	"github.com/jasonlabz/potato/consts"
)

func contextValue(ctx context.Context, key string) any {
	return ctx.Value(key)
}

func GetClientIP(ctx context.Context) string {
	return contextValue(ctx, consts.ContextClientAddr).(string)
}

func GetUserID(ctx context.Context) string {
	return contextValue(ctx, consts.ContextUserID).(string)
}

func GetToken(ctx context.Context) string {
	return contextValue(ctx, consts.ContextToken).(string)
}
