package fcartv1

import "errors"

var (
	ErrInvalidInput       = errors.New("FCART1 input is invalid")
	ErrInvalidHeader      = errors.New("FCART1 header is invalid")
	ErrInvalidFormat      = errors.New("FCART1 format is invalid")
	ErrInvalidDEK         = errors.New("FCART1 DEK is invalid")
	ErrInvalidMetadata    = errors.New("FCART1 metadata is invalid")
	ErrIdentityMismatch   = errors.New("FCART1 identity does not match metadata")
	ErrCiphertextLimit    = errors.New("FCART1 ciphertext limit exceeded")
	ErrRandomSource       = errors.New("FCART1 random source failed")
	ErrNonceReuse         = errors.New("FCART1 nonce was reused")
	ErrPlaintextRead      = errors.New("FCART1 plaintext read failed")
	ErrCiphertextRead     = errors.New("FCART1 ciphertext read failed")
	ErrShortWrite         = errors.New("FCART1 destination short write")
	ErrTruncated          = errors.New("FCART1 ciphertext is truncated")
	ErrTrailingData       = errors.New("FCART1 ciphertext has trailing data")
	ErrSequence           = errors.New("FCART1 chunk sequence is invalid")
	ErrChunkLength        = errors.New("FCART1 chunk length is invalid")
	ErrAuthentication     = errors.New("FCART1 chunk authentication failed")
	ErrChecksum           = errors.New("FCART1 ciphertext checksum mismatch")
	ErrInvalidKEK         = errors.New("FCART1 key-encryption key is invalid")
	ErrInvalidKeyEnvelope = errors.New("FCART1 key envelope is invalid")
	ErrKEKUnavailable     = errors.New("FCART1 key-encryption key is unavailable")
	ErrKeyAuthentication  = errors.New("FCART1 key envelope authentication failed")
)
