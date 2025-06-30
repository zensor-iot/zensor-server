CREATE MATERIALIZED VIEW IF NOT EXISTS scheduled_tasks_final AS
  SELECT *
  FROM scheduled_tasks; 