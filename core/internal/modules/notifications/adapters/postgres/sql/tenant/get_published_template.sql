SELECT t.id::text, v.id::text, t.notification_type, v.template_kind,
       v.locale, v.subject_pattern, v.body_pattern, v.variables,
       v.mandatory_email, v.version_number
FROM notifications_templates t
JOIN notifications_template_versions v ON v.id = t.published_version_id
WHERE v.id = $1::uuid AND v.locale = ANY($2::text[]);
