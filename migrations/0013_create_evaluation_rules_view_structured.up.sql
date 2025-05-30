CREATE MATERIALIZED VIEW IF NOT EXISTS evaluation_rules_final AS
  SELECT
    (data->>'id')::uuid AS id,
    (data->>'version')::integer AS version,
    (data->>'device_id')::uuid AS device_id,
    data->>'description' AS description,
    data->>'kind' AS kind,
    (data ->> 'enabled') = 'true' AS enabled,
    data->>'parameters' AS parameters,
    try_parse_monotonic_iso8601_timestamp(data->>'created_at') AS created_at,
    try_parse_monotonic_iso8601_timestamp(data->>'updated_at') AS updated_at
  FROM evaluation_rules;