package lifecycledelivery

import _ "embed"

//go:embed sql/tenant/read_receipt.sql
var readReceiptSQL string

//go:embed sql/tenant/read_artifact.sql
var readArtifactSQL string

//go:embed sql/tenant/claim_artifact.sql
var claimArtifactSQL string

//go:embed sql/tenant/read_claim.sql
var readClaimSQL string

//go:embed sql/tenant/release_claim.sql
var releaseClaimSQL string

//go:embed sql/tenant/mark_unknown.sql
var markUnknownSQL string

//go:embed sql/tenant/insert_receipt.sql
var insertReceiptSQL string

//go:embed sql/tenant/delete_artifact.sql
var deleteArtifactSQL string

//go:embed sql/tenant/record_setup_outcome.sql
var recordSetupOutcomeSQL string
