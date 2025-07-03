-- Fix device_commands_final view to have proper structured columns
DROP MATERIALIZED VIEW IF EXISTS device_commands_final;
CREATE MATERIALIZED VIEW device_commands_final AS
SELECT (data->>'id')::uuid AS id,
    (data->>'version')::integer AS version,
    (data->>'device_id')::uuid AS device_id,
    (data->>'task_id')::uuid AS task_id,
    data->>'device_name' AS device_name,
    (data->>'port')::integer AS port,
    data->>'priority' AS priority,
    data->>'payload' AS payload,
    (data->>'sent') = 'true' AS sent,
    (data->>'ready') = 'true' AS ready,
    try_parse_monotonic_iso8601_timestamp(data->>'created_at') AS created_at,
    try_parse_monotonic_iso8601_timestamp(data->>'dispatch_after') AS dispatch_after,
    try_parse_monotonic_iso8601_timestamp(data->>'sent_at') AS sent_at
FROM device_commands;