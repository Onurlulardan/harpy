package artifact

const (
	StateSchema       = "protected-authority-state.v1"
	ProposalSchema    = "protected-authority-proposal.v1"
	AckSchema         = "protected-authority-ack.v1"
	BindingSchema     = "protected-authority-binding-proof.v1"
	CheckpointSchema  = "protected-authority-checkpoint.v1"
	TrustBundleSchema = "protected-authority-ack-trust-bundle.v1"
	AlgorithmEd25519  = "Ed25519"
	MaximumBytes      = 256 * 1024
)

const (
	domainPrefix      = "harpy.protected-authority."
	proposalDomain    = domainPrefix + "proposal.v1"
	ackDomain         = domainPrefix + "ack.v1"
	bindingDomain     = domainPrefix + "binding-proof.v1"
	checkpointDomain  = domainPrefix + "checkpoint.v1"
	trustBundleDomain = domainPrefix + "ack-trust-bundle.v1"
)

type DeploymentMode string

const (
	ModeOnPremise DeploymentMode = "onpremise"
	ModeSaaS      DeploymentMode = "saas"
)

type ScopeKind string

const (
	ScopeEnvironment ScopeKind = "environment"
	ScopeTenant      ScopeKind = "tenant"
)

type RegistryKind string

const (
	RegistryDevelopment RegistryKind = "development"
	RegistrySaaS        RegistryKind = "saas_provider"
)
