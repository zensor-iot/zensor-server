CREATE MATERIALIZED VIEW IF NOT EXISTS tenants_final AS
  SELECT
    (data->>'id')::uuid AS id,
    (data->>'version')::integer AS version,
    data->>'name' AS name,
    data->>'email' AS email,
    data->>'description' AS description,
    (data->>'is_active')::boolean AS is_active,
    try_parse_monotonic_iso8601_timestamp(data->>'created_at') AS created_at,
    try_parse_monotonic_iso8601_timestamp(data->>'updated_at') AS updated_at,
    CASE 
      WHEN data->>'deleted_at' IS NULL THEN NULL
      ELSE try_parse_monotonic_iso8601_timestamp(data->>'deleted_at')
    END AS deleted_at
  FROM tenants; 