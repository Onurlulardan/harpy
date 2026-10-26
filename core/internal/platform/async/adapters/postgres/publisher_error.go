package postgres

import "errors"

var (
	ErrOutboxAuthorityPageLimit  = errors.New("outbox authority page limit must be between 1 and 256")
	ErrOutboxAuthorityToken      = errors.New("outbox authority token is invalid")
	ErrOutboxAuthorityInvalid    = errors.New("outbox authority is invalid")
	ErrOutboxAuthorityStale      = errors.New("outbox authority is no longer current")
	ErrOutboxCatalogIncompatible = errors.New("live outbox is incompatible with the producer catalog")
	ErrPublishClaimLimit         = errors.New("publish claim limit must be between 1 and 50")
	ErrPublishClaimInvalid       = errors.New("publish claim is invalid")
	ErrPublishLeaseLost          = errors.New("publish lease is no longer owned")
	ErrPublishConfirmDigest      = errors.New("publish confirmation digest is invalid")
	ErrPublishReleaseInvalid     = errors.New("publish release is invalid")
	ErrPublishRecoveryLimit      = errors.New("publish recovery limit must be between 1 and 50")
)
