package service

import (
	"context"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/usagestats"
)

type InputCacheMetrics struct {
	Cumulative *InputCacheMetricsSnapshot `json:"cumulative,omitempty"`
	Window     *InputCacheMetricsWindow   `json:"window,omitempty"`
}

type InputCacheMetricsSnapshot struct {
	InputTokens         int64      `json:"input_tokens"`
	OutputTokens        int64      `json:"output_tokens"`
	CacheReadTokens     int64      `json:"cache_read_tokens"`
	CacheCreationTokens int64      `json:"cache_creation_tokens"`
	CacheReadRatio      *float64   `json:"cache_read_ratio"`
	TotalCacheTokens    int64      `json:"total_cache_tokens"`
	TotalTokens         int64      `json:"total_tokens"`
	HasData             bool       `json:"has_data"`
	Partial             bool       `json:"partial"`
	AvailableFrom       *time.Time `json:"available_from,omitempty"`
}

type InputCacheMetricsWindow struct {
	InputTokens         int64      `json:"input_tokens"`
	OutputTokens        int64      `json:"output_tokens"`
	CacheReadTokens     int64      `json:"cache_read_tokens"`
	CacheCreationTokens int64      `json:"cache_creation_tokens"`
	CacheReadRatio      *float64   `json:"cache_read_ratio"`
	TotalCacheTokens    int64      `json:"total_cache_tokens"`
	TotalTokens         int64      `json:"total_tokens"`
	HasData             bool       `json:"has_data"`
	Partial             bool       `json:"partial"`
	AvailableFrom       *time.Time `json:"available_from,omitempty"`
	StartTime           time.Time  `json:"start_time"`
	EndTime             time.Time  `json:"end_time"`
}

func BuildInputCacheMetricsSnapshot(summary *usagestats.InputCacheSummary, partial bool, availableFrom *time.Time) *InputCacheMetricsSnapshot {
	if summary == nil {
		summary = &usagestats.InputCacheSummary{}
	}
	return &InputCacheMetricsSnapshot{
		InputTokens:         summary.InputTokens,
		OutputTokens:        summary.OutputTokens,
		CacheReadTokens:     summary.CacheReadTokens,
		CacheCreationTokens: summary.CacheCreationTokens,
		CacheReadRatio:      computeInputCacheReadRatio(summary),
		TotalCacheTokens:    totalInputCacheTokens(summary),
		TotalTokens:         totalInputCacheVolume(summary),
		HasData:             hasInputCacheData(summary),
		Partial:             partial,
		AvailableFrom:       availableFrom,
	}
}

func BuildInputCacheMetricsWindow(summary *usagestats.InputCacheSummary, startTime, endTime time.Time, partial bool, availableFrom *time.Time) *InputCacheMetricsWindow {
	if summary == nil {
		summary = &usagestats.InputCacheSummary{}
	}
	return &InputCacheMetricsWindow{
		InputTokens:         summary.InputTokens,
		OutputTokens:        summary.OutputTokens,
		CacheReadTokens:     summary.CacheReadTokens,
		CacheCreationTokens: summary.CacheCreationTokens,
		CacheReadRatio:      computeInputCacheReadRatio(summary),
		TotalCacheTokens:    totalInputCacheTokens(summary),
		TotalTokens:         totalInputCacheVolume(summary),
		HasData:             hasInputCacheData(summary),
		Partial:             partial,
		AvailableFrom:       availableFrom,
		StartTime:           startTime,
		EndTime:             endTime,
	}
}

func SumTrendInputCacheSummary(trend []usagestats.TrendDataPoint) *usagestats.InputCacheSummary {
	out := &usagestats.InputCacheSummary{}
	for _, point := range trend {
		out.InputTokens += point.InputTokens
		out.OutputTokens += point.OutputTokens
		out.CacheReadTokens += point.CacheReadTokens
		out.CacheCreationTokens += point.CacheCreationTokens
	}
	return out
}

func InputCacheSummaryFromDashboardStats(stats *usagestats.DashboardStats) *usagestats.InputCacheSummary {
	if stats == nil {
		return &usagestats.InputCacheSummary{}
	}
	return &usagestats.InputCacheSummary{
		InputTokens:         stats.TotalInputTokens,
		OutputTokens:        stats.TotalOutputTokens,
		CacheReadTokens:     stats.TotalCacheReadTokens,
		CacheCreationTokens: stats.TotalCacheCreationTokens,
	}
}

func (s *DashboardService) GetInputCacheSummary(ctx context.Context, filters usagestats.UsageLogFilters) (*usagestats.InputCacheSummary, error) {
	if s == nil || s.usageRepo == nil {
		return nil, fmt.Errorf("usage repository not available")
	}
	return s.usageRepo.GetInputCacheSummary(ctx, filters)
}

func (s *DashboardService) GetCumulativeInputCacheSummary(ctx context.Context, filters usagestats.UsageLogFilters) (*usagestats.InputCacheSummary, error) {
	cumulativeFilters := filters
	cumulativeFilters.StartTime = nil
	cumulativeFilters.EndTime = nil

	if !HasDashboardInputCacheScopedFilters(cumulativeFilters) {
		stats, err := s.GetDashboardStats(ctx)
		if err != nil {
			return nil, err
		}
		return InputCacheSummaryFromDashboardStats(stats), nil
	}

	availableFrom := s.InputCacheAvailableFrom(time.Now().UTC())
	cumulativeFilters.StartTime = &availableFrom
	return s.GetInputCacheSummary(ctx, cumulativeFilters)
}

func (s *DashboardService) InputCacheAvailableFrom(now time.Time) time.Time {
	days := 90
	if s != nil && s.aggUsageDays > 0 {
		days = s.aggUsageDays
	}
	return truncateToDayUTC(now.UTC().AddDate(0, 0, -days))
}

func (s *OpsService) GetInputCacheSummary(ctx context.Context, filter *OpsDashboardFilter) (*usagestats.InputCacheSummary, error) {
	if s == nil {
		return nil, fmt.Errorf("ops service not available")
	}
	if err := s.RequireMonitoringEnabled(ctx); err != nil {
		return nil, err
	}
	if s.opsRepo == nil {
		return nil, fmt.Errorf("ops repository not available")
	}
	return s.opsRepo.GetInputCacheSummary(ctx, filter)
}

func (s *OpsService) InputCacheAvailableFrom(now time.Time) time.Time {
	days := 90
	if s != nil && s.cfg != nil && s.cfg.DashboardAgg.Retention.UsageLogsDays > 0 {
		days = s.cfg.DashboardAgg.Retention.UsageLogsDays
	}
	return truncateToDayUTC(now.UTC().AddDate(0, 0, -days))
}

func HasDashboardInputCacheScopedFilters(filters usagestats.UsageLogFilters) bool {
	if filters.UserID > 0 || filters.APIKeyID > 0 || filters.AccountID > 0 || filters.GroupID > 0 {
		return true
	}
	if strings.TrimSpace(filters.Model) != "" {
		return true
	}
	return filters.RequestType != nil || filters.Stream != nil || filters.BillingType != nil
}

func HasOpsInputCacheScopedFilters(filter *OpsDashboardFilter) bool {
	if filter == nil {
		return false
	}
	if filter.GroupID != nil && *filter.GroupID > 0 {
		return true
	}
	return strings.TrimSpace(filter.Platform) != ""
}

func computeInputCacheReadRatio(summary *usagestats.InputCacheSummary) *float64 {
	if summary == nil {
		return nil
	}
	denominator := totalInputCacheVolume(summary)
	if denominator <= 0 {
		return nil
	}
	value := (float64(totalInputCacheTokens(summary)) / float64(denominator)) * 100
	value = math.Round(value*100) / 100
	return &value
}

func hasInputCacheData(summary *usagestats.InputCacheSummary) bool {
	if summary == nil {
		return false
	}
	return totalInputCacheVolume(summary) > 0
}

func totalInputCacheTokens(summary *usagestats.InputCacheSummary) int64 {
	if summary == nil {
		return 0
	}
	return summary.CacheCreationTokens + summary.CacheReadTokens
}

func totalInputCacheVolume(summary *usagestats.InputCacheSummary) int64 {
	if summary == nil {
		return 0
	}
	return summary.InputTokens + summary.OutputTokens + totalInputCacheTokens(summary)
}
