SELECT operation_id,anchor_id,kind,authority_generation,replayed
FROM platform.iam_recovery_anchor_commit_v1($1::uuid,$2::uuid,$3::uuid,$4::timestamptz);
