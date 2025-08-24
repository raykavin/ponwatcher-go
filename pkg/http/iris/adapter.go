package iris

import (
	"context"
	"net/http"

	"github.com/kataras/iris/v12"
)

// IrisAdapter implements the RequestContext interface using Iris
type IrisAdapter struct {
	Ctx iris.Context
}

// Abort implements http_engine.RequestContext.
func (r *IrisAdapter) Abort() {
	r.Ctx.StopExecution()
}

// BindJSON implements http_engine.RequestContext.
func (r *IrisAdapter) BindJSON(data any) {
	if err := r.Ctx.ReadJSON(data); err != nil {
		r.Ctx.StatusCode(http.StatusBadRequest)
		r.Ctx.JSON(map[string]string{"error": err.Error()})
	}
}

// Context implements http_engine.RequestContext.
func (r *IrisAdapter) Context() context.Context {
	return r.Request().Context()
}

// Get implements http_engine.RequestContext.
func (r *IrisAdapter) Get(key string) (any, bool) {
	value := r.Ctx.Values().Get(key)
	return value, value != nil
}

// GetParam implements http_engine.RequestContext.
func (r *IrisAdapter) GetParam(key string) string {
	return r.Ctx.Params().Get(key)
}

// GetQuery implements http_engine.RequestContext.
func (r *IrisAdapter) GetQuery(key string) string {
	return r.Ctx.URLParam(key)
}

// JSON implements http_engine.RequestContext.
func (r *IrisAdapter) JSON(statusCode int, data any) {
	r.Ctx.StatusCode(statusCode)
	r.Ctx.JSON(data)
}

// Next implements http_engine.RequestContext.
func (r *IrisAdapter) Next() {
	r.Ctx.Next()
}

// Redirect implements http_engine.RequestContext.
func (r *IrisAdapter) Redirect(statusCode int, to string) {
	r.Ctx.Redirect(to, statusCode)
}

// Request implements http_engine.RequestContext.
func (r *IrisAdapter) Request() *http.Request {
	return r.Ctx.Request()
}

// Set implements http_engine.RequestContext.
func (r *IrisAdapter) Set(key string, value any) {
	r.Ctx.Values().Set(key, value)
}

// SetCookie implements http_engine.RequestContext.
func (r *IrisAdapter) SetCookie(c *http.Cookie) {
	r.Ctx.SetCookie(c)
}

// Writer implements http_engine.RequestContext.
func (r *IrisAdapter) Writer() http.ResponseWriter {
	return r.Ctx.ResponseWriter()
}
