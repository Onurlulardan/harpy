SELECT writer, lifecycle_state, generation
FROM notifications_writer_authority
WHERE singleton
FOR SHARE;
