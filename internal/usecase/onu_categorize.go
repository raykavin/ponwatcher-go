package usecase

import (
	"context"
	"pon_watcher/internal/dto"
	"pon_watcher/pkg/log"
	"strings"
	"time"
)

// ONUCategorizer defines the interface for ONU categorization
type ONUCategorizer interface {
	Execute(ctx context.Context, onus []*dto.OpticalNetworkUnitInfo) (*dto.CategorizedOpticalNetworkUnits, error)
}

type onuCategorizer struct {
	log log.Smart
}

// NewONUCategorizer creates a new ONU categorizer use case
func NewONUCategorizer(log log.Smart) ONUCategorizer {
	return &onuCategorizer{
		log: log,
	}
}

// Execute separates ONUs into online and offline categories based on their status
func (uc *onuCategorizer) Execute(ctx context.Context, onus []*dto.OpticalNetworkUnitInfo) (*dto.CategorizedOpticalNetworkUnits, error) {
	ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	uc.log.Debug("Starting ONU categorization process...")

	if onus == nil {
		uc.log.Warn("Received nil ONU slice")
		return uc.createEmptyResult(), nil
	}

	result := uc.createEmptyResult()

	start := time.Now()
	defer uc.benchmark(start)

	for _, onu := range onus {
		// Check if context is cancelled or deadline exceeded
		select {
		case <-ctx.Done():
			uc.log.Warn("ONU categorization cancelled due to context", "error", ctx.Err())
			return nil, ctx.Err()
		default:
		}

		if onu == nil {
			uc.log.Warn("Skipping nil ONU in slice")
			continue
		}

		result.Total++

		if uc.isOffline(onu) {
			uc.addToCategory(&result.Offline, onu)
		} else {
			uc.addToCategory(&result.Online, onu)
		}
	}

	uc.log.WithFields(map[string]any{
		"total":   result.Total,
		"online":  result.Online.Total,
		"offline": result.Offline.Total,
	}).Debug("ONU categorization completed")

	return result, nil
}

// createEmptyResult initializes an empty categorization result
func (uc *onuCategorizer) createEmptyResult() *dto.CategorizedOpticalNetworkUnits {
	return &dto.CategorizedOpticalNetworkUnits{
		Online:  dto.OpticalNetworkUnitCategory{ONUs: make([]*dto.OpticalNetworkUnitInfo, 0), Total: 0},
		Offline: dto.OpticalNetworkUnitCategory{ONUs: make([]*dto.OpticalNetworkUnitInfo, 0), Total: 0},
		Total:   0,
	}
}

// addToCategory adds an ONU to the specified category and updates the count
func (uc *onuCategorizer) addToCategory(category *dto.OpticalNetworkUnitCategory, onu *dto.OpticalNetworkUnitInfo) {
	category.ONUs = append(category.ONUs, onu)
	category.Total++
}

// isOffline determines if an ONU is offline based on voltage status and RX power
func (uc *onuCategorizer) isOffline(onu *dto.OpticalNetworkUnitInfo) bool {
	if onu == nil {
		return true
	}

	hasLowVoltage := strings.ToLower(strings.TrimSpace(onu.VoltageStatus)) == "low"
	hasZeroRxPower := strings.TrimSpace(onu.RxPower) == "0,00"

	return hasLowVoltage && hasZeroRxPower
}

func (uc *onuCategorizer) benchmark(start time.Time) {
	end := time.Now()
	dur := end.Sub(start)
	uc.log.Benchmark("ONU's categorization", dur)
}
