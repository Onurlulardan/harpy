package licenseauthority

import licensev2 "github.com/Onurlulardan/harpy/contracts/license/v2"

const (
	EnvelopeSchema       = licensev2.SnapshotEnvelopeSchema
	SnapshotSchema       = licensev2.SnapshotSchema
	SnapshotAudience     = licensev2.SnapshotAudience
	SnapshotProduct      = licensev2.SnapshotProduct
	SignatureAlgorithm   = licensev2.SignatureAlgorithm
	Canonicalization     = licensev2.Canonicalization
	MaximumDocumentBytes = 64 * 1024
	TrustBundleSchema    = licensev2.TrustBundleSchema
	KeysetEnvelopeSchema = licensev2.KeysetEnvelopeSchema
	KeysetSchema         = licensev2.KeysetSchema
	KeysetAudience       = licensev2.KeysetAudience
)

var signatureDomain = []byte(EnvelopeSchema + "\x00")

var keysetSignatureDomain = []byte(KeysetEnvelopeSchema + "\x00")
