package usecase

import (
	"context"
	"pon_watcher/internal/dto"
	"time"
)

const (
	maxWorkers     = 20
	defaultTimeout = 5 * time.Minute // It's a default timeout of use case operation
)

// OpticalNetworkUnitManager handles ONU data operations
type OpticalNetworkUnitManager interface {
	FetchAllONU(ctx context.Context, params OnuParam) ([]*dto.OpticalNetworkUnit, error)
	FetchAllONUInformation(ctx context.Context, params OnuInfoParam) (*dto.OpticalNetworkUnitInfo, error)
}

type OnuParam struct {
	PONSlot   uint
	PONNumber uint
	OLT       string
	Filter    string
}

type OnuInfoParam struct {
	PONSlot      uint
	PONNumber    uint
	OLT          string
	PhysicalAddr string
	DeviceName   string
}
