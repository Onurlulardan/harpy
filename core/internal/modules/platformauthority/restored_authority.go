package platformauthority

import (
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformauthority/ports"
)

type RestoredAuthorityEvidence = ports.RestoredAuthorityEvidence

func VerifyRestoredAuthority(
	evidence RestoredAuthorityEvidence,
	trustDocument []byte,
	now time.Time,
) error {
	return ports.VerifyRestoredAuthority(evidence, trustDocument, now)
}
