-- Add deleted_at column to scheduled_tasks table for soft deletion
-- Drop and recreate the materialized view to include the new column
DROP MATERIALIZED VIEW IF EXISTS scheduled_tasks_final;
CREATE MATERIALIZED VIEW scheduled_tasks_final AS
SELECT (data->>'id')::uuid AS id,
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
  END AS last_executed_at,
  CASE
    WHEN data->>'deleted_at' IS NULL THEN NULL
    ELSE try_parse_monotonic_iso8601_timestamp(data->>'deleted_at')
  END AS deleted_at
FROM scheduled_tasks;
-- Create index on deleted_at for efficient filtering
CREATE INDEX idx_scheduled_tasks_deleted_at ON scheduled_tasks_final(deleted_at);