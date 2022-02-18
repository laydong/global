package httpx

import (
	"github.com/climber-dong/global/core/alarmx"
	"github.com/climber-dong/global/core/logx"
	"github.com/climber-dong/global/core/tracex"
	"github.com/climber-dong/global/gtools"
	"github.com/gin-gonic/gin"
	uuid "github.com/satori/go.uuid"
)

// WebHandlerFunc http请求的处理者
type WebHandlerFunc func(*WebContext)

// WebContext http 的context
// WebContext 继承了 gin.Context, 并且扩展了日志功能
type WebContext struct {
	*gin.Context
	*logx.LogContext
	*tracex.TraceContext
	*alarmx.AlarmContext
}

const ginFlag = "__gin__gin"

// NewWebContext 创建 http contextx
func NewWebContext(ginContext *gin.Context) *WebContext {
	obj, existed := ginContext.Get(ginFlag)
	if existed {
		return obj.(*WebContext)
	}

	logId := ginContext.GetHeader(gtools.RequestIdKey)
	if logId == "" {
		logId = gtools.Md5(uuid.NewV4().String())
		ginContext.Request.Header.Set(gtools.RequestIdKey, logId)
		ginContext.Set(gtools.RequestIdKey, logId)
	}

	tmp := &WebContext{
		Context:      ginContext,
		LogContext:   logx.NewLogContext(logId),
		TraceContext: tracex.NewTraceContext(ginContext.Request.RequestURI, ginContext.Request.Header),
	}
	ginContext.Set(ginFlag, tmp)

	return tmp
}
