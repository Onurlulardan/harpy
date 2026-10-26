package licenseauthority

import "errors"

var (
	ErrInvalidDocument       = errors.New("license snapshot document is invalid")
	ErrUntrustedKey          = errors.New("license snapshot signing key is untrusted")
	ErrInvalidSignature      = errors.New("license snapshot signature is invalid")
	ErrWrongBinding          = errors.New("license snapshot binding does not match the environment")
	ErrNotYetValid           = errors.New("license snapshot is not yet valid")
	ErrSequenceRollback      = errors.New("license snapshot sequence rollback")
	ErrSequenceConflict      = errors.New("license snapshot sequence conflicts with the accepted digest")
	ErrAuthorityState        = errors.New("persisted license authority state is invalid")
	ErrIssueRequest          = errors.New("license snapshot issue request is invalid")
	ErrInvalidTrustBundle    = errors.New("license trust bundle is invalid")
	ErrUntrustedTrustRoot    = errors.New("license trust root is untrusted")
	ErrInvalidTrustSignature = errors.New("license trust keyset signature is invalid")
	ErrRefreshRetryable      = errors.New("license refresh is retryable")
	ErrRefreshFenced         = errors.New("license refresh is fenced")
)
