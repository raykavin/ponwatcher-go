package usecase

import (
	"context"
	"time"

	"pon_watcher/internal/config"
	"pon_watcher/internal/dto"
	"pon_watcher/pkg/log"
)

// FetchAllONU defines the interface for PON watcher operations
type FetchAllONU interface {
	Execute(ctx context.Context, params OnuParam) ([]*dto.OpticalNetworkUnit, error)
}

// fetchAllONU implements FetchAllONU
type fetchAllONU struct {
	config     config.ServicesProvider
	onuManager OpticalNetworkUnitManager
	log        log.Smart
}

// NewFetchAllONU creates a new PON watcher use case
func NewFetchAllONU(config config.ServicesProvider, onuManager OpticalNetworkUnitManager, log log.Smart) FetchAllONU {
	return &fetchAllONU{
		config:     config,
		onuManager: onuManager,
		log:        log,
	}
}

// Execute implements FetchAllONU
func (uc *fetchAllONU) Execute(ctx context.Context, params OnuParam) ([]*dto.OpticalNetworkUnit, error) {
	ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	uc.logSearchOperation(params.Filter)

	start := time.Now()
	defer uc.benchmark(start)

	onus, err := uc.onuManager.FetchAllONU(ctx, params)
	if err != nil {
		uc.log.WithError(err).Error("Failed to fetch optical network units")
		return nil, err
	}

	return onus, nil
}

// logSearchOperation logs the search operation with appropriate details
func (uc *fetchAllONU) logSearchOperation(filter string) {
	msg := "Searching for ONUs..."

	if filter != "" {
		uc.log.WithField("filter", filter).Debug(msg)
	} else {
		uc.log.Debug(msg)
	}
}

func (uc *fetchAllONU) benchmark(start time.Time) {
	end := time.Now()
	dur := end.Sub(start)
	uc.log.Benchmark("Fetch all ONU's", dur)
}
