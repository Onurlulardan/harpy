package platformauthority

import (
	"crypto/ed25519"
	"net/http"

	postgresadapter "github.com/Onurlulardan/harpy/core/internal/modules/platformauthority/adapters/postgres"
	"github.com/Onurlulardan/harpy/core/internal/modules/platformauthority/ports"
)

type SaaSPlannerDependencies struct {
	Pool               *postgresadapter.Pool
	Clock              ports.Clock
	KMSKeyID           string
	KMSPublicKeySHA256 string
	KMSClient          KMSClient
}

type SaaSAcknowledgerDependencies struct {
	Pool             *postgresadapter.Pool
	Clock            ports.Clock
	HTTPClient       *http.Client
	RegistryURL      string
	AckTrustBundle   []byte
	ReleaseTrustRoot ed25519.PublicKey
}
