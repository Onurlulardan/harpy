SELECT environment_id = $1
FROM platform.environment_identity
WHERE singleton;
