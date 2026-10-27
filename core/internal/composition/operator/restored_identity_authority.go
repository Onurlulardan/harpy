package operator

import (
	"time"

	iammodule "github.com/Onurlulardan/harpy/core/internal/modules/iam"
	platformauthoritymodule "github.com/Onurlulardan/harpy/core/internal/modules/platformauthority"
)

func verifyProtectedAuthority(
	security iammodule.RestoredSecurityEvidence,
	trustDocument []byte,
	environmentID iammodule.UUID,
	now time.Time,
) error {
	return platformauthoritymodule.VerifyRestoredAuthority(
		platformauthoritymodule.RestoredAuthorityEvidence{
			EnvironmentID:             environmentID.String(),
			ProtectedHeadSHA256:       security.ProtectedHeadSHA256,
			ProtectedAckSHA256:        security.ProtectedAckSHA256,
			ProtectedCheckpointSHA256: security.ProtectedCheckpointSHA256,
			AckTrustSHA256:            security.AckTrustSHA256,
			ProtectedAck:              security.ProtectedAck,
			ProtectedCheckpoint:       security.ProtectedCheckpoint,
		},
		trustDocument,
		now,
	)
}
