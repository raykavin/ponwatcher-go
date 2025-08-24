package module

import (
	"context"
	"pon_watcher/internal/config"
	"pon_watcher/internal/usecase"
	"pon_watcher/pkg/log"

	"go.uber.org/fx"
)

type UseCaseParams struct {
	fx.In
	AppConfig      config.ApplicationProvider
	ServicesConfig config.ServicesProvider
	ONUManager     usecase.OpticalNetworkUnitManager
	OLTRepository  usecase.OLTRepository
	Log            log.Smart
}

type UseCaseContainer struct {
	fx.Out
	FetchAllONU            usecase.FetchAllONU
	FetchAllONUInformation usecase.FetchAllONUInformation
	FetchOLTIPByID         usecase.FetchOLTIPByID
	ONUCategorizer         usecase.ONUCategorizer
}

func NewUseCases(params UseCaseParams) (UseCaseContainer, error) {
	return UseCaseContainer{
		FetchAllONU:            usecase.NewFetchAllONU(params.ServicesConfig, params.ONUManager, params.Log),
		FetchAllONUInformation: usecase.NewFetchAllONUInformation(params.ServicesConfig, params.ONUManager, params.Log),
		FetchOLTIPByID:         usecase.NewFetchOLTIPByID(params.OLTRepository),
		ONUCategorizer:         usecase.NewONUCategorizer(params.Log),
	}, nil
}

func ucInvoker(
	lc fx.Lifecycle,
	fetchAllONU usecase.FetchAllONU,
	fetchAllONUInformation usecase.FetchAllONUInformation,
	log log.Smart,
) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			log.Info("Use cases initialized")
			return nil
		},
		OnStop: func(ctx context.Context) error {
			log.Info("Use cases shutting down")
			return nil
		},
	})
}

func UseCase() fx.Option {
	return fx.Module("usecases",
		fx.Provide(NewUseCases),
		fx.Invoke(ucInvoker),
	)
}
