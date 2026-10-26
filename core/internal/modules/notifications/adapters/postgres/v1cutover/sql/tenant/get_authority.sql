SELECT writer, lifecycle_state, generation, row_version
FROM notifications_writer_authority
WHERE singleton;
