ALTER TABLE usage_log_details
    ADD COLUMN IF NOT EXISTS request_headers JSONB;
