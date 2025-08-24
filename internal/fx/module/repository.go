package module

import (
	"context"
	"pon_watcher/internal/repository"
	"pon_watcher/internal/usecase"
	"pon_watcher/pkg/log"

	"go.uber.org/fx"
)

type RepositoryParams struct {
	Log log.Smart
}

type RepositoryContainer struct {
	fx.Out
	OLTRepository usecase.OLTRepository
}

func NewRepositories() (RepositoryContainer, error) {
	return RepositoryContainer{
		OLTRepository: repository.NewInMemoryOLTRepository(),
	}, nil
}

func rpInvoker(
	lc fx.Lifecycle,
	oltRepository usecase.OLTRepository,
	log log.Smart,
) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			log.Info("Repositories initialized")
			return nil
		},
		OnStop: func(ctx context.Context) error {
			log.Info("Repositories shutting down")
			return nil
		},
	})
}

func Repository() fx.Option {
	return fx.Module("repository",
		fx.Provide(NewRepositories),
		fx.Invoke(rpInvoker),
	)
}
