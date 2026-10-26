INSERT INTO work_resource_links (
    id, work_resource_kind, work_resource_id, target_owner, target_kind,
    target_id, target_version_id, usage, target_reference_epoch, permission_revision,
    state, row_version, created_at, released_at
) VALUES ($1, 'TASK', $2, $3, $4, $5, $6, 'READ', $7, $8,
          'ACTIVE', $9, transaction_timestamp(), NULL)
