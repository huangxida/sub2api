package admin

import (
	"context"
	"log"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/usagestats"
	"github.com/Wei-Shaw/sub2api/internal/service"
)

func (h *OpsHandler) attachInputCacheMetrics(ctx context.Context, filter *service.OpsDashboardFilter, overview *service.OpsDashboardOverview) {
	if h == nil || h.opsService == nil || overview == nil || filter == nil {
		return
	}

	metrics, err := h.buildInputCacheMetrics(ctx, filter)
	if err != nil {
		log.Printf("[OpsDashboard] build input cache metrics failed: %v", err)
		return
	}
	overview.InputCacheMetrics = metrics
}

func (h *OpsHandler) buildInputCacheMetrics(ctx context.Context, filter *service.OpsDashboardFilter) (*service.InputCacheMetrics, error) {
	if h == nil || h.opsService == nil || filter == nil {
		return nil, nil
	}

	now := time.Now().UTC()
	filtered := service.HasOpsInputCacheScopedFilters(filter)
	availableFrom := h.opsService.InputCacheAvailableFrom(now)

	windowSummary := &usagestats.InputCacheSummary{}
	windowPartial := false
	var windowAvailableFrom *time.Time

	windowFilter := *filter
	if filtered && windowFilter.StartTime.UTC().Before(availableFrom) {
		windowFilter.StartTime = availableFrom
		windowPartial = true
		windowAvailableFrom = &availableFrom
	}
	if windowFilter.StartTime.Before(windowFilter.EndTime) {
		summary, err := h.opsService.GetInputCacheSummary(ctx, &windowFilter)
		if err != nil {
			return nil, err
		}
		if summary != nil {
			windowSummary = summary
		}
	}

	cumulativeSummary := &usagestats.InputCacheSummary{}
	cumulativePartial := false
	var cumulativeAvailableFrom *time.Time

	if !filtered && h.dashboardService != nil {
		stats, err := h.dashboardService.GetDashboardStats(ctx)
		if err != nil {
			return nil, err
		}
		cumulativeSummary = service.InputCacheSummaryFromDashboardStats(stats)
	} else {
		cumulativeFilter := *filter
		cumulativeFilter.StartTime = availableFrom
		cumulativeFilter.EndTime = now
		cumulativePartial = true
		cumulativeAvailableFrom = &availableFrom

		summary, err := h.opsService.GetInputCacheSummary(ctx, &cumulativeFilter)
		if err != nil {
			return nil, err
		}
		if summary != nil {
			cumulativeSummary = summary
		}
	}

	return &service.InputCacheMetrics{
		Cumulative: service.BuildInputCacheMetricsSnapshot(cumulativeSummary, cumulativePartial, cumulativeAvailableFrom),
		Window:     service.BuildInputCacheMetricsWindow(windowSummary, filter.StartTime, filter.EndTime, windowPartial, windowAvailableFrom),
	}, nil
}
