package module

import (
	"context"
	unmAdapter "pon_watcher/internal/adapter/unm"
	"pon_watcher/internal/config"
	"pon_watcher/internal/usecase"
	"pon_watcher/pkg/log"
	"pon_watcher/pkg/tl1"
	"pon_watcher/pkg/unm"

	"go.uber.org/fx"
)

type UNMParams struct {
	fx.In
	ServicesConfig config.ServicesProvider
	Log            log.Smart
}

type UNMContainer struct {
	fx.Out
	OpticalNetworkManager usecase.OpticalNetworkUnitManager
}

func NewUNM(params UNMParams) (UNMContainer, error) {
	params.Log.WithFields(map[string]any{
		"hostname": params.ServicesConfig.GetUNM().GetHost(),
		"port":     params.ServicesConfig.GetUNM().GetPort(),
	}).Info("Initializing UNM connection")

	transporter, err := tl1.NewTransport(
		params.ServicesConfig.GetUNM().GetHost(),
		params.ServicesConfig.GetUNM().GetPort(),
	)

	if err != nil {
		params.Log.WithError(err).Error("Failed to create TL1 transport")
		return UNMContainer{}, err
	}

	unmServer := unm.New(
		params.ServicesConfig.GetUNM().GetUsername(),
		params.ServicesConfig.GetUNM().GetPassword(),
		transporter,
		params.Log,
	)

	adapter := unmAdapter.NewONUManagerAdapter(unmServer)

	params.Log.Info("UNM connection initialized successfully")

	return UNMContainer{OpticalNetworkManager: adapter}, nil
}

func unmInvoker(lc fx.Lifecycle, manager usecase.OpticalNetworkUnitManager, logger log.Smart) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			logger.Info("UNM services started")
			return nil
		},
		OnStop: func(ctx context.Context) error {
			logger.Info("UNM services shutting down")
			return nil
		},
	})
}

func UNM() fx.Option {
	return fx.Module("unm",
		fx.Provide(NewUNM),
		fx.Invoke(unmInvoker),
	)
}
