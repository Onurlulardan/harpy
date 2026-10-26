SELECT offering_version_id,revision,offering_sha256,effective_at,
       source_sequence,source_sha256
FROM platform.billing_installation_offering_read_v1($1,$2)
