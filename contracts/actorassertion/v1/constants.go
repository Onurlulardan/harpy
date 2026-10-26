package actorassertionv1

const (
	ContractVersion = "iam-v1-actor-assertion.v1"
	Algorithm       = "Ed25519"
	Issuer          = "urn:harpy:core-api"
	HeaderName      = "Harpy-Actor-Assertion"
	LifetimeSeconds = int64(30)
	MaximumWireSize = 4096
	wirePrefix      = "fc_actor_v1"
)

func Audience(domain Domain) string {
	if domain.Validate() != nil {
		return ""
	}
	return "urn:harpy:v1-domain:" + string(domain)
}
