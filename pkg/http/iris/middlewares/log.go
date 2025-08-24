package middlewares

import (
	"pon_watcher/pkg/log"
	"time"

	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/context"
)

func RequestLog(log log.Smart) context.Handler {
	return func(ctx iris.Context) {
		start := time.Now()
		ctx.Next()

		duration := time.Since(start)

		log.API(ctx.Method(), ctx.Path(), ctx.GetStatusCode(), duration)
	}
}
