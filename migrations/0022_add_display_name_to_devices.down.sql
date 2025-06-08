-- Drop the materialized view with display_name
DROP MATERIALIZED VIEW IF EXISTS devices_final;

-- Recreate devices_final materialized view without display_name column (rollback)
CREATE MATERIALIZED VIEW devices_final AS
  SELECT
    (data->>'id')::uuid AS id,
    (data->>'version')::integer AS version,
    data->>'name' AS name,
    data->>'app_eui' AS app_eui,
    data->>'dev_eui' AS dev_eui,
    data->>'app_key' AS app_key,
    CASE 
      WHEN data->>'tenant_id' IS NULL THEN NULL
      ELSE (data->>'tenant_id')::uuid
    END AS tenant_id,
    try_parse_monotonic_iso8601_timestamp(data->>'created_at') AS created_at,
    try_parse_monotonic_iso8601_timestamp(data->>'updated_at') AS updated_at
  FROM devices; 