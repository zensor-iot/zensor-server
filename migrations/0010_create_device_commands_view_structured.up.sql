CREATE MATERIALIZED VIEW IF NOT EXISTS device_commands_final AS
  SELECT
    data->>'id' AS id,
    data->>'device_id' AS device_id,
    data->>'device_name' AS device_name,
    (data->>'port')::integer AS port,
    data->>'priority' AS priority,
    data->>'payload' AS payload,
    (data ->> 'sent') = 'true' AS sent,
    (data ->> 'ready') = 'true' AS ready,
    try_parse_monotonic_iso8601_timestamp(data->>'created_at') AS created_at,
    try_parse_monotonic_iso8601_timestamp(data->>'dispatch_after') AS dispatch_after,
    try_parse_monotonic_iso8601_timestamp(data->>'sent_at') AS sent_at
  FROM device_commands;
