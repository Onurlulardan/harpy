package licensev2

const (
	SnapshotEnvelopeSchema = "harpy-license-envelope.v2"
	SnapshotSchema         = "harpy-license-snapshot.v2"
	SnapshotAudience       = "harpy-billing"
	SnapshotProduct        = "harpy"
	TrustBundleSchema      = "harpy-license-trust-bundle.v1"
	KeysetEnvelopeSchema   = "harpy-license-keyset-envelope.v1"
	KeysetSchema           = "harpy-license-online-keyset.v1"
	KeysetAudience         = "harpy-license-snapshot"
	IssueRequestSchema     = "harpy-license-issue-request.v2"
	DerivedOfferingSchema  = "harpy-license-derived-offering.v1"
	SignatureAlgorithm     = "Ed25519"
	Canonicalization       = "RFC8785"
)

type DeliveryKind string

const (
	DeliveryConnected   DeliveryKind = "connected"
	DeliveryAirGap      DeliveryKind = "airgap"
	DeliveryDevelopment DeliveryKind = "development"
)

type Status string

const (
	StatusActive    Status = "active"
	StatusSuspended Status = "suspended"
	StatusRevoked   Status = "revoked"
	StatusExpired   Status = "expired"
)

type TrustKeyStatus string

const (
	TrustKeyActive     TrustKeyStatus = "active"
	TrustKeyVerifyOnly TrustKeyStatus = "verify_only"
	TrustKeyRevoked    TrustKeyStatus = "revoked"
)
