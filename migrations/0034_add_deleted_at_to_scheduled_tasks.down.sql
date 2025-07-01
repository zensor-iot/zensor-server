-- Remove deleted_at column from scheduled_tasks table
ALTER TABLE scheduled_tasks_final 
DROP COLUMN IF EXISTS deleted_at;

-- Drop the index
DROP INDEX IF EXISTS idx_scheduled_tasks_deleted_at;

-- Update the materialized view to remove the deleted_at column
DROP MATERIALIZED VIEW IF EXISTS scheduled_tasks_final;

CREATE MATERIALIZED VIEW scheduled_tasks_final AS
  SELECT
    (data->>'id')::uuid AS id,
    (data->>'version')::integer AS version,
    data->>'tenant_id' AS tenant_id,
    data->>'device_id' AS device_id,
    data->>'command_templates' AS command_templates,
    data->>'schedule' AS schedule,
    (data->>'is_active')::boolean AS is_active,
    try_parse_monotonic_iso8601_timestamp(data->>'created_at') AS created_at,
    try_parse_monotonic_iso8601_timestamp(data->>'updated_at') AS updated_at,
    CASE 
      WHEN data->>'last_executed_at' IS NULL THEN NULL
      ELSE try_parse_monotonic_iso8601_timestamp(data->>'last_executed_at')
    END AS last_executed_at
  FROM scheduled_tasks; 