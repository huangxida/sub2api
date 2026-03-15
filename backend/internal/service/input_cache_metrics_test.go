package service

import (
	"context"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/pkg/usagestats"
	"github.com/stretchr/testify/require"
)

type inputCacheUsageRepoStub struct {
	usageRepoStub
	inputCacheSummary *usagestats.InputCacheSummary
	inputCacheFilters usagestats.UsageLogFilters
	inputCacheCalls   int
}

func (s *inputCacheUsageRepoStub) GetInputCacheSummary(ctx context.Context, filters usagestats.UsageLogFilters) (*usagestats.InputCacheSummary, error) {
	s.inputCacheCalls++
	s.inputCacheFilters = filters
	if s.inputCacheSummary != nil {
		return s.inputCacheSummary, nil
	}
	return &usagestats.InputCacheSummary{}, nil
}

func TestBuildInputCacheMetricsSnapshot(t *testing.T) {
	availableFrom := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
	snapshot := BuildInputCacheMetricsSnapshot(&usagestats.InputCacheSummary{
		InputTokens:         80,
		CacheReadTokens:     20,
		CacheCreationTokens: 5,
	}, true, &availableFrom)

	require.NotNil(t, snapshot)
	require.True(t, snapshot.HasData)
	require.True(t, snapshot.Partial)
	require.Equal(t, int64(80), snapshot.InputTokens)
	require.Equal(t, int64(20), snapshot.CacheReadTokens)
	require.Equal(t, int64(5), snapshot.CacheCreationTokens)
	require.Equal(t, int64(25), snapshot.TotalCacheTokens)
	require.Equal(t, int64(105), snapshot.TotalTokens)
	require.NotNil(t, snapshot.CacheReadRatio)
	require.InDelta(t, 23.81, *snapshot.CacheReadRatio, 0.001)
	require.Equal(t, &availableFrom, snapshot.AvailableFrom)
}

func TestDashboardServiceGetCumulativeInputCacheSummaryUsesDashboardStatsWhenUnfiltered(t *testing.T) {
	repo := &inputCacheUsageRepoStub{
		usageRepoStub: usageRepoStub{
			stats: &usagestats.DashboardStats{
				TotalInputTokens:         120,
				TotalCacheReadTokens:     30,
				TotalCacheCreationTokens: 10,
			},
		},
	}

	svc := NewDashboardService(repo, nil, nil, &config.Config{})
	summary, err := svc.GetCumulativeInputCacheSummary(context.Background(), usagestats.UsageLogFilters{})
	require.NoError(t, err)
	require.Equal(t, int64(120), summary.InputTokens)
	require.Equal(t, int64(30), summary.CacheReadTokens)
	require.Equal(t, int64(10), summary.CacheCreationTokens)
	require.Zero(t, repo.inputCacheCalls)
}

func TestDashboardServiceGetCumulativeInputCacheSummaryScopesToRetentionWindow(t *testing.T) {
	repo := &inputCacheUsageRepoStub{
		usageRepoStub: usageRepoStub{
			stats: &usagestats.DashboardStats{},
		},
	}

	cfg := &config.Config{}
	cfg.DashboardAgg.Retention.UsageLogsDays = 30
	svc := NewDashboardService(repo, nil, nil, cfg)

	before := time.Now().UTC()
	summary, err := svc.GetCumulativeInputCacheSummary(context.Background(), usagestats.UsageLogFilters{
		Model: "gpt-5",
	})
	after := time.Now().UTC()

	require.NoError(t, err)
	require.NotNil(t, summary)
	require.Equal(t, 1, repo.inputCacheCalls)
	require.Equal(t, "gpt-5", repo.inputCacheFilters.Model)
	require.Nil(t, repo.inputCacheFilters.EndTime)
	require.NotNil(t, repo.inputCacheFilters.StartTime)

	expectedLower := truncateToDayUTC(before.AddDate(0, 0, -30))
	expectedUpper := truncateToDayUTC(after.AddDate(0, 0, -30))
	require.False(t, repo.inputCacheFilters.StartTime.Before(expectedLower))
	require.False(t, repo.inputCacheFilters.StartTime.After(expectedUpper))
}
