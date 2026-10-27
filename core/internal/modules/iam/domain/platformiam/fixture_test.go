package platformiam

import "time"

func fixtureID(value byte) ID {
	var id ID
	id[0], id[15] = value, value
	return id
}

func fixtureEvidence() Evidence {
	return Evidence{
		OperationID: fixtureID(1), RequestSHA256: [32]byte{1},
		ActorID: fixtureID(2), ExpectedTokenVersion: 3,
		ExpectedAuthorizationVersion: 4, ExpectedPlatformClock: 5,
	}
}

func fixtureStepUp(now time.Time, purpose [32]byte) StepUpEvidence {
	return StepUpEvidence{
		PurposeSHA256: purpose, AuthTime: now.Add(-time.Minute), ACR: StepUpACR,
		AMRWebAuthn: true,
	}
}
