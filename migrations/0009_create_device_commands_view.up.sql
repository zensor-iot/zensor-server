CREATE MATERIALIZED VIEW IF NOT EXISTS device_commands_final AS
SELECT
    *
FROM
    device_commands;
