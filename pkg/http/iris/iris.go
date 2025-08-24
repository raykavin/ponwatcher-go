package iris

import (
	"context"

	"github.com/kataras/iris/v12"
)

type Engine struct {
	app  *iris.Application
	addr string
}

func New(listen string, loggerLevel string) (*Engine, error) {
	if len(listen) == 0 {
		return nil, ErrInvalidListenAddress
	}

	return &Engine{
		app:  iris.New(),
		addr: listen,
	}, nil
}

// Listen implements HttpServer.
func (s *Engine) Listen() error {
	return s.app.Listen(s.addr)
}

// Shutdown implements HttpServer.
func (s *Engine) Shutdown(ctx context.Context) error {
	return s.app.Shutdown(ctx)
}

// App provides iris engine application
func (s *Engine) App() *iris.Application {
	return s.app
}
