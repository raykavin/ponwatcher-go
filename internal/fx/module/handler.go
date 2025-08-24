package module

import (
	"context"
	"fmt"
	"pon_watcher/internal/config"
	"pon_watcher/internal/http/handler"
	"pon_watcher/internal/usecase"
	"pon_watcher/pkg/http/gin"
	"pon_watcher/pkg/http/gin/middlewares"
	"pon_watcher/pkg/log"
	"sync"
	"time"

	"github.com/gin-contrib/cors"
	ginfw "github.com/gin-gonic/gin"
	"go.uber.org/fx"
)

type HTTPHandlerParams struct {
	fx.In
	WaitGroup              *sync.WaitGroup
	Context                context.Context
	AppConfig              config.ApplicationProvider
	ServicesConfig         config.ServicesProvider
	Log                    log.Smart
	FetchAllONU            usecase.FetchAllONU
	FetchAllONUInformation usecase.FetchAllONUInformation
	FetchOLTIPByID         usecase.FetchOLTIPByID
	ONUCategorizer         usecase.ONUCategorizer
}

type HTTPHandlerContainer struct {
	fx.Out
	PonWatcherHandler handler.PONWatcher
	Engine            *gin.Engine
	Router            *ginfw.Engine
}

func NewHTTPHandlers(params HTTPHandlerParams) (HTTPHandlerContainer, error) {
	// Create handlers
	ponWatcherHandler := handler.NewPONWatcher(
		params.FetchAllONU,
		params.FetchAllONUInformation,
		params.FetchOLTIPByID,
		params.ONUCategorizer,
		params.Log,
	)

	// Get port from config
	port := params.AppConfig.GetWeb().GetListen()
	if port == 0 {
		port = 3000
	}

	// Create Gin engine
	engine, err := gin.New(fmt.Sprintf(":%d", port), params.AppConfig.GetLoggerLevel())
	if err != nil {
		return HTTPHandlerContainer{}, err
	}

	router := engine.Router()

	// Setup middleware
	setupMiddleware(router, params)

	// Setup routes
	setupRoutes(router, ponWatcherHandler)

	return HTTPHandlerContainer{
		PonWatcherHandler: ponWatcherHandler,
		Engine:            engine,
		Router:            router,
	}, nil
}

func setupMiddleware(router *ginfw.Engine, params HTTPHandlerParams) {
	// CORS middleware
	corsSetting := params.AppConfig.GetWeb().GetCORS()
	corsConfig := cors.Config{
		AllowAllOrigins: true,
		AllowOrigins:    []string{"*"},
		AllowHeaders:    []string{"Content-Type", "Content-Length", "Accept", "Origin"},
		AllowMethods:    []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"},
	}

	if len(corsSetting.GetAllowMethods()) > 0 {
		corsConfig.AllowMethods = corsSetting.GetAllowMethods()
	}

	if len(corsSetting.GetAllowHeaders()) > 0 {
		corsConfig.AllowHeaders = corsSetting.GetAllowHeaders()
	}

	if len(corsSetting.GetAllowOrigins()) > 0 {
		corsConfig.AllowAllOrigins = false
		corsConfig.AllowOrigins = corsSetting.GetAllowOrigins()
	}

	router.Use(cors.New(corsConfig))

	// Request logging middleware
	router.Use(middlewares.RequestLog(params.Log))
}

func setupRoutes(router *ginfw.Engine, ponWatcherHandler handler.PONWatcher) {
	// API v1 routes
	v1 := router.Group("/api/v1")
	ponWatcher := v1.Group("/pon-watcher")
	ponWatcher.GET("", gin.WrapHandler(ponWatcherHandler.Execute))
}

func hdlInvoker(
	lc fx.Lifecycle,
	router *ginfw.Engine,
	engine *gin.Engine,
	logger log.Smart,
	wg *sync.WaitGroup,
) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			wg.Add(1)
			go func() {
				defer wg.Done()
				if err := engine.Listen(); err != nil {
					logger.WithError(err).Error("HTTP server error")
				}
			}()

			time.Sleep(100 * time.Millisecond)
			logger.Info("HTTP server started successfully")
			return nil
		},
		OnStop: func(ctx context.Context) error {
			logger.Info("Stopping HTTP server...")
			shutdownCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
			defer cancel()

			if err := engine.Shutdown(shutdownCtx); err != nil {
				logger.Error("HTTP server shutdown error", "error", err)
				return err
			}

			logger.Info("HTTP server stopped successfully")
			return nil
		},
	})
}

func HTTPHandler() fx.Option {
	return fx.Module("http_handlers",
		fx.Provide(NewHTTPHandlers),
		fx.Invoke(hdlInvoker),
	)
}
