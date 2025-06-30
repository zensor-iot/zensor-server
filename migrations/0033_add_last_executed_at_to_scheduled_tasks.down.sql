-- Drop the existing materialized view to recreate it without the last_executed_at field
DROP MATERIALIZED VIEW IF EXISTS scheduled_tasks_final;

-- Recreate the materialized view without the last_executed_at field
CREATE MATERIALIZED VIEW IF NOT EXISTS scheduled_tasks_final AS
  SELECT
    (data->>'id')::uuid AS id,
    (data->>'version')::integer AS version,
    (data->>'tenant_id')::uuid AS tenant_id,
    (data->>'device_id')::uuid AS device_id,
    data->>'command_templates' AS command_templates,
    (data->>'schedule')::text AS schedule,
    (data->>'is_active')::boolean AS is_active,
    try_parse_monotonic_iso8601_timestamp(data->>'created_at') AS created_at,
    try_parse_monotonic_iso8601_timestamp(data->>'updated_at') AS updated_at
  FROM scheduled_tasks; 