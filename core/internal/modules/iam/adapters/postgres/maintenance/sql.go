package maintenance

import _ "embed"

//go:embed sql/tenant/read_receipt.sql
var readReceiptSQL string

//go:embed sql/tenant/execute_bcrypt_retire.sql
var executeBcryptRetireSQL string

//go:embed sql/tenant/execute_identifier_retire.sql
var executeIdentifierRetireSQL string

//go:embed sql/tenant/execute_directory_tombstone_purge.sql
var executeDirectoryTombstonePurgeSQL string

//go:embed sql/tenant/read_directory_tombstone_candidates.sql
var readDirectoryTombstoneCandidatesSQL string

//go:embed sql/platform/purge_directory_projection.sql
var purgeDirectoryProjectionSQL string

//go:embed sql/tenant/execute_security_artifact_purge.sql
var executeSecurityArtifactPurgeSQL string

//go:embed sql/tenant/read_directory_page.sql
var readDirectoryPageSQL string

//go:embed sql/tenant/apply_directory_repair.sql
var applyDirectoryRepairSQL string

//go:embed sql/platform/apply_directory_projection.sql
var applyDirectoryProjectionSQL string

//go:embed sql/tenant/finish_directory_repair.sql
var finishDirectoryRepairSQL string

//go:embed sql/tenant/read_directory_sources.sql
var readDirectorySourcesSQL string

//go:embed sql/platform/read_directory_snapshot.sql
var readDirectorySnapshotSQL string

//go:embed sql/tenant/finish_directory_checksum.sql
var finishDirectoryChecksumSQL string

//go:embed sql/tenant/read_profile_page.sql
var readProfilePageSQL string

//go:embed sql/tenant/read_profile_resources.sql
var readProfileResourcesSQL string

//go:embed sql/tenant/read_profile_resource_evidence.sql
var readProfileResourceEvidenceSQL string

//go:embed sql/tenant/read_profile_mappings.sql
var readProfileMappingsSQL string

//go:embed sql/tenant/read_profile_schema.sql
var readProfileSchemaSQL string

//go:embed sql/tenant/apply_profile_backfill.sql
var applyProfileBackfillSQL string

//go:embed sql/tenant/record_profile_error.sql
var recordProfileErrorSQL string

//go:embed sql/tenant/finish_profile_backfill.sql
var finishProfileBackfillSQL string
