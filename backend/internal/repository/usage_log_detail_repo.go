package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/lib/pq"
)

type usageLogDetailRepository struct {
	client *dbent.Client
	sql    sqlExecutor
}

func NewUsageLogDetailRepository(client *dbent.Client, sqlDB *sql.DB) service.UsageLogDetailRepository {
	return &usageLogDetailRepository{client: client, sql: sqlDB}
}

func (r *usageLogDetailRepository) Upsert(ctx context.Context, detail *service.UsageLogDetail) error {
	if r == nil || detail == nil {
		return nil
	}

	requestID := strings.TrimSpace(detail.RequestID)
	if requestID == "" || detail.APIKeyID <= 0 {
		return nil
	}

	sqlq := r.sql
	if tx := dbent.TxFromContext(ctx); tx != nil {
		sqlq = tx.Client()
	}

	var responseFrames any
	if len(detail.ResponseFrames) > 0 {
		payload, err := json.Marshal(detail.ResponseFrames)
		if err != nil {
			return err
		}
		responseFrames = string(payload)
	}

	_, err := sqlq.ExecContext(ctx, `
		INSERT INTO usage_log_details (
			usage_log_id,
			request_id,
			api_key_id,
			request_headers,
			request_body,
			request_content_type,
			request_bytes,
			request_is_json,
			request_complete,
			final_request_body,
			final_request_content_type,
			final_request_bytes,
			final_request_is_json,
			final_request_complete,
			response_headers,
			response_body,
			response_frames,
			response_content_type,
			response_bytes,
			response_is_json,
			response_complete
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8,
			$9, $10, $11, $12, $13, $14, $15, $16,
			$17, $18, $19, $20, $21
		)
		ON CONFLICT (request_id, api_key_id) DO UPDATE SET
			usage_log_id = COALESCE(EXCLUDED.usage_log_id, usage_log_details.usage_log_id),
			request_headers = EXCLUDED.request_headers,
			request_body = EXCLUDED.request_body,
			request_content_type = EXCLUDED.request_content_type,
			request_bytes = EXCLUDED.request_bytes,
			request_is_json = EXCLUDED.request_is_json,
			request_complete = EXCLUDED.request_complete,
			final_request_body = EXCLUDED.final_request_body,
			final_request_content_type = EXCLUDED.final_request_content_type,
			final_request_bytes = EXCLUDED.final_request_bytes,
			final_request_is_json = EXCLUDED.final_request_is_json,
			final_request_complete = EXCLUDED.final_request_complete,
			response_headers = EXCLUDED.response_headers,
			response_body = EXCLUDED.response_body,
			response_frames = EXCLUDED.response_frames,
			response_content_type = EXCLUDED.response_content_type,
			response_bytes = EXCLUDED.response_bytes,
			response_is_json = EXCLUDED.response_is_json,
			response_complete = EXCLUDED.response_complete
	`,
		nullOptionalInt64(detail.UsageLogID),
		requestID,
		detail.APIKeyID,
		nullOptionalText(detail.RequestHeaders),
		nullOptionalText(detail.RequestBody),
		nullOptionalText(detail.RequestContentType),
		detail.RequestBytes,
		detail.RequestIsJSON,
		detail.RequestComplete,
		nullOptionalText(detail.FinalRequestBody),
		nullOptionalText(detail.FinalRequestContentType),
		detail.FinalRequestBytes,
		detail.FinalRequestIsJSON,
		detail.FinalRequestComplete,
		nullOptionalText(detail.ResponseHeaders),
		nullOptionalText(detail.ResponseBody),
		responseFrames,
		nullOptionalText(detail.ResponseContentType),
		detail.ResponseBytes,
		detail.ResponseIsJSON,
		detail.ResponseComplete,
	)
	return err
}

func (r *usageLogDetailRepository) GetByRequestKey(ctx context.Context, key service.UsageLogDetailKey) (*service.UsageLogDetail, error) {
	if r == nil {
		return nil, nil
	}

	requestID := strings.TrimSpace(key.RequestID)
	if requestID == "" || key.APIKeyID <= 0 {
		return nil, nil
	}

	return r.getDetail(ctx, `
		SELECT
			COALESCE(usage_log_id, 0),
			COALESCE(request_id, ''),
			COALESCE(api_key_id, 0),
			request_headers,
			request_body,
			request_content_type,
			request_bytes,
			request_is_json,
			request_complete,
			final_request_body,
			final_request_content_type,
			final_request_bytes,
			final_request_is_json,
			final_request_complete,
			response_headers,
			response_body,
			response_frames,
			response_content_type,
			response_bytes,
			response_is_json,
			response_complete,
			created_at
		FROM usage_log_details
		WHERE request_id = $1 AND api_key_id = $2
	`, requestID, key.APIKeyID)
}

func (r *usageLogDetailRepository) GetByUsageLogID(ctx context.Context, usageLogID int64) (*service.UsageLogDetail, error) {
	if r == nil || usageLogID <= 0 {
		return nil, nil
	}
	return r.getDetail(ctx, `
		SELECT
			COALESCE(usage_log_id, 0),
			COALESCE(request_id, ''),
			COALESCE(api_key_id, 0),
			request_headers,
			request_body,
			request_content_type,
			request_bytes,
			request_is_json,
			request_complete,
			final_request_body,
			final_request_content_type,
			final_request_bytes,
			final_request_is_json,
			final_request_complete,
			response_headers,
			response_body,
			response_frames,
			response_content_type,
			response_bytes,
			response_is_json,
			response_complete,
			created_at
		FROM usage_log_details
		WHERE usage_log_id = $1
	`, usageLogID)
}

func (r *usageLogDetailRepository) BatchHasDetailsByRequestKeys(ctx context.Context, keys []service.UsageLogDetailKey) (map[service.UsageLogDetailKey]bool, error) {
	result := make(map[service.UsageLogDetailKey]bool, len(keys))
	if r == nil || len(keys) == 0 {
		return result, nil
	}

	args := make([]any, 0, len(keys)*2)
	values := make([]string, 0, len(keys))
	seen := make(map[service.UsageLogDetailKey]struct{}, len(keys))
	for _, key := range keys {
		normalized := service.UsageLogDetailKey{
			RequestID: strings.TrimSpace(key.RequestID),
			APIKeyID:  key.APIKeyID,
		}
		if normalized.RequestID == "" || normalized.APIKeyID <= 0 {
			continue
		}
		if _, ok := seen[normalized]; ok {
			continue
		}
		seen[normalized] = struct{}{}
		values = append(values, fmt.Sprintf("($%d::text, $%d::bigint)", len(args)+1, len(args)+2))
		args = append(args, normalized.RequestID, normalized.APIKeyID)
	}
	if len(values) == 0 {
		return result, nil
	}

	query := fmt.Sprintf(`
		WITH input(request_id, api_key_id) AS (
			VALUES %s
		)
		SELECT d.request_id, d.api_key_id
		FROM usage_log_details d
		INNER JOIN input i
			ON i.request_id = d.request_id
			AND i.api_key_id = d.api_key_id
	`, strings.Join(values, ", "))

	rows, err := r.sql.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	for rows.Next() {
		var key service.UsageLogDetailKey
		if err := rows.Scan(&key.RequestID, &key.APIKeyID); err != nil {
			return nil, err
		}
		result[key] = true
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

func (r *usageLogDetailRepository) BatchHasDetailsByUsageLogIDs(ctx context.Context, usageLogIDs []int64) (map[int64]bool, error) {
	result := make(map[int64]bool, len(usageLogIDs))
	if r == nil || len(usageLogIDs) == 0 {
		return result, nil
	}

	normalized := make([]int64, 0, len(usageLogIDs))
	seen := make(map[int64]struct{}, len(usageLogIDs))
	for _, usageLogID := range usageLogIDs {
		if usageLogID <= 0 {
			continue
		}
		if _, ok := seen[usageLogID]; ok {
			continue
		}
		seen[usageLogID] = struct{}{}
		normalized = append(normalized, usageLogID)
	}
	if len(normalized) == 0 {
		return result, nil
	}

	query := `
		SELECT usage_log_id
		FROM usage_log_details
		WHERE usage_log_id = ANY($1)
	`
	rows, err := r.sql.QueryContext(ctx, query, pq.Array(normalized))
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	for rows.Next() {
		var usageLogID int64
		if err := rows.Scan(&usageLogID); err != nil {
			return nil, err
		}
		result[usageLogID] = true
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

func (r *usageLogDetailRepository) getDetail(ctx context.Context, query string, args ...any) (*service.UsageLogDetail, error) {
	rows, err := r.sql.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return nil, err
		}
		return nil, nil
	}

	var (
		detail                  service.UsageLogDetail
		requestHeadersBytes     []byte
		requestBody             sql.NullString
		requestContentType      sql.NullString
		finalRequestBody        sql.NullString
		finalRequestContentType sql.NullString
		responseHeadersBytes    []byte
		responseBody            sql.NullString
		responseFramesBytes     []byte
		responseContentType     sql.NullString
	)
	if err := rows.Scan(
		&detail.UsageLogID,
		&detail.RequestID,
		&detail.APIKeyID,
		&requestHeadersBytes,
		&requestBody,
		&requestContentType,
		&detail.RequestBytes,
		&detail.RequestIsJSON,
		&detail.RequestComplete,
		&finalRequestBody,
		&finalRequestContentType,
		&detail.FinalRequestBytes,
		&detail.FinalRequestIsJSON,
		&detail.FinalRequestComplete,
		&responseHeadersBytes,
		&responseBody,
		&responseFramesBytes,
		&responseContentType,
		&detail.ResponseBytes,
		&detail.ResponseIsJSON,
		&detail.ResponseComplete,
		&detail.CreatedAt,
	); err != nil {
		return nil, err
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}

	detail.RequestHeaders = jsonTextPtrFromBytes(requestHeadersBytes)
	detail.RequestBody = stringPtrFromNullString(requestBody)
	detail.RequestContentType = stringPtrFromNullString(requestContentType)
	detail.FinalRequestBody = stringPtrFromNullString(finalRequestBody)
	detail.FinalRequestContentType = stringPtrFromNullString(finalRequestContentType)
	detail.ResponseHeaders = jsonTextPtrFromBytes(responseHeadersBytes)
	detail.ResponseBody = stringPtrFromNullString(responseBody)
	detail.ResponseContentType = stringPtrFromNullString(responseContentType)
	if len(responseFramesBytes) > 0 {
		if err := json.Unmarshal(responseFramesBytes, &detail.ResponseFrames); err != nil {
			return nil, err
		}
	}
	return &detail, nil
}

func nullOptionalText(value *string) sql.NullString {
	if value == nil {
		return sql.NullString{}
	}
	return sql.NullString{String: *value, Valid: true}
}

func stringPtrFromNullString(value sql.NullString) *string {
	if !value.Valid {
		return nil
	}
	out := value.String
	return &out
}

func jsonTextPtrFromBytes(value []byte) *string {
	trimmed := strings.TrimSpace(string(value))
	if trimmed == "" || trimmed == "null" {
		return nil
	}
	return &trimmed
}

func nullOptionalInt64(value int64) sql.NullInt64 {
	if value <= 0 {
		return sql.NullInt64{}
	}
	return sql.NullInt64{Int64: value, Valid: true}
}
