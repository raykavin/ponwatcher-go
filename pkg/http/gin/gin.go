package gin

import (
	"context"
	netHttp "net/http"
	"pon_watcher/pkg/http"

	"github.com/gin-gonic/gin"
)

// Engine wraps Gin engine and implements HttpServer interface
type Engine struct {
	router *gin.Engine
	server *netHttp.Server
	addr   string
}

// Ensure it implements http interfaces
var _ http.HttpServer = (*Engine)(nil)

// New creates a new Gin engine with the specified configuration
func New(listen string, loggerLevel string) (*Engine, error) {
	if len(listen) == 0 {
		return nil, ErrInvalidListenAddress
	}

	// Set Gin mode based on logger level
	switch loggerLevel {
	case "debug":
		gin.SetMode(gin.DebugMode)
	case "release", "production":
		gin.SetMode(gin.ReleaseMode)
	default:
		gin.SetMode(gin.TestMode)
	}

	router := gin.New()

	// Add default middleware (Recovery)
	router.Use(gin.Recovery())

	return &Engine{
		router: router,
		addr:   listen,
		server: &netHttp.Server{
			Addr:    listen,
			Handler: router,
		},
	}, nil
}

// Listen implements HttpServer interface
func (e *Engine) Listen() error {
	return e.server.ListenAndServe()
}

// Shutdown implements HttpServer interface
func (e *Engine) Shutdown(ctx context.Context) error {
	return e.server.Shutdown(ctx)
}

// Router provides access to the underlying Gin router
func (e *Engine) Router() *gin.Engine {
	return e.router
}

// Server provides access to the underlying HTTP server
func (e *Engine) Server() *netHttp.Server {
	return e.server
}
