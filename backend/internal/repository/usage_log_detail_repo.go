package repository

import (
	"context"
	"database/sql"
	"encoding/json"

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
	if r == nil || detail == nil || detail.UsageLogID <= 0 {
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
			request_body,
			request_content_type,
			request_bytes,
			request_is_json,
			request_complete,
			response_body,
			response_frames,
			response_content_type,
			response_bytes,
			response_is_json,
			response_complete
		) VALUES (
			$1, $2, $3, $4, $5, $6,
			$7, $8, $9, $10, $11, $12
		)
		ON CONFLICT (usage_log_id) DO UPDATE SET
			request_body = EXCLUDED.request_body,
			request_content_type = EXCLUDED.request_content_type,
			request_bytes = EXCLUDED.request_bytes,
			request_is_json = EXCLUDED.request_is_json,
			request_complete = EXCLUDED.request_complete,
			response_body = EXCLUDED.response_body,
			response_frames = EXCLUDED.response_frames,
			response_content_type = EXCLUDED.response_content_type,
			response_bytes = EXCLUDED.response_bytes,
			response_is_json = EXCLUDED.response_is_json,
			response_complete = EXCLUDED.response_complete
	`,
		detail.UsageLogID,
		nullOptionalText(detail.RequestBody),
		nullOptionalText(detail.RequestContentType),
		detail.RequestBytes,
		detail.RequestIsJSON,
		detail.RequestComplete,
		nullOptionalText(detail.ResponseBody),
		responseFrames,
		nullOptionalText(detail.ResponseContentType),
		detail.ResponseBytes,
		detail.ResponseIsJSON,
		detail.ResponseComplete,
	)
	return err
}

func (r *usageLogDetailRepository) GetByUsageLogID(ctx context.Context, usageLogID int64) (*service.UsageLogDetail, error) {
	if r == nil || usageLogID <= 0 {
		return nil, nil
	}

	query := `
		SELECT
			usage_log_id,
			request_body,
			request_content_type,
			request_bytes,
			request_is_json,
			request_complete,
			response_body,
			response_frames,
			response_content_type,
			response_bytes,
			response_is_json,
			response_complete,
			created_at
		FROM usage_log_details
		WHERE usage_log_id = $1
	`

	rows, err := r.sql.QueryContext(ctx, query, usageLogID)
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
		detail              service.UsageLogDetail
		requestBody         sql.NullString
		requestContentType  sql.NullString
		responseBody        sql.NullString
		responseFramesBytes []byte
		responseContentType sql.NullString
	)
	if err := rows.Scan(
		&detail.UsageLogID,
		&requestBody,
		&requestContentType,
		&detail.RequestBytes,
		&detail.RequestIsJSON,
		&detail.RequestComplete,
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

	detail.RequestBody = stringPtrFromNullString(requestBody)
	detail.RequestContentType = stringPtrFromNullString(requestContentType)
	detail.ResponseBody = stringPtrFromNullString(responseBody)
	detail.ResponseContentType = stringPtrFromNullString(responseContentType)
	if len(responseFramesBytes) > 0 {
		if err := json.Unmarshal(responseFramesBytes, &detail.ResponseFrames); err != nil {
			return nil, err
		}
	}
	return &detail, nil
}

func (r *usageLogDetailRepository) BatchHasDetails(ctx context.Context, usageLogIDs []int64) (map[int64]bool, error) {
	result := make(map[int64]bool, len(usageLogIDs))
	if r == nil || len(usageLogIDs) == 0 {
		return result, nil
	}

	rows, err := r.sql.QueryContext(ctx, `
		SELECT usage_log_id
		FROM usage_log_details
		WHERE usage_log_id = ANY($1)
	`, pq.Array(usageLogIDs))
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
