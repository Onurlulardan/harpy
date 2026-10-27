package application

import (
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformauthority/domain/artifact"
)

type operationCommand struct {
	operationID    string
	purpose        string
	state          artifact.CanonicalState
	generation     int64
	previousHead   *string
	compensatesAck *string
	rotation       *artifact.SignerRotationEvidence
	issuedAt       time.Time
	policy         artifact.ExpiryPolicy
}

func expiryFor(policy artifact.ExpiryPolicy) time.Duration {
	if policy == artifact.ExpiryAirGap {
		return 7 * 24 * time.Hour
	}
	return 10 * time.Minute
}
