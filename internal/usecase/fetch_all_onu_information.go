package usecase

import (
	"context"
	"errors"
	"fmt"
	"pon_watcher/internal/config"
	"pon_watcher/internal/dto"
	"pon_watcher/pkg/log"
	"sync"
	"time"
)

// Job is a single ONU processing job
type Job struct {
	ONU    *dto.OpticalNetworkUnit
	Params OnuInfoParam
}

// Result is the result of processing a single ONU
type Result struct {
	Info *dto.OpticalNetworkUnitInfo
	Err  error
}

// ProcessingStats holds statistics about the processing operation
type ProcessingStats struct {
	Processed       int
	Successful      int
	Failed          int
	ContextCanceled int
}

// FetchAllONUInformation defines the interface for fetching ONU information
type FetchAllONUInformation interface {
	Execute(ctx context.Context, params OnuInfoParam, onus []*dto.OpticalNetworkUnit) ([]*dto.OpticalNetworkUnitInfo, error)
}

// fetchAllONUInformation implements the FetchAllONUInformation use case
type fetchAllONUInformation struct {
	config     config.ServicesProvider
	onuManager OpticalNetworkUnitManager
	log        log.Smart
}

// NewFetchAllONUInformation creates a new fetch all ONU information use case
func NewFetchAllONUInformation(
	config config.ServicesProvider,
	onuManager OpticalNetworkUnitManager,
	log log.Smart,
) FetchAllONUInformation {
	return &fetchAllONUInformation{
		config:     config,
		onuManager: onuManager,
		log:        log,
	}
}

// Execute fetches information for all provided ONUs using concurrent workers
func (uc *fetchAllONUInformation) Execute(
	ctx context.Context,
	params OnuInfoParam,
	onus []*dto.OpticalNetworkUnit,
) ([]*dto.OpticalNetworkUnitInfo, error) {
	ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	if err := uc.validateParams(params, onus); err != nil {
		return nil, err
	}

	if len(onus) == 0 {
		uc.log.Debug("No ONUs provided for information fetching")
		return []*dto.OpticalNetworkUnitInfo{}, nil
	}

	defer uc.trackExecutionTime(time.Now())

	return uc.processONUsConcurrently(ctx, params, onus)
}

// validateParams validates input parameters and ONUs
func (uc *fetchAllONUInformation) validateParams(params OnuInfoParam, onus []*dto.OpticalNetworkUnit) error {
	if err := uc.validateBaseParams(params); err != nil {
		return fmt.Errorf("invalid parameters: %w", err)
	}

	for _, onu := range onus {
		if err := uc.validateONU(onu); err != nil {
			return fmt.Errorf("invalid ONU %s: %w", uc.getONUIdentifier(onu), err)
		}
	}

	return nil
}

// processONUsConcurrently processes ONUs using worker goroutines
func (uc *fetchAllONUInformation) processONUsConcurrently(
	ctx context.Context,
	params OnuInfoParam,
	onus []*dto.OpticalNetworkUnit,
) ([]*dto.OpticalNetworkUnitInfo, error) {
	if err := uc.checkContext(ctx, "before concurrent processing"); err != nil {
		return nil, err
	}

	numWorkers := min(len(onus), maxWorkers)
	uc.logProcessingStart(numWorkers, len(onus))

	jobs := make(chan Job, len(onus))
	results := make(chan Result, len(onus))

	// Start workers
	var wg sync.WaitGroup
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go uc.worker(ctx, i, jobs, results, &wg)
	}

	// Send jobs
	go uc.sendJobs(ctx, jobs, onus, params)

	// Wait for completion
	go func() {
		wg.Wait()
		close(results)
	}()

	return uc.collectResults(ctx, results, len(onus))
}

// sendJobs sends all jobs to the jobs channel
func (uc *fetchAllONUInformation) sendJobs(ctx context.Context, jobs chan<- Job, onus []*dto.OpticalNetworkUnit, params OnuInfoParam) {
	defer close(jobs)

	for _, onu := range onus {
		select {
		case jobs <- Job{ONU: onu, Params: params}:
		case <-ctx.Done():
			uc.log.Warn("Context cancelled while sending jobs")
			return
		}
	}
}

// worker processes jobs from the jobs channel
func (uc *fetchAllONUInformation) worker(
	ctx context.Context,
	workerID int,
	jobs <-chan Job,
	results chan<- Result,
	wg *sync.WaitGroup,
) {
	defer wg.Done()

	logger := uc.log.WithField("worker_id", workerID)
	logger.Debug("Worker started")

	for {
		select {
		case job, ok := <-jobs:
			if !ok {
				logger.Debug("Worker finished - no more jobs")
				return
			}

			if err := uc.checkContext(ctx, "during job processing"); err != nil {
				uc.sendResult(ctx, results, Result{Err: err})
				return
			}

			result := uc.processJob(ctx, job, workerID)
			if !uc.sendResult(ctx, results, result) {
				return
			}

		case <-ctx.Done():
			logger.WithError(ctx.Err()).Warn("Worker stopped - context cancelled")
			return
		}
	}
}

// processJob processes a single ONU job
func (uc *fetchAllONUInformation) processJob(ctx context.Context, job Job, workerID int) Result {
	logger := uc.log.WithField("worker_id", workerID).WithField("physical_address", job.ONU.Mac)

	if err := uc.checkContext(ctx, "before ONU manager call"); err != nil {
		return Result{Err: err}
	}

	logger.Debug("Fetching ONU information...")

	// Prepare parameters
	onuParams := job.Params
	onuParams.PhysicalAddr = job.ONU.Mac
	onuParams.DeviceName = job.ONU.Name

	info, err := uc.onuManager.FetchAllONUInformation(ctx, onuParams)
	if err != nil {
		uc.logProcessingError(err, job.ONU.Mac, workerID)
		return Result{Err: err}
	}

	logger.Debug("Successfully fetched ONU information")
	return Result{Info: info}
}

// sendResult safely sends a result to the results channel
func (uc *fetchAllONUInformation) sendResult(ctx context.Context, results chan<- Result, result Result) bool {
	select {
	case results <- result:
		return true
	case <-ctx.Done():
		uc.log.WithError(ctx.Err()).Warn("Context cancelled while sending result")
		return false
	}
}

// collectResults collects all results from the results channel
func (uc *fetchAllONUInformation) collectResults(
	ctx context.Context,
	results <-chan Result,
	expectedCount int,
) ([]*dto.OpticalNetworkUnitInfo, error) {
	onusInfo := make([]*dto.OpticalNetworkUnitInfo, 0, expectedCount)
	stats := ProcessingStats{}

	for result := range results {
		if err := uc.checkContext(ctx, "during result collection"); err != nil {
			return nil, err
		}

		stats.Processed++
		uc.updateStats(&stats, result)

		if result.Err == nil && result.Info != nil {
			onusInfo = append(onusInfo, result.Info)
		}
	}

	uc.logProcessingCompletion(stats)

	if stats.ContextCanceled == stats.Processed && stats.Processed > 0 {
		return nil, ctx.Err()
	}

	return onusInfo, nil
}

// updateStats updates processing statistics based on result
func (uc *fetchAllONUInformation) updateStats(stats *ProcessingStats, result Result) {
	if result.Err != nil {
		if uc.isContextError(result.Err) {
			stats.ContextCanceled++
		} else {
			stats.Failed++
		}
	} else if result.Info != nil {
		stats.Successful++
	}
}

// checkContext checks if context is cancelled and returns appropriate error
func (uc *fetchAllONUInformation) checkContext(ctx context.Context, stage string) error {
	select {
	case <-ctx.Done():
		if stage != "" {
			uc.log.WithError(ctx.Err()).Warnf("Operation cancelled %s", stage)
		} else {
			uc.log.WithError(ctx.Err()).Warn("Operation cancelled")
		}
		return ctx.Err()
	default:
		return nil
	}
}

// isContextError checks if error is due to context cancellation
func (uc *fetchAllONUInformation) isContextError(err error) bool {
	return errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded)
}

// validateBaseParams validates the base parameters required for ONU information fetching
func (uc *fetchAllONUInformation) validateBaseParams(params OnuInfoParam) error {
	var validationErrors []string

	if len(params.OLT) == 0 {
		validationErrors = append(validationErrors, "OLT address is required")
	}

	if params.PONNumber == 0 {
		validationErrors = append(validationErrors, "PON card number must be greater than 0")
	}

	if params.PONSlot == 0 {
		validationErrors = append(validationErrors, "PON slot must be greater than 0")
	}

	if len(validationErrors) > 0 {
		return fmt.Errorf("validation failed: %v", validationErrors)
	}

	return nil
}

// validateONU validates a single ONU
func (uc *fetchAllONUInformation) validateONU(onu *dto.OpticalNetworkUnit) error {
	if onu == nil {
		return errors.New("ONU cannot be nil")
	}

	if len(onu.Mac) == 0 {
		return errors.New("ONU physical address (MAC) is required")
	}

	return nil
}

// getONUIdentifier returns a string identifier for the ONU for logging purposes
func (uc *fetchAllONUInformation) getONUIdentifier(onu *dto.OpticalNetworkUnit) string {
	if onu == nil {
		return "nil"
	}
	if len(onu.Mac) > 0 {
		return onu.Mac
	}
	return "unknown"
}

// Logging helper methods
func (uc *fetchAllONUInformation) logProcessingStart(numWorkers, totalONUs int) {
	uc.log.WithField("workers", numWorkers).
		WithField("total_onus", totalONUs).
		Debug("Starting concurrent ONU information fetching")
}

func (uc *fetchAllONUInformation) logProcessingError(err error, mac string, workerID int) {
	logger := uc.log.WithError(err).
		WithField("physical_address", mac).
		WithField("worker_id", workerID)

	if uc.isContextError(err) {
		logger.Warn("ONU information fetch cancelled due to context")
	} else {
		logger.Error("Failed to fetch ONU information")
	}
}

func (uc *fetchAllONUInformation) logProcessingCompletion(stats ProcessingStats) {
	uc.log.WithField("processed", stats.Processed).
		WithField("successful", stats.Successful).
		WithField("failed", stats.Failed).
		WithField("context_cancelled", stats.ContextCanceled).
		Info("Completed concurrent ONU information fetching")
}

func (uc *fetchAllONUInformation) trackExecutionTime(start time.Time) {
	duration := time.Since(start)
	uc.log.Benchmark("Fetch all ONU's information", duration)
}
