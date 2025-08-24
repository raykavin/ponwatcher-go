package handler

import (
	"context"
	"errors"
	"strconv"

	"pon_watcher/internal/dto"
	"pon_watcher/internal/http/response"
	"pon_watcher/internal/usecase"
	"pon_watcher/pkg/http"
	"pon_watcher/pkg/log"
)

// PONWatcher handles PON-related HTTP requests
type PONWatcher interface {
	Execute(ctx http.RequestContext)
}

// ponWatcherHandler implements PONWatcher interface
type ponWatcherHandler struct {
	fetchAllONU     usecase.FetchAllONU
	fetchAllONUInfo usecase.FetchAllONUInformation
	fetchOLTIPByID  usecase.FetchOLTIPByID
	onuCategorizer  usecase.ONUCategorizer
	logger          log.Smart
}

// NewPONWatcher creates a new PON watcher handler
func NewPONWatcher(
	fetchAllONU usecase.FetchAllONU,
	fetchAllONUInfo usecase.FetchAllONUInformation,
	fetchOLTIPByID usecase.FetchOLTIPByID,
	onuCategorizer usecase.ONUCategorizer,
	logger log.Smart,
) PONWatcher {
	return &ponWatcherHandler{
		fetchAllONU:     fetchAllONU,
		fetchAllONUInfo: fetchAllONUInfo,
		fetchOLTIPByID:  fetchOLTIPByID,
		onuCategorizer:  onuCategorizer,
		logger:          logger,
	}
}

// queryParams holds the parsed query parameters
type queryParams struct {
	slot   uint
	card   uint
	oltID  int
	filter string
}

// Execute implements PONWatcher interface
func (h *ponWatcherHandler) Execute(ctx http.RequestContext) {
	params, err := h.parseQueryParams(ctx)
	if err != nil {
		h.logger.Error("Failed to parse query parameters", "error", err)
		response.NewBadRequest().Write(ctx.Writer())
		ctx.Abort()
		return
	}

	oltIP, exists := h.fetchOLTIPByID.Execute(ctx.Context(), params.oltID)
	if !exists {
		h.logger.WithError(err).
			WithField("olt_id", params.oltID).
			Error("Failed to get OLT IP")

		response.NewUnprocessableEntity().Write(ctx.Writer())
		ctx.Abort()
		return
	}

	onusInfo, err := h.fetchONUInformation(ctx.Context(), params, oltIP)
	if err != nil {
		h.logger.WithError(err).Error("Failed to fetch ONU information")
		response.NewError(500, err).Write(ctx.Writer())
		ctx.Abort()
		return
	}

	categorizedONUs, err := h.onuCategorizer.Execute(ctx.Context(), onusInfo)
	if err != nil {
		h.logger.WithError(err).Error("Failed to categorize ONU's")
		response.NewError(500, err).Write(ctx.Writer())
		ctx.Abort()
		return
	}

	response.NewOK(categorizedONUs).Write(ctx.Writer())
}

// parseQueryParams extracts and validates query parameters
func (h *ponWatcherHandler) parseQueryParams(ctx http.RequestContext) (*queryParams, error) {
	slot, err := h.parsePositiveInt(ctx.GetQuery("slot"))
	if err != nil {
		return nil, errors.New("must be a positive integer")
	}

	card, err := h.parsePositiveInt(ctx.GetQuery("card"))
	if err != nil {
		return nil, errors.New("must be a positive integer")
	}

	oltID, err := h.parseInt(ctx.GetQuery("olt"))
	if err != nil {
		return nil, errors.New("must be a valid integer")
	}

	return &queryParams{
		slot:   uint(slot),
		card:   uint(card),
		oltID:  oltID,
		filter: ctx.GetQuery("s"),
	}, nil
}

// parsePositiveInt parses a string to a positive integer
func (h *ponWatcherHandler) parsePositiveInt(s string) (int, error) {
	val, err := strconv.Atoi(s)
	if err != nil || val <= 0 {
		return 0, errors.New("invalid positive integer")
	}
	return val, nil
}

// parseInt parses a string to an integer
func (h *ponWatcherHandler) parseInt(s string) (int, error) {
	val, err := strconv.Atoi(s)
	if err != nil {
		return 0, errors.New("invalid integer")
	}
	return val, nil
}

// fetchONUInformation fetches ONU and ONU information
func (h *ponWatcherHandler) fetchONUInformation(ctx context.Context, params *queryParams, oltIP string) ([]*dto.OpticalNetworkUnitInfo, error) {
	onus, err := h.fetchAllONU.Execute(ctx, usecase.OnuParam{
		PONSlot:   params.slot,
		PONNumber: params.card,
		OLT:       oltIP,
		Filter:    params.filter,
	})
	if err != nil {
		return nil, err
	}

	onusInfo, err := h.fetchAllONUInfo.Execute(ctx, usecase.OnuInfoParam{
		PONSlot:   params.slot,
		PONNumber: params.card,
		OLT:       oltIP,
	}, onus)
	if err != nil {
		return nil, err
	}

	return onusInfo, nil
}
