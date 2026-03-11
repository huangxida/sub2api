package service

import (
	"context"
	"encoding/json"
	"strings"
	"time"
)

type UsageLogDetail struct {
	UsageLogID int64

	RequestBody        *string
	RequestContentType *string
	RequestBytes       int64
	RequestIsJSON      bool
	RequestComplete    bool

	ResponseBody        *string
	ResponseFrames      []string
	ResponseContentType *string
	ResponseBytes       int64
	ResponseIsJSON      bool
	ResponseComplete    bool

	CreatedAt time.Time
}

type UsageLogDetailSaveInput struct {
	UsageLogID int64

	RequestBody        []byte
	RequestBytes       int64
	RequestContentType string
	RequestComplete    bool

	ResponseBody        []byte
	ResponseFrames      [][]byte
	ResponseBytes       int64
	ResponseContentType string
	ResponseComplete    bool
}

type UsageLogDetailRepository interface {
	Upsert(ctx context.Context, detail *UsageLogDetail) error
	GetByUsageLogID(ctx context.Context, usageLogID int64) (*UsageLogDetail, error)
	BatchHasDetails(ctx context.Context, usageLogIDs []int64) (map[int64]bool, error)
}

type UsageLogDetailService struct {
	repo UsageLogDetailRepository
}

func NewUsageLogDetailService(repo UsageLogDetailRepository) *UsageLogDetailService {
	return &UsageLogDetailService{repo: repo}
}

func (s *UsageLogDetailService) Save(ctx context.Context, input UsageLogDetailSaveInput) error {
	if s == nil || s.repo == nil || input.UsageLogID <= 0 {
		return nil
	}

	detail := &UsageLogDetail{
		UsageLogID:          input.UsageLogID,
		RequestBody:         rawBytesToOptionalString(input.RequestBody),
		RequestBytes:        normalizeObservedBytes(input.RequestBytes, input.RequestBody, nil),
		RequestIsJSON:       input.RequestComplete && len(input.RequestBody) > 0 && json.Valid(input.RequestBody),
		RequestComplete:     input.RequestComplete,
		ResponseBody:        rawBytesToOptionalString(input.ResponseBody),
		ResponseFrames:      rawFramesToStrings(input.ResponseFrames),
		ResponseBytes:       normalizeObservedBytes(input.ResponseBytes, input.ResponseBody, input.ResponseFrames),
		ResponseIsJSON:      input.ResponseComplete && len(input.ResponseFrames) == 0 && len(input.ResponseBody) > 0 && json.Valid(input.ResponseBody),
		ResponseComplete:    input.ResponseComplete,
		RequestContentType:  normalizeOptionalString(input.RequestContentType),
		ResponseContentType: normalizeOptionalString(input.ResponseContentType),
	}

	return s.repo.Upsert(ctx, detail)
}

func (s *UsageLogDetailService) GetByUsageLogID(ctx context.Context, usageLogID int64) (*UsageLogDetail, error) {
	if s == nil || s.repo == nil || usageLogID <= 0 {
		return nil, nil
	}
	return s.repo.GetByUsageLogID(ctx, usageLogID)
}

func (s *UsageLogDetailService) BatchHasDetails(ctx context.Context, usageLogIDs []int64) (map[int64]bool, error) {
	if s == nil || s.repo == nil || len(usageLogIDs) == 0 {
		return map[int64]bool{}, nil
	}
	return s.repo.BatchHasDetails(ctx, usageLogIDs)
}

func rawBytesToOptionalString(raw []byte) *string {
	if raw == nil {
		return nil
	}
	value := string(raw)
	return &value
}

func rawFramesToStrings(frames [][]byte) []string {
	if len(frames) == 0 {
		return nil
	}
	out := make([]string, 0, len(frames))
	for _, frame := range frames {
		out = append(out, string(frame))
	}
	return out
}

func rawPayloadBytes(body []byte, frames [][]byte) int64 {
	if len(frames) > 0 {
		var total int64
		for _, frame := range frames {
			total += int64(len(frame))
		}
		return total
	}
	return int64(len(body))
}

func normalizeObservedBytes(observed int64, body []byte, frames [][]byte) int64 {
	if observed > 0 {
		return observed
	}
	return rawPayloadBytes(body, frames)
}

func normalizeOptionalString(value string) *string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}
