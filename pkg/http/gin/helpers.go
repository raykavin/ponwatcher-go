package gin

import (
	"pon_watcher/pkg/http"

	"github.com/gin-gonic/gin"
)

// RequestContext interface - you'll need to import this from your http package
// This is just for reference and should be imported from your actual http package
/*
type RequestContext interface {
	Context() context.Context
	Writer() http.ResponseWriter
	Request() *http.Request
	Set(key string, value any)
	Get(key string) (any, bool)
	JSON(statusCode int, data any)
	BindJSON(outPtr any)
	GetParam(key string) string
	GetQuery(key string) string
	Redirect(statusCode int, to string)
	SetCookie(c *http.Cookie)
	Abort()
	Next()
}
*/

// WrapHandler converts a RequestContext-based handler to a Gin handler
func WrapHandler(handler func(ctx http.RequestContext)) gin.HandlerFunc {
	return func(c *gin.Context) {
		adapter := &GinAdapter{Ctx: c}
		handler(adapter)
	}
}

// WrapMiddleware converts a RequestContext-based middleware to a Gin middleware
func WrapMiddleware(middleware func(ctx http.RequestContext)) gin.HandlerFunc {
	return func(c *gin.Context) {
		adapter := &GinAdapter{Ctx: c}
		middleware(adapter)
	}
}

// GetAdapter extracts the GinAdapter from a Gin context
// Useful if you need to access Gin-specific features
func GetAdapter(c *gin.Context) *GinAdapter {
	return &GinAdapter{Ctx: c}
}

// SetupRoutes is a helper function to demonstrate route setup
// You can customize this based on your needs
func (e *Engine) SetupRoutes(routes func(router *gin.Engine)) {
	routes(e.router)
}
