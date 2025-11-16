package api

import (
	"github.com/getsentry/sentry-go"
	"github.com/gin-gonic/gin"
)

func captureSentry(c *gin.Context, err error, msg string) {
	if err == nil {
		return
	}
	sentry.WithScope(func(scope *sentry.Scope) {
		scope.SetTag("path", c.FullPath())
		scope.SetTag("method", c.Request.Method)
		scope.SetExtra("params", c.Params)
		scope.SetExtra("query", c.Request.URL.Query())
		scope.SetExtra("msg", msg)
		sentry.CaptureException(err)
	})
}
