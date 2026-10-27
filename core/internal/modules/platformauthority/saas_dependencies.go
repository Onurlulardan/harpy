package platformauthority

import (
	"crypto/ed25519"
	"net/http"

	postgresadapter "github.com/Onurlulardan/harpy/core/internal/modules/platformauthority/adapters/postgres"
	"github.com/Onurlulardan/harpy/core/internal/modules/platformauthority/ports"
)

type SaaSDependencies struct {
	Pool               *postgresadapter.Pool
	Clock              ports.Clock
	HTTPClient         *http.Client
	RegistryURL        string
	KMSKeyID           string
	KMSPublicKeySHA256 string
	KMSClient          KMSClient
	AckTrustBundle     []byte
	ReleaseTrustRoot   ed25519.PublicKey
}
