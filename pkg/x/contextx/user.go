package contextx

import (
	"context"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/yunhanshu-net/pkg/constants"
)

func GetRequestUserName(ctx context.Context) string {
	// 首先尝试从gin.Context获取
	if v, ok := ctx.(*gin.Context); ok {
		return v.GetString("user")
	}

	// 然后尝试从context.Value获取
	if user, ok := ctx.Value("user").(string); ok && user != "" {
		return user
	}

	return "beiluo"
}

func WithRequestUserName(ctx context.Context, username string) context.Context {
	return context.WithValue(ctx, "user", username)
}
func GetTraceID(ctx context.Context) string {
	v, ok := ctx.(*gin.Context)
	if ok {
		return v.GetString(constants.TraceID)
	}
	return ""
}
func GetFunctionID(ctx context.Context) int {
	v, ok := ctx.(*gin.Context)
	if ok {
		s := v.GetString(constants.FunctionID)

		i, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return 0
		}
		return int(i)
	}
	return 0
}
