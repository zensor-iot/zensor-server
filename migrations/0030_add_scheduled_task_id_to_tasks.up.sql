-- Drop the existing tasks views
DROP MATERIALIZED VIEW IF EXISTS tasks_final;

-- Recreate the tasks view with the new scheduled_task_id field
CREATE MATERIALIZED VIEW IF NOT EXISTS tasks_final AS
  SELECT *
  FROM tasks; 