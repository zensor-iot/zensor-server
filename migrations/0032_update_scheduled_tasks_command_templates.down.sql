DROP MATERIALIZED VIEW IF EXISTS scheduled_tasks_final;

CREATE MATERIALIZED VIEW IF NOT EXISTS scheduled_tasks_final AS
  SELECT
    (data->>'id')::uuid AS id,
    (data->>'version')::integer AS version,
    (data->>'tenant_id')::uuid AS tenant_id,
    (data->>'device_id')::uuid AS device_id,
    data->>'commands' AS commands,
    (data->>'schedule')::text AS schedule,
    (data->>'is_active')::boolean AS is_active,
    try_parse_monotonic_iso8601_timestamp(data->>'created_at') AS created_at,
    try_parse_monotonic_iso8601_timestamp(data->>'updated_at') AS updated_at
  FROM scheduled_tasks; 