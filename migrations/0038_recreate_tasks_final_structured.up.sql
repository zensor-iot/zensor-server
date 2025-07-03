-- Drop the existing tasks_final view if it exists
DROP MATERIALIZED VIEW IF EXISTS tasks_final;
-- Recreate the tasks_final view with the requested structured columns
CREATE MATERIALIZED VIEW IF NOT EXISTS tasks_final AS
SELECT (data->>'id')::uuid AS id,
    (data->>'device_id')::uuid AS device_id,
    (data->>'scheduled_task_id')::uuid AS scheduled_task_id,
    (data->>'version')::integer AS version,
    try_parse_monotonic_iso8601_timestamp(data->>'created_at') AS created_at,
    try_parse_monotonic_iso8601_timestamp(data->>'updated_at') AS updated_at
FROM tasks;