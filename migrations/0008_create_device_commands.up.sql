CREATE SOURCE IF NOT EXISTS device_commands
FROM
    KAFKA CONNECTION kafka_connection (TOPIC 'device_commands') FORMAT JSON;
