SELECT principal_id, authorization_version, supported_locales,
       email_target, email_verified, telegram_target, telegram_verified
FROM iam_notifications_recipient_read_v1(
    (pg_catalog.current_schemas(false))[2]::name, $1::uuid
);
