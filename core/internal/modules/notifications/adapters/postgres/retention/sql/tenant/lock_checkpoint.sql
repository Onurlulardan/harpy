SELECT revision
FROM notifications_retention_checkpoints
WHERE singleton
FOR UPDATE;
