SELECT license_id, licensee_organization_id, environment_id,
       delivery_kind, issuer, sequence, payload_sha256,
       authority_revision, database_now
FROM platform.billing_license_restore_binding_v1($1);
