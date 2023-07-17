CREATE MATERIALIZED VIEW IF NOT EXISTS events AS
  SELECT CAST(data AS jsonb) AS data
  FROM (
      SELECT convert_from(data, 'utf8') AS data
      FROM event_emitted
  );