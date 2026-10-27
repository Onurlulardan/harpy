SELECT state,receipt_sha256,reactivated_at
FROM platform.iam_post_restore_observe_v1($1,$2)
