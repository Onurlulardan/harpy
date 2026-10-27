INSERT INTO work_resource_links (
    id, work_resource_kind, work_resource_id, target_owner, target_kind,
    target_id, target_version_id, usage, target_reference_epoch,
    permission_revision, state, row_version, created_at, released_at
) VALUES ($1,$2,$3,'FORMS','FORM',$4,$5,'READ',$6,$7,'ACTIVE',1,
    transaction_timestamp(),NULL)
ON CONFLICT (id) DO UPDATE SET
    target_id=EXCLUDED.target_id, target_version_id=EXCLUDED.target_version_id,
    target_reference_epoch=EXCLUDED.target_reference_epoch,
    permission_revision=EXCLUDED.permission_revision, state='ACTIVE', released_at=NULL
