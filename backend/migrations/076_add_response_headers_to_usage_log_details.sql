ALTER TABLE usage_log_details
    ADD COLUMN IF NOT EXISTS response_headers JSONB;
