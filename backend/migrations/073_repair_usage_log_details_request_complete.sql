CREATE TABLE IF NOT EXISTS usage_log_details (
    usage_log_id BIGINT PRIMARY KEY REFERENCES usage_logs(id) ON DELETE CASCADE,
    request_body TEXT,
    request_content_type TEXT,
    request_bytes BIGINT NOT NULL DEFAULT 0,
    request_is_json BOOLEAN NOT NULL DEFAULT FALSE,
    request_complete BOOLEAN NOT NULL DEFAULT TRUE,
    response_body TEXT,
    response_frames JSONB,
    response_content_type TEXT,
    response_bytes BIGINT NOT NULL DEFAULT 0,
    response_is_json BOOLEAN NOT NULL DEFAULT FALSE,
    response_complete BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

ALTER TABLE usage_log_details
    ADD COLUMN IF NOT EXISTS request_complete BOOLEAN NOT NULL DEFAULT TRUE;
