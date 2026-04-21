ALTER TABLE usage_log_details
    ADD COLUMN IF NOT EXISTS final_request_body TEXT,
    ADD COLUMN IF NOT EXISTS final_request_content_type TEXT,
    ADD COLUMN IF NOT EXISTS final_request_bytes BIGINT NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS final_request_is_json BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN IF NOT EXISTS final_request_complete BOOLEAN NOT NULL DEFAULT TRUE;
