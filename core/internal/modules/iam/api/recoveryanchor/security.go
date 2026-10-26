package recoveryanchor

import "context"

type RegistrationExpectation struct {
	Challenge      string
	RPID           string
	AllowedOrigins []string
}

type RegistrationEvidence struct {
	CredentialID, PublicKeyCOSE        []byte
	AAGUID                             string
	AttestationFormat, AttestationType string
	SignCount                          int64
}

type RegistrationVerifier interface {
	Verify(context.Context, []byte, RegistrationExpectation) (RegistrationEvidence, error)
}
