CREATE MATERIALIZED VIEW IF NOT EXISTS devices_final AS
  SELECT
    (data->>'id')::uuid AS id,
    (data->>'version')::integer AS version,
    data->>'name' AS name,
    data->>'app_eui' AS app_eui,
    data->>'dev_eui' AS dev_eui,
    data->>'app_key' AS app_key,
    try_parse_monotonic_iso8601_timestamp(data->>'created_at') AS created_at,
    try_parse_monotonic_iso8601_timestamp(data->>'updated_at') AS updated_at
  FROM devices;