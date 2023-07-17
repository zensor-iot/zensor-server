CREATE MATERIALIZED VIEW IF NOT EXISTS devices AS
  SELECT CAST(data AS jsonb) AS data
  FROM (
      SELECT convert_from(data, 'utf8') AS data
      FROM device_registered
  );