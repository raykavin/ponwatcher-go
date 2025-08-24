package gin

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
)

// GinAdapter implements the RequestContext interface using Gin
type GinAdapter struct {
	Ctx *gin.Context
}

// Abort implements http_engine.RequestContext.
func (g *GinAdapter) Abort() {
	g.Ctx.Abort()
}

// BindJSON implements http_engine.RequestContext.
func (g *GinAdapter) BindJSON(data any) {
	if err := g.Ctx.ShouldBindJSON(data); err != nil {
		g.Ctx.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
		g.Ctx.Abort()
	}
}

// Context implements http_engine.RequestContext.
func (g *GinAdapter) Context() context.Context {
	return g.Ctx.Request.Context()
}

// Get implements http_engine.RequestContext.
func (g *GinAdapter) Get(key string) (any, bool) {
	return g.Ctx.Get(key)
}

// GetParam implements http_engine.RequestContext.
func (g *GinAdapter) GetParam(key string) string {
	return g.Ctx.Param(key)
}

// GetQuery implements http_engine.RequestContext.
func (g *GinAdapter) GetQuery(key string) string {
	return g.Ctx.Query(key)
}

// JSON implements http_engine.RequestContext.
func (g *GinAdapter) JSON(statusCode int, data any) {
	g.Ctx.JSON(statusCode, data)
}

// Next implements http_engine.RequestContext.
func (g *GinAdapter) Next() {
	g.Ctx.Next()
}

// Redirect implements http_engine.RequestContext.
func (g *GinAdapter) Redirect(statusCode int, to string) {
	g.Ctx.Redirect(statusCode, to)
}

// Request implements http_engine.RequestContext.
func (g *GinAdapter) Request() *http.Request {
	return g.Ctx.Request
}

// Set implements http_engine.RequestContext.
func (g *GinAdapter) Set(key string, value any) {
	g.Ctx.Set(key, value)
}

// SetCookie implements http_engine.RequestContext.
func (g *GinAdapter) SetCookie(c *http.Cookie) {
	g.Ctx.SetCookie(c.Name, c.Value, c.MaxAge, c.Path, c.Domain, c.Secure, c.HttpOnly)
}

// Writer implements http_engine.RequestContext.
func (g *GinAdapter) Writer() http.ResponseWriter {
	return g.Ctx.Writer
}
