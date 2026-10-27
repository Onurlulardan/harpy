package api

import "errors"

var (
	ErrConfiguration      = errors.New("artifact object configuration is invalid")
	ErrWorkloadIdentity   = errors.New("artifact object requires workload identity credentials")
	ErrVersioningRequired = errors.New("artifact object bucket requires enabled versioning")
	ErrPrivateACLRequired = errors.New("artifact object bucket requires a private ACL")
	ErrObjectInput        = errors.New("artifact object input is invalid")
	ErrProviderEvidence   = errors.New("artifact object provider evidence is incomplete")
	ErrUnavailable        = errors.New("artifact object provider is unavailable")
	ErrOutcomeUnknown     = errors.New("artifact object effect outcome is unknown")
	ErrIntegrity          = errors.New("artifact object integrity verification failed")
)
