package contextx

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/yunhanshu-net/pkg/constants"
	"strconv"
)

func GetRequestUserName(ctx context.Context) string {
	v, ok := ctx.(*gin.Context)
	if ok {
		return v.GetString("user")
	}
	return ""
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
