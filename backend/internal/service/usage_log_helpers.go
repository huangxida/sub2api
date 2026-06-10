package service

import "strings"

func optionalTrimmedStringPtr(raw string) *string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

// optionalNonEqualStringPtr returns a pointer to value if it is non-empty and
// differs from compare; otherwise nil. Used to store upstream_model only when
// it differs from the requested model.
func optionalNonEqualStringPtr(value, compare string) *string {
	value = strings.TrimSpace(value)
	compare = strings.TrimSpace(compare)
	if value == "" || value == compare {
		return nil
	}
	return &value
}

func buildUsageModelMappingChain(requestedModel, upstreamModel string) string {
	requestedModel = strings.TrimSpace(requestedModel)
	upstreamModel = strings.TrimSpace(upstreamModel)
	if requestedModel == "" || upstreamModel == "" || requestedModel == upstreamModel {
		return ""
	}
	return requestedModel + "→" + upstreamModel
}

func forwardResultBillingModel(requestedModel, upstreamModel string) string {
	if trimmed := strings.TrimSpace(requestedModel); trimmed != "" {
		return trimmed
	}
	return strings.TrimSpace(upstreamModel)
}

func optionalInt64Ptr(v int64) *int64 {
	if v == 0 {
		return nil
	}
	return &v
}
