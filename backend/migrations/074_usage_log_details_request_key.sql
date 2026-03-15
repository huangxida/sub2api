ALTER TABLE usage_log_details
    ADD COLUMN IF NOT EXISTS request_id VARCHAR(64),
    ADD COLUMN IF NOT EXISTS api_key_id BIGINT;

UPDATE usage_log_details d
SET
    request_id = NULLIF(ul.request_id, ''),
    api_key_id = ul.api_key_id
FROM usage_logs ul
WHERE d.usage_log_id = ul.id
  AND (
      d.request_id IS DISTINCT FROM NULLIF(ul.request_id, '')
      OR d.api_key_id IS DISTINCT FROM ul.api_key_id
  );

UPDATE usage_log_details
SET request_id = NULL
WHERE request_id = '';

WITH ranked AS (
    SELECT
        ctid,
        ROW_NUMBER() OVER (
            PARTITION BY request_id, api_key_id
            ORDER BY (usage_log_id IS NOT NULL) DESC, usage_log_id DESC NULLS LAST, created_at DESC, ctid DESC
        ) AS rn
    FROM usage_log_details
    WHERE request_id IS NOT NULL
      AND api_key_id IS NOT NULL
)
DELETE FROM usage_log_details d
USING ranked r
WHERE d.ctid = r.ctid
  AND r.rn > 1;

DO $$
DECLARE
    pk_name TEXT;
BEGIN
    SELECT conname
    INTO pk_name
    FROM pg_constraint
    WHERE conrelid = 'usage_log_details'::regclass
      AND contype = 'p';

    IF pk_name IS NOT NULL THEN
        EXECUTE format('ALTER TABLE usage_log_details DROP CONSTRAINT %I', pk_name);
    END IF;
END $$;

ALTER TABLE usage_log_details
    ALTER COLUMN usage_log_id DROP NOT NULL;

CREATE UNIQUE INDEX IF NOT EXISTS idx_usage_log_details_usage_log_id_unique
    ON usage_log_details (usage_log_id)
    WHERE usage_log_id IS NOT NULL;

CREATE UNIQUE INDEX IF NOT EXISTS idx_usage_log_details_request_id_api_key_unique
    ON usage_log_details (request_id, api_key_id);

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conrelid = 'usage_logs'::regclass
          AND conname = 'usage_logs_request_id_api_key_key'
    ) THEN
        ALTER TABLE usage_logs
            ADD CONSTRAINT usage_logs_request_id_api_key_key
            UNIQUE USING INDEX idx_usage_logs_request_id_api_key_unique;
    END IF;
END $$;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conrelid = 'usage_log_details'::regclass
          AND conname = 'usage_log_details_request_id_api_key_fkey'
    ) THEN
        ALTER TABLE usage_log_details
            ADD CONSTRAINT usage_log_details_request_id_api_key_fkey
            FOREIGN KEY (request_id, api_key_id)
            REFERENCES usage_logs (request_id, api_key_id)
            ON DELETE CASCADE;
    END IF;
END $$;
