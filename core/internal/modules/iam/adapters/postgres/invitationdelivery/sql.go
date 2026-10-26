package invitationdelivery

import _ "embed"

//go:embed sql/tenant/read_receipt.sql
var readReceiptSQL string

//go:embed sql/tenant/claim_artifact.sql
var claimArtifactSQL string

//go:embed sql/tenant/read_claim.sql
var readClaimSQL string

//go:embed sql/tenant/release_claim.sql
var releaseClaimSQL string

//go:embed sql/tenant/mark_unknown.sql
var markUnknownSQL string

//go:embed sql/tenant/finalize.sql
var finalizeSQL string
