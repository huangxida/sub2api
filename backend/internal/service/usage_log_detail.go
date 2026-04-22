package service

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"
)

type UsageLogDetailKey struct {
	RequestID string
	APIKeyID  int64
}

type UsageLogDetail struct {
	UsageLogID int64
	RequestID  string
	APIKeyID   int64

	RequestHeaders     *string
	RequestBody        *string
	RequestContentType *string
	RequestBytes       int64
	RequestIsJSON      bool
	RequestComplete    bool

	FinalRequestBody        *string
	FinalRequestContentType *string
	FinalRequestBytes       int64
	FinalRequestIsJSON      bool
	FinalRequestComplete    bool

	ResponseHeaders     *string
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
	RequestID  string
	APIKeyID   int64

	RequestHeaders     http.Header
	RequestBody        []byte
	RequestBytes       int64
	RequestContentType string
	RequestComplete    bool

	FinalRequestBody        []byte
	FinalRequestBytes       int64
	FinalRequestContentType string
	FinalRequestComplete    bool
	FinalRequestTransformed bool

	ResponseHeaders     http.Header
	ResponseBody        []byte
	ResponseFrames      [][]byte
	ResponseBytes       int64
	ResponseContentType string
	ResponseComplete    bool
}

type UsageLogDetailRepository interface {
	Upsert(ctx context.Context, detail *UsageLogDetail) error
	GetByRequestKey(ctx context.Context, key UsageLogDetailKey) (*UsageLogDetail, error)
	BatchHasDetailsByRequestKeys(ctx context.Context, keys []UsageLogDetailKey) (map[UsageLogDetailKey]bool, error)
	GetByUsageLogID(ctx context.Context, usageLogID int64) (*UsageLogDetail, error)
	BatchHasDetailsByUsageLogIDs(ctx context.Context, usageLogIDs []int64) (map[int64]bool, error)
}

type UsageLogDetailService struct {
	repo UsageLogDetailRepository
}

func NewUsageLogDetailService(repo UsageLogDetailRepository) *UsageLogDetailService {
	return &UsageLogDetailService{repo: repo}
}

func (s *UsageLogDetailService) Save(ctx context.Context, input UsageLogDetailSaveInput) error {
	if s == nil || s.repo == nil {
		return nil
	}

	requestID := strings.TrimSpace(input.RequestID)
	if requestID == "" || input.APIKeyID <= 0 {
		return nil
	}

	detail := &UsageLogDetail{
		UsageLogID:              input.UsageLogID,
		RequestID:               requestID,
		APIKeyID:                input.APIKeyID,
		RequestHeaders:          httpHeadersToOptionalJSONText(input.RequestHeaders),
		RequestBody:             rawBytesToOptionalString(input.RequestBody),
		RequestBytes:            normalizeObservedBytes(input.RequestBytes, input.RequestBody, nil),
		RequestIsJSON:           input.RequestComplete && len(input.RequestBody) > 0 && json.Valid(input.RequestBody),
		RequestComplete:         input.RequestComplete,
		FinalRequestBody:        rawBytesToOptionalString(input.FinalRequestBody),
		FinalRequestBytes:       normalizeObservedBytes(input.FinalRequestBytes, input.FinalRequestBody, nil),
		FinalRequestIsJSON:      input.FinalRequestComplete && len(input.FinalRequestBody) > 0 && json.Valid(input.FinalRequestBody),
		FinalRequestComplete:    input.FinalRequestComplete,
		ResponseHeaders:         httpHeadersToOptionalJSONText(input.ResponseHeaders),
		ResponseBody:            rawBytesToOptionalString(input.ResponseBody),
		ResponseFrames:          rawFramesToStrings(input.ResponseFrames),
		ResponseBytes:           normalizeObservedBytes(input.ResponseBytes, input.ResponseBody, input.ResponseFrames),
		ResponseIsJSON:          input.ResponseComplete && len(input.ResponseFrames) == 0 && len(input.ResponseBody) > 0 && json.Valid(input.ResponseBody),
		ResponseComplete:        input.ResponseComplete,
		RequestContentType:      normalizeOptionalString(input.RequestContentType),
		FinalRequestContentType: normalizeOptionalString(input.FinalRequestContentType),
		ResponseContentType:     normalizeOptionalString(input.ResponseContentType),
	}

	return s.repo.Upsert(ctx, detail)
}

func (s *UsageLogDetailService) GetByUsageLog(ctx context.Context, usageLog *UsageLog) (*UsageLogDetail, error) {
	if s == nil || s.repo == nil || usageLog == nil {
		return nil, nil
	}

	key := usageLogDetailKeyFromUsageLog(usageLog)
	if key.RequestID != "" && key.APIKeyID > 0 {
		detail, err := s.repo.GetByRequestKey(ctx, key)
		if err != nil || detail != nil {
			if detail != nil && detail.UsageLogID == 0 {
				detail.UsageLogID = usageLog.ID
			}
			return detail, err
		}
	}
	if usageLog.ID <= 0 {
		return nil, nil
	}
	detail, err := s.repo.GetByUsageLogID(ctx, usageLog.ID)
	if detail != nil && detail.UsageLogID == 0 {
		detail.UsageLogID = usageLog.ID
	}
	return detail, err
}

func (s *UsageLogDetailService) BatchHasDetailsByUsageLogs(ctx context.Context, usageLogs []UsageLog) (map[int64]bool, error) {
	result := make(map[int64]bool, len(usageLogs))
	if s == nil || s.repo == nil || len(usageLogs) == 0 {
		return result, nil
	}

	keys := make([]UsageLogDetailKey, 0, len(usageLogs))
	logIDsByKey := make(map[UsageLogDetailKey][]int64, len(usageLogs))
	legacyLogIDs := make([]int64, 0, len(usageLogs))
	seen := make(map[UsageLogDetailKey]struct{}, len(usageLogs))
	for i := range usageLogs {
		if usageLogs[i].ID <= 0 {
			continue
		}
		key := usageLogDetailKeyFromUsageLog(&usageLogs[i])
		if key.RequestID == "" || key.APIKeyID <= 0 {
			legacyLogIDs = append(legacyLogIDs, usageLogs[i].ID)
			continue
		}
		logIDsByKey[key] = append(logIDsByKey[key], usageLogs[i].ID)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		keys = append(keys, key)
	}
	if len(keys) > 0 {
		found, err := s.repo.BatchHasDetailsByRequestKeys(ctx, keys)
		if err != nil {
			return nil, err
		}
		for key, logIDs := range logIDsByKey {
			if !found[key] {
				continue
			}
			for _, logID := range logIDs {
				result[logID] = true
			}
		}
	}
	if len(legacyLogIDs) > 0 {
		legacyFound, err := s.repo.BatchHasDetailsByUsageLogIDs(ctx, legacyLogIDs)
		if err != nil {
			return nil, err
		}
		for _, logID := range legacyLogIDs {
			if legacyFound[logID] {
				result[logID] = true
			}
		}
	}
	return result, nil
}

func usageLogDetailKeyFromUsageLog(usageLog *UsageLog) UsageLogDetailKey {
	if usageLog == nil {
		return UsageLogDetailKey{}
	}
	return UsageLogDetailKey{
		RequestID: strings.TrimSpace(usageLog.RequestID),
		APIKeyID:  usageLog.APIKeyID,
	}
}

func rawBytesToOptionalString(raw []byte) *string {
	if raw == nil {
		return nil
	}
	value := string(raw)
	return &value
}

func transformedRawBytesToOptionalString(raw []byte, transformed bool) *string {
	if !transformed || raw == nil {
		return nil
	}
	value := string(raw)
	return &value
}

func normalizeOptionalStringWhen(value string, condition bool) *string {
	if !condition {
		return nil
	}
	return normalizeOptionalString(value)
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

func httpHeadersToOptionalJSONText(headers http.Header) *string {
	if len(headers) == 0 {
		return nil
	}
	payload, err := json.Marshal(headers)
	if err != nil || len(payload) == 0 || string(payload) == "null" {
		return nil
	}
	value := string(payload)
	return &value
}
