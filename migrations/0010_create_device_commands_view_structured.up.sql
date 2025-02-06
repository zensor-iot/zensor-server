CREATE MATERIALIZED VIEW IF NOT EXISTS device_commands_final AS
  SELECT
    data->>'device' AS device,
    data->>'raw_payload' AS raw_payload,
    (data->>'port')::integer AS port,
    data->>'priority' AS priority,
    try_parse_monotonic_iso8601_timestamp(data->>'created_at') AS created_at
  FROM device_commands;
