SELECT tenant_id,owner_user_id,anchor_id,new_database_system_id,security_epoch,tenant_slug
FROM platform.iam_post_restore_binding_v1($1,$2)
