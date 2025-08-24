package module

import (
	"context"
	"pon_watcher/internal/config"
	"pon_watcher/pkg/log"
	"pon_watcher/pkg/log/smart"
	"pon_watcher/pkg/viper"
	"sync"

	"go.uber.org/fx"
)

type ConfigParams struct {
	fx.In
	ConfigFile  string `name:"config_file"`
	WatchConfig bool   `name:"watch_config"`
}

type ConfigContainer struct {
	fx.Out
	App      config.ApplicationProvider
	Services config.ServicesProvider
}

type LoggerParams struct {
	fx.In
	Config config.ApplicationProvider
}

type ContextResult struct {
	fx.Out
	Context context.Context
	Cancel  context.CancelFunc
}

func NewContext() ContextResult {
	ctx, cancel := context.WithCancel(context.Background())
	return ContextResult{
		Context: ctx,
		Cancel:  cancel,
	}
}

func NewWaitGroup() *sync.WaitGroup {
	return &sync.WaitGroup{}
}

func NewConfig(params ConfigParams) (ConfigContainer, error) {
	loader := viper.New[config.Config](nil)
	cfg, err := loader.Load()
	if err != nil {
		return ConfigContainer{}, err
	}

	return ConfigContainer{
		App:      cfg.GetApplication(),
		Services: cfg.GetServices(),
	}, nil
}

func NewLogger(params LoggerParams) (log.Smart, error) {
	return smart.NewSmartZerologContextFromConfig(
		params.Config.GetLoggerLevel(),
		"2006-01-02 15:04:05",
		true,
		false,
	)
}

func crInvoker(lc fx.Lifecycle, cancel context.CancelFunc) {
	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			cancel()
			return nil
		},
	})
}

func Core(configFile string, watchConfig bool) fx.Option {
	return fx.Module("core",
		fx.Supply(
			fx.Annotate(configFile, fx.ResultTags(`name:"config_file"`)),
			fx.Annotate(watchConfig, fx.ResultTags(`name:"watch_config"`)),
		),
		fx.Provide(
			NewContext,
			NewWaitGroup,
			NewConfig,
			NewLogger,
		),
		fx.Invoke(crInvoker),
	)
}
